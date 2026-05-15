package domain

import (
	"strings"
	"testing"
)

func TestValidatePolicyRequiresSopsAgeRecipients(t *testing.T) {
	policy := Policy{SensitiveFiles: []SensitiveFile{{
		Path:          ".env",
		EncryptedPath: ".env.sops.yaml",
		DecryptedPath: ".env",
		Reason:        "local env",
		HasAction:     true,
		HasCrypto:     true,
		Action: Action{
			Encrypt: true,
		},
		Crypto: Crypto{
			Method:         "sops-age",
			EncryptCommand: "encrypt",
			DecryptCommand: "decrypt",
			ManualEdit:     "decrypted",
		},
	}}}

	errs := ValidatePolicy(policy)
	if !containsError(errs, "crypto.recipients") {
		t.Fatalf("ValidatePolicy errors = %#v, want crypto.recipients error", errs)
	}
}

func TestValidateRepoPathRejectsAbsolutePath(t *testing.T) {
	if err := ValidateRepoPath("/tmp/.env"); err == nil {
		t.Fatal("ValidateRepoPath accepted absolute path")
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
