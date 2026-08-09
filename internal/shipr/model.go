package shipr

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	ModelRelPath   = ".shipr/product-release-model.json"
	AttemptsRelDir = ".shipr/release-attempts"
	TestrModelRel  = ".testr/product-test-model.json"
	// Version is the Go CLI / config schema generation version.
	Version = "0.3.1"
)

// ReleaseModel is the durable per-product ship config AI agents read.
// Shipr does not ship or deploy; it stores how-to-ship + attempt memory.
type ReleaseModel map[string]any

func exists(root string, parts ...string) bool {
	p := filepath.Join(append([]string{root}, parts...)...)
	_, err := os.Stat(p)
	return err == nil
}

// configIgnoreForms must never appear in product .gitignore.
// .shipr/ and .testr/ are committed product config, not local noise.
var configIgnoreForms = map[string]struct{}{
	".shipr": {}, ".shipr/": {}, ".shipr/*": {}, ".shipr/**": {},
	".testr": {}, ".testr/": {}, ".testr/*": {}, ".testr/**": {},
}

// ensureNotGitignored strips .shipr/.testr ignore rules so configs can be tracked.
func ensureNotGitignored(root string) {
	gi := filepath.Join(root, ".gitignore")
	b, err := os.ReadFile(gi)
	if err != nil {
		return
	}
	lines := strings.Split(string(b), "\n")
	var keep []string
	changed := false
	for _, line := range lines {
		s := strings.TrimSpace(line)
		if _, drop := configIgnoreForms[s]; drop {
			changed = true
			continue
		}
		keep = append(keep, line)
	}
	if !changed {
		return
	}
	for len(keep) > 0 && keep[len(keep)-1] == "" {
		keep = keep[:len(keep)-1]
	}
	out := strings.Join(keep, "\n")
	if out != "" {
		out += "\n"
	}
	_ = os.WriteFile(gi, []byte(out), 0o644)
}

// bootstrapTestrModel writes a minimal testr config when shipr creates a product
// that has no .testr model yet (sibling ensure).
func bootstrapTestrModel(root string, ship ReleaseModel) map[string]any {
	product, _ := ship["product_id"].(string)
	if product == "" {
		product = filepath.Base(root)
	}
	cmds := stringSlice(ship["proof_commands"])
	if len(cmds) == 0 {
		if exists(root, "go.mod") {
			cmds = append(cmds, "go test ./...", "go build ./...")
		}
		if exists(root, "pyproject.toml") || exists(root, "tests") {
			cmds = append(cmds, "python -m pytest -q")
		}
		if len(cmds) == 0 {
			cmds = []string{"define product-specific test command before claiming green"}
		}
	}
	return map[string]any{
		"schema_version":    1,
		"role":              "ai_config_and_memory",
		"purpose":           "Tell AI agents how this product is proven. Store repeatable test config and attempt ledgers. Does not execute tests.",
		"product_id":        product,
		"project_root":      root,
		"description":       "bootstrapped by shipr (sibling ensure)",
		"test_suites":       []string{},
		"test_commands":     uniqueKeepOrder(cmds),
		"evidence_paths":    []string{"test output / junit / coverage when configured"},
		"methods_source":    "bootstrapped-by-shipr",
		"related_operators": []string{"shipr"},
		"related_shipr": map[string]any{
			"operator":   "shipr",
			"model_path": ModelRelPath,
			"loaded":     true,
			"note":       "shipr should use these test_commands as proof_commands when shipping",
		},
		"learning_questions": []string{
			"What failed that the ship path assumed green?",
			"Which test should become a shipr proof_command?",
			"What is flaky vs broken?",
		},
		"memory_paths": map[string]string{
			"model":        TestrModelRel,
			"attempts_dir": ".testr/test-attempts",
		},
		"updated_at": time.Now().UTC().Format(time.RFC3339Nano),
	}
}

// EnsureProductConfigs strips ignore rules and creates missing .shipr / .testr models.
func EnsureProductConfigs(project string) error {
	root, _ := filepath.Abs(project)
	ensureNotGitignored(root)
	if !exists(root, ModelRelPath) {
		m := DetectReleaseModel(root, "")
		if err := writeJSON(filepath.Join(root, ModelRelPath), m); err != nil {
			return err
		}
	}
	if !exists(root, TestrModelRel) {
		ship, err := LoadReleaseModel(root)
		if err != nil {
			ship = DetectReleaseModel(root, "")
		}
		if err := writeJSON(filepath.Join(root, TestrModelRel), bootstrapTestrModel(root, ship)); err != nil {
			return err
		}
	}
	return nil
}

