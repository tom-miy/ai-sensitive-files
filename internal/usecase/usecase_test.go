package usecase

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tom-miy/ai-sensitive-files/internal/domain"
)

func TestBuildGeneratedFilesIncludesGuidanceSnippet(t *testing.T) {
	files := BuildGeneratedFiles(testPolicy())
	var found bool
	for _, file := range files {
		if file.Path == "generated/ai-agent-guidance.md" {
			found = true
			if !bytes.Contains(file.Content, []byte(".env")) {
				t.Fatalf("guidance content does not include sensitive path: %s", file.Content)
			}
		}
	}
	if !found {
		t.Fatal("generated/ai-agent-guidance.md was not generated")
	}
}

func TestCheckFailsWhenGitTrackedCannotBeRead(t *testing.T) {
	dir := t.TempDir()
	writeGeneratedFixture(t, dir, testPolicy())
	result := Check(testPolicy(), dir)
	if !containsError(result.Errors, "git tracked file check failed") {
		t.Fatalf("Check errors = %#v, want git tracked failure", result.Errors)
	}
}

func TestCheckFailsWhenDecryptedFileExistsWithoutEncryptedFile(t *testing.T) {
	dir := t.TempDir()
	initGit(t, dir)
	policy := testPolicy()
	writeGeneratedFixture(t, dir, policy)
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("SECRET=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result := Check(policy, dir)
	if !containsError(result.Errors, "encrypted path is missing") {
		t.Fatalf("Check errors = %#v, want missing encrypted path error", result.Errors)
	}
}

func TestCheckWarnsWhenEncryptedFileExistsWithoutDecryptedFile(t *testing.T) {
	dir := t.TempDir()
	initGit(t, dir)
	policy := testPolicy()
	writeGeneratedFixture(t, dir, policy)
	if err := os.WriteFile(filepath.Join(dir, ".env.sops.yaml"), []byte("encrypted\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result := Check(policy, dir)
	if len(result.Errors) != 0 {
		t.Fatalf("Check errors = %#v, want none", result.Errors)
	}
	if !containsString(result.Warnings, "decrypted path is missing") {
		t.Fatalf("Check warnings = %#v, want missing decrypted warning", result.Warnings)
	}
}

func testPolicy() domain.Policy {
	return domain.Policy{SensitiveFiles: []domain.SensitiveFile{{
		Path:          ".env",
		EncryptedPath: ".env.sops.yaml",
		DecryptedPath: ".env",
		Reason:        "local env",
		HasAction:     true,
		HasCrypto:     true,
		Action: domain.Action{
			AIIgnore:    true,
			GitIgnore:   true,
			Encrypt:     true,
			CommitBlock: true,
		},
		Crypto: domain.Crypto{
			Method:         "sops-age",
			Recipients:     []string{"age1example"},
			EncryptCommand: "sops --encrypt --output {encrypted_path} {decrypted_path}",
			DecryptCommand: "sops --decrypt --output {decrypted_path} {encrypted_path}",
			ManualEdit:     "decrypted",
		},
	}}}
}

func writeGeneratedFixture(t *testing.T, dir string, policy domain.Policy) {
	t.Helper()
	if err := WriteGeneratedFiles(dir, BuildGeneratedFiles(policy), true); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(".env\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func initGit(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init", "-q")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
}

func containsError(errs []error, want string) bool {
	for _, err := range errs {
		if strings.Contains(err.Error(), want) {
			return true
		}
	}
	return false
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if strings.Contains(value, want) {
			return true
		}
	}
	return false
}
