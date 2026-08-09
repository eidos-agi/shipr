package shipr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDetectReleaseModel_AbsorbsTestrCommands(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/p\n\ngo 1.22\n")
	writeFile(t, filepath.Join(dir, "README.md"), "# p\n")
	// testr model with distinctive commands
	tm := map[string]any{
		"schema_version": 1,
		"product_id":     "p",
		"test_commands":  []string{"make prove", "go test ./..."},
	}
	b, _ := json.MarshalIndent(tm, "", "  ")
	writeFile(t, filepath.Join(dir, TestrModelRel), string(b)+"\n")

	m := DetectReleaseModel(dir, "test product")
	if m["role"] != "ai_config_and_memory" {
		t.Fatalf("role: %v", m["role"])
	}
	if m["proof_source"] != "testr" {
		t.Fatalf("proof_source want testr got %v", m["proof_source"])
	}
	proofs, ok := m["proof_commands"].([]string)
	if !ok {
		// may be []string from uniqueKeepOrder
		raw, _ := json.Marshal(m["proof_commands"])
		var ss []string
		_ = json.Unmarshal(raw, &ss)
		proofs = ss
	}
	if len(proofs) == 0 || proofs[0] != "make prove" {
		t.Fatalf("expected testr commands first, got %v", m["proof_commands"])
	}
	rel, ok := m["related_testr"].(map[string]any)
	if !ok || rel["loaded"] != true {
		t.Fatalf("related_testr: %v", m["related_testr"])
	}
}

func TestDetectReleaseModel_NoTestr(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/p\n\ngo 1.22\n")
	m := DetectReleaseModel(dir, "")
	if m["proof_source"] != "detected" {
		t.Fatalf("proof_source: %v", m["proof_source"])
	}
}

func TestWriteAndLoadAndAttempt(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/p\n\ngo 1.22\n")
	writeFile(t, filepath.Join(dir, ".gitignore"), "")
	m := DetectReleaseModel(dir, "x")
	path, err := WriteReleaseModel(dir, m)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadReleaseModel(dir)
	if err != nil {
		t.Fatal(err)
	}
	if loaded["product_id"] != filepath.Base(dir) {
		t.Fatalf("product_id: %v", loaded["product_id"])
	}
	ap, attempt, err := RecordAttempt(dir, "first ship", "planned", "note", []string{"go test ./..."}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(ap); err != nil {
		t.Fatal(err)
	}
	if attempt["status"] != "planned" {
		t.Fatalf("attempt: %v", attempt)
	}
	fr := ReleaseFrontier(dir)
	if fr["status"] != "model_ready" {
		t.Fatalf("frontier: %v", fr)
	}
	if fr["role"] != "ai_config_and_memory" {
		t.Fatalf("frontier role: %v", fr["role"])
	}
}

func TestWriteDoesNotGitignoreAndCreatesSibling(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/p\n\ngo 1.22\n")
	writeFile(t, filepath.Join(dir, ".gitignore"), "node_modules/\n.shipr/\n.testr/\n*.log\n")
	m := DetectReleaseModel(dir, "")
	// force no testr absorb path
	path, err := WriteReleaseModel(dir, m)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	// sibling testr created
	if _, err := os.Stat(filepath.Join(dir, TestrModelRel)); err != nil {
		t.Fatalf("expected sibling testr model: %v", err)
	}
	gi, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	text := string(gi)
	if strings.Contains(text, ".shipr/") || strings.Contains(text, ".testr/") {
		t.Fatalf("gitignore still ignores configs:\n%s", text)
	}
	if !strings.Contains(text, "node_modules/") {
		t.Fatalf("lost other ignore lines: %s", text)
	}
}

func TestEnsureProductConfigsCreatesBoth(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/p\n\ngo 1.22\n")
	writeFile(t, filepath.Join(dir, ".gitignore"), ".shipr/\n")
	if err := EnsureProductConfigs(dir); err != nil {
		t.Fatal(err)
	}
	if !exists(dir, ModelRelPath) || !exists(dir, TestrModelRel) {
		t.Fatal("expected both models")
	}
}

func TestFrontierUsesLatestAttemptByTime(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/p\n\ngo 1.22\n")
	m := DetectReleaseModel(dir, "")
	if _, err := WriteReleaseModel(dir, m); err != nil {
		t.Fatal(err)
	}
	for _, st := range []string{"planned", "blocked", "ready"} {
		if _, _, err := RecordAttempt(dir, "stress-"+st, st, "", []string{"go test ./..."}, nil); err != nil {
			t.Fatal(err)
		}
	}
	fr := ReleaseFrontier(dir)
	if fr["latest_status"] != "ready" {
		t.Fatalf("want latest_status=ready got %v (latest=%v)", fr["latest_status"], fr["latest_attempt"])
	}
}