func declaredLicense(root string) any {
	py := filepath.Join(root, "pyproject.toml")
	if b, err := os.ReadFile(py); err == nil {
		re := regexp.MustCompile(`(?m)^\s*license\s*=\s*\{?\s*text\s*=\s*"([^"]+)"`)
		if m := re.FindSubmatch(b); m != nil {
			return string(m[1])
		}
		re2 := regexp.MustCompile(`(?m)^\s*license\s*=\s*"([^"]+)"`)
		if m := re2.FindSubmatch(b); m != nil {
			return string(m[1])
		}
	}
	for _, n := range []string{"LICENSE", "LICENSE.txt", "LICENSE.md"} {
		if exists(root, n) {
			return "declared"
		}
	}
	return nil
}

func repositoryVisibility(root string) string {
	if !exists(root, ".git") {
		return "unknown"
	}
	if _, err := exec.LookPath("gh"); err != nil {
		return "unknown"
	}
	out, err := exec.Command("git", "-C", root, "remote", "get-url", "origin").Output()
	if err != nil {
		return "unknown"
	}
	remote := strings.TrimSpace(string(out))
	re := regexp.MustCompile(`github\.com[/:]([^/]+)/([^/]+?)(?:\.git)?$`)
	m := re.FindStringSubmatch(remote)
	if m == nil {
		return "unknown"
	}
	gh, err := exec.Command("gh", "repo", "view", m[1]+"/"+m[2], "--json", "visibility").Output()
	if err != nil {
		return "unknown"
	}
	var payload struct {
		Visibility string `json:"visibility"`
	}
	if json.Unmarshal(gh, &payload) != nil {
		return "unknown"
	}
	v := strings.ToLower(payload.Visibility)
	switch v {
	case "public", "private", "internal":
		return v
	default:
		return "unknown"
	}
}

