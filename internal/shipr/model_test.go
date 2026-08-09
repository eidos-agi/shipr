package shipr

import (
	"encoding/json"
	"os"
	"path/filepath"
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
