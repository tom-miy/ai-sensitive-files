package infra

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPolicySupportsBlockSequences(t *testing.T) {
	dir := t.TempDir()
	policyPath := filepath.Join(dir, "policy.yaml")
	content := `sensitive_files:
  - path: ".env"
    encrypted_path: ".env.sops.yaml"
    decrypted_path: ".env"
    reason: "local env"
    tags:
      - env
      - secret
    crypto:
      method: "sops-age"
      recipients:
        - age1team
        - age1ci
      encrypt_command: "sops --encrypt --output {encrypted_path} {decrypted_path}"
      decrypt_command: "sops --decrypt --output {decrypted_path} {encrypted_path}"
      manual_edit: "decrypted"
    action:
      aiignore: true
      gitignore: true
      encrypt: true
      commit_block: true
`
	if err := os.WriteFile(policyPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	policy, err := LoadPolicy(policyPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(policy.SensitiveFiles); got != 1 {
		t.Fatalf("len(SensitiveFiles) = %d, want 1", got)
	}
	item := policy.SensitiveFiles[0]
	if got, want := item.Tags, []string{"env", "secret"}; !equalStrings(got, want) {
		t.Fatalf("Tags = %#v, want %#v", got, want)
	}
	if got, want := item.Crypto.Recipients, []string{"age1team", "age1ci"}; !equalStrings(got, want) {
		t.Fatalf("Recipients = %#v, want %#v", got, want)
	}
	if !item.HasAction || !item.Action.CommitBlock {
		t.Fatalf("action mapping was not loaded: %#v", item.Action)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