func unique(ss []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range ss {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func uniqueKeepOrder(ss []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range ss {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func stringSlice(v any) []string {
	switch t := v.(type) {
	case []string:
		return t
	case []any:
		var out []string
		for _, x := range t {
			if s, ok := x.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

// loadTestrProofs reads sibling testr model and returns its test_commands.
// Shipr prefers these as proof_commands so ship and test config stay aligned.
func loadTestrProofs(root string) (commands []string, modelPath string, ok bool) {
	path := filepath.Join(root, TestrModelRel)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, "", false
	}
	var tm map[string]any
	if json.Unmarshal(b, &tm) != nil {
		return nil, path, false
	}
	cmds := stringSlice(tm["test_commands"])
	if len(cmds) == 0 {
		return nil, path, false
	}
	return cmds, path, true
}

// DetectReleaseModel builds (does not write) the product release config.
// When .testr/product-test-model.json exists, its test_commands become the
// preferred proof_commands so AI ships using the same gates testr records.
func DetectReleaseModel(project, description string) ReleaseModel {
	root, _ := filepath.Abs(project)
	product := filepath.Base(root)
	var artifactTypes, channels, proofs, rollback []string
	gates := []string{"credentials", "payments", "production mutations", "public publish/tag", "customer/outbound messaging"}
	license := declaredLicense(root)
	vis := repositoryVisibility(root)
	oss := "unknown"
	switch vis {
	case "public":
		if license != nil {
			oss = "ready"
		} else {
			oss = "license-missing"
		}
	case "private", "internal":
		oss = "private"
	default:
		if license != nil {
			oss = "candidate"
		}
	}
	// forge_stack names methods companions; "testr" is the sibling config tool
	companions := []string{"forge-forge", "security-forge"}
	if oss == "ready" || oss == "license-missing" || oss == "candidate" {
		companions = append(companions, "foss-forge")
	}
	companions = append(companions, "learning-forge", "loss-forge", "testr")

	if exists(root, "pyproject.toml") {
		artifactTypes = append(artifactTypes, "python-package")
		channels = append(channels, "PyPI or uvx")
		proofs = append(proofs, "python -m pytest -q", "python -m ruff check .", "python -m ruff format --check .")
		rollback = append(rollback, "bump patch version and release a fixed package; yank only for severe package faults")
	}
	if exists(root, "go.mod") {
		artifactTypes = append(artifactTypes, "go-module")
		channels = append(channels, "go install / binary")
		// Go-first proofs when both py + go present: put go test first via uniqueKeepOrder later with testr
		proofs = append(proofs, "go test ./...", "go build ./...")
		rollback = append(rollback, "revert tag and republish previous binary")
	}
	if exists(root, "package.json") {
		artifactTypes = append(artifactTypes, "web-or-node-app")
		channels = append(channels, "npm/web deploy")
		proofs = append(proofs, "npm test", "npm run build")
		rollback = append(rollback, "redeploy previous build or revert deployment")
	}
	if exists(root, "Dockerfile") || exists(root, "railway.json") {
		artifactTypes = append(artifactTypes, "service")
		channels = append(channels, "service deploy")
		proofs = append(proofs, "docker build .", "curl -fsS <health-url>")
		rollback = append(rollback, "redeploy previous image or rollback provider deployment")
	}
	if exists(root, "docs", "emf") {
		proofs = append(proofs, "python3 -m emf.validate docs/emf/")
	}
	if exists(root, "README.md") || exists(root, "docs") {
		artifactTypes = append(artifactTypes, "docs")
		proofs = append(proofs, "verify README, changelog, and release notes match artifact version")
	}

	proofSource := "detected"
	var testrPath any
	if tc, path, ok := loadTestrProofs(root); ok {
		// testr wins: put its commands first; keep extra detected proofs after
		proofs = uniqueKeepOrder(append(tc, proofs...))
		proofSource = "testr"
		testrPath = path
	}
	if len(proofs) == 0 {
		proofs = append(proofs, "define product-specific proof command before shipping")
	}
	if len(artifactTypes) == 0 {
		artifactTypes = append(artifactTypes, "unknown")
		channels = append(channels, "undiscovered")
	}

	return ReleaseModel{
		"schema_version": 1,
		"role":           "ai_config_and_memory",
		"purpose":        "Tell AI agents how this product ships. Store repeatable release config and attempt ledgers. Does not ship, deploy, or run proofs.",
		"product_id":     product,
		"project_root":   root,
		"description":    description,
		"repository_visibility": vis,
		"license":               license,
		"open_source_status":    oss,
		"artifact_types":        unique(artifactTypes),
		"distribution_channels": unique(channels),
		"proof_commands":        uniqueKeepOrder(proofs),
		"proof_source":          proofSource,
		"related_testr": map[string]any{
			"model_path": TestrModelRel,
			"loaded":     testrPath != nil,
			"abs_path":   testrPath,
			"note":       "When present, testr test_commands become shipr proof_commands",
		},
		"approval_gates": uniqueKeepOrder(gates),
		"rollback_paths": uniqueKeepOrder(rollback),
		"forge_stack":    uniqueKeepOrder(companions),
		"learning_questions": []string{
			"What broke or slowed this release?",
			"What proof was missing until late?",
			"Which gate should become automatic next time?",
			"Which human approval should remain explicit?",
		},
		"memory_paths": map[string]string{
			"model":        ModelRelPath,
			"attempts_dir": AttemptsRelDir,
		},
		"updated_at": time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

func WriteReleaseModel(project string, model ReleaseModel) (string, error) {
	root, _ := filepath.Abs(project)
	ensureNotGitignored(root)
	path := filepath.Join(root, ModelRelPath)
	if err := writeJSON(path, model); err != nil {
		return path, err
	}
	// Sibling: create .testr model if missing (do not overwrite).
	if !exists(root, TestrModelRel) {
		_ = writeJSON(filepath.Join(root, TestrModelRel), bootstrapTestrModel(root, model))
	}
	return path, nil
}

func LoadReleaseModel(project string) (ReleaseModel, error) {
	root, _ := filepath.Abs(project)
	path := filepath.Join(root, ModelRelPath)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m ReleaseModel
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func slug(text string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	s := strings.Trim(re.ReplaceAllString(strings.ToLower(text), "-"), "-")
	if s == "" {
		return "release"
	}
	if len(s) > 80 {
		s = s[:80]
	}
	return s
}

func nextActions(status string, blockers []string) []string {
	switch status {
	case "blocked":
		return []string{"clear the blocker, then re-run the product's proof_commands (AI runs them; shipr only records)"}
	case "ready":
		return []string{
			"request explicit human approval for public publish/deploy if needed",
			"record shipped or rolled_back after the irreversible step",
		}
	case "shipped":
		return []string{"route lessons to learning-forge", "watch rollback and support signals"}
	case "rolled_back":
		return []string{"record root cause", "define the next automatic release gate"}
	default:
		return []string{"read proof_commands from the model; AI runs them; record attempt status"}
	}
}

// RecordAttempt appends a release-attempt ledger entry. It does not execute proofs.
func RecordAttempt(project, goal, status, notes string, proofs, blockers []string) (string, map[string]any, error) {
	root, _ := filepath.Abs(project)
	_ = EnsureProductConfigs(root)
	model, err := LoadReleaseModel(root)
	if err != nil {
		model = DetectReleaseModel(root, "")
		_, _ = WriteReleaseModel(root, model)
	}
	ts := time.Now().UTC().Format("20060102T150405Z")
	path := filepath.Join(root, AttemptsRelDir, ts+"-"+slug(goal)+".json")
	if proofs == nil {
		proofs = []string{}
	}
	if blockers == nil {
		blockers = []string{}
	}
	attempt := map[string]any{
		"schema_version": 1,
		"product_id":     model["product_id"],
		"goal":           goal,
		"status":         status,
		"notes":          notes,
		"proofs":         proofs,
		"blockers":       blockers,
		"blocker_records": []any{},
		"gate_summary":   []any{},
		"source":         nil,
		"next_actions":   nextActions(status, blockers),
		"release_model_snapshot": map[string]any{
			"artifact_types":        model["artifact_types"],
			"distribution_channels": model["distribution_channels"],
			"proof_commands":        model["proof_commands"],
			"approval_gates":        model["approval_gates"],
			"forge_stack":           model["forge_stack"],
			"repository_visibility": model["repository_visibility"],
			"license":               model["license"],
			"open_source_status":    model["open_source_status"],
			"proof_source":          model["proof_source"],
		},
		"learning_prompts": model["learning_questions"],
		"created_at":       time.Now().UTC().Format(time.RFC3339Nano),
	}
	return path, attempt, writeJSON(path, attempt)
}

func ReleaseFrontier(project string) map[string]any {
	root, _ := filepath.Abs(project)
	model, err := LoadReleaseModel(root)
	attemptsDir := filepath.Join(root, AttemptsRelDir)
	var attemptFiles []string
	if entries, err := os.ReadDir(attemptsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
				attemptFiles = append(attemptFiles, filepath.Join(attemptsDir, e.Name()))
			}
		}
		sort.Strings(attemptFiles)
	}
	if err != nil {
		return map[string]any{
			"status":        "needs_release_model",
			"role":          "ai_config_and_memory",
			"next_actions":  []string{"run `shipr model --write` to materialize ship config for AI"},
			"attempt_count": len(attemptFiles),
			"related_testr": map[string]any{
				"model_path": filepath.Join(root, TestrModelRel),
				"exists":     exists(root, TestrModelRel),
			},
		}
	}
	var latest map[string]any
	if len(attemptFiles) > 0 {
		b, _ := os.ReadFile(attemptFiles[len(attemptFiles)-1])
		_ = json.Unmarshal(b, &latest)
	}
	next := []string{
		"AI: run proof_commands from the model (shipr does not execute them)",
		"record release attempt with `shipr attempt`",
		"route lessons to learning-forge after release",
	}
	if len(attemptFiles) == 0 {
		next = append([]string{"record the first release attempt"}, next...)
	} else if latest != nil {
		st, _ := latest["status"].(string)
		var blockers []string
		if raw, ok := latest["blockers"].([]any); ok {
			for _, b := range raw {
				blockers = append(blockers, str(b))
			}
		}
		next = nextActions(st, blockers)
	}
	out := map[string]any{
		"status":                 "model_ready",
		"role":                   "ai_config_and_memory",
		"product_id":             model["product_id"],
		"artifact_types":         model["artifact_types"],
		"distribution_channels":  model["distribution_channels"],
		"proof_commands":         model["proof_commands"],
		"proof_source":           model["proof_source"],
		"approval_gates":         model["approval_gates"],
		"attempt_count":          len(attemptFiles),
		"latest_attempt":         nil,
		"latest_status":          nil,
		"latest_blockers":        []any{},
		"latest_blocker_records": []any{},
		"recurring_blockers":     []any{},
		"next_actions":           next,
		"related_testr": map[string]any{
			"model_path": filepath.Join(root, TestrModelRel),
			"exists":     exists(root, TestrModelRel),
			"note":       "Align ship proofs with testr test_commands",
		},
	}
	if len(attemptFiles) > 0 {
		out["latest_attempt"] = attemptFiles[len(attemptFiles)-1]
		if latest != nil {
			out["latest_status"] = latest["status"]
			if b, ok := latest["blockers"]; ok {
				out["latest_blockers"] = b
			}
		}
	}
	return out
}

func str(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	b, _ := json.Marshal(v)
	return string(b)
}
