package domain

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Policy struct {
	SensitiveFiles []SensitiveFile `json:"sensitive_files"`
}

type SensitiveFile struct {
	Path          string   `json:"path"`
	EncryptedPath string   `json:"encrypted_path,omitempty"`
	DecryptedPath string   `json:"decrypted_path,omitempty"`
	Reason        string   `json:"reason"`
	Tags          []string `json:"tags,omitempty"`
	Crypto        Crypto   `json:"crypto,omitempty"`
	Action        Action   `json:"action"`
	HasAction     bool     `json:"-"`
	HasCrypto     bool     `json:"-"`
}

type Crypto struct {
	Method         string   `json:"method,omitempty"`
	SecretRef      string   `json:"secret_ref,omitempty"`
	Recipients     []string `json:"recipients,omitempty"`
	EncryptCommand string   `json:"encrypt_command,omitempty"`
	DecryptCommand string   `json:"decrypt_command,omitempty"`
	ManualEdit     string   `json:"manual_edit,omitempty"`
}

type Action struct {
	AIIgnore    bool `json:"aiignore"`
	GitIgnore   bool `json:"gitignore"`
	Encrypt     bool `json:"encrypt"`
	CommitBlock bool `json:"commit_block"`
}

func (a Action) Any() bool {
	return a.AIIgnore || a.GitIgnore || a.Encrypt || a.CommitBlock
}

func (a Action) Names() []string {
	var names []string
	if a.AIIgnore {
		names = append(names, "aiignore")
	}
	if a.GitIgnore {
		names = append(names, "gitignore")
	}
	if a.Encrypt {
		names = append(names, "encrypt")
	}
	if a.CommitBlock {
		names = append(names, "commit_block")
	}
	return names
}

func ValidatePolicy(p Policy) []error {
	var errs []error
	if len(p.SensitiveFiles) == 0 {
		errs = append(errs, fmt.Errorf("sensitive_files must contain at least one entry"))
	}
	for i, item := range p.SensitiveFiles {
		prefix := fmt.Sprintf("sensitive_files[%d]", i)
		if strings.TrimSpace(item.Path) == "" {
			errs = append(errs, fmt.Errorf("%s.path is required", prefix))
		}
		if strings.TrimSpace(item.Reason) == "" {
			errs = append(errs, fmt.Errorf("%s.reason is required", prefix))
		}
		if !item.HasAction {
			errs = append(errs, fmt.Errorf("%s.action is required", prefix))
		} else if !item.Action.Any() {
			errs = append(errs, fmt.Errorf("%s.action must enable at least one action", prefix))
		}
		if err := ValidateRepoPath(item.Path); err != nil {
			errs = append(errs, fmt.Errorf("%s.path: %w", prefix, err))
		}
		if strings.TrimSpace(item.EncryptedPath) != "" {
			if err := ValidateRepoPath(item.EncryptedPath); err != nil {
				errs = append(errs, fmt.Errorf("%s.encrypted_path: %w", prefix, err))
			}
		}
		if item.Action.Encrypt && strings.TrimSpace(item.DecryptedPath) == "" {
			errs = append(errs, fmt.Errorf("%s.decrypted_path is required when action.encrypt is true", prefix))
		}
		if strings.TrimSpace(item.DecryptedPath) != "" {
			if err := ValidateDecryptedPath(item.DecryptedPath); err != nil {
				errs = append(errs, fmt.Errorf("%s.decrypted_path: %w", prefix, err))
			}
		}
		if item.Action.Encrypt {
			if !item.HasCrypto {
				errs = append(errs, fmt.Errorf("%s.crypto is required when action.encrypt is true", prefix))
			}
			if strings.TrimSpace(item.Crypto.Method) == "" {
				errs = append(errs, fmt.Errorf("%s.crypto.method is required when action.encrypt is true", prefix))
			}
			if IsExternalSecretMethod(item.Crypto.Method) && strings.TrimSpace(item.Crypto.SecretRef) == "" {
				errs = append(errs, fmt.Errorf("%s.crypto.secret_ref is required when crypto.method is %s", prefix, item.Crypto.Method))
			}
			if item.Crypto.Method == "sops-age" && len(item.Crypto.Recipients) == 0 {
				errs = append(errs, fmt.Errorf("%s.crypto.recipients must include at least one age public key when crypto.method is sops-age", prefix))
			}
			if !IsExternalSecretMethod(item.Crypto.Method) && strings.TrimSpace(item.Crypto.EncryptCommand) == "" {
				errs = append(errs, fmt.Errorf("%s.crypto.encrypt_command is required when action.encrypt is true", prefix))
			}
			if strings.TrimSpace(item.Crypto.DecryptCommand) == "" {
				errs = append(errs, fmt.Errorf("%s.crypto.decrypt_command is required when action.encrypt is true", prefix))
			}
			if err := ValidateManualEdit(item.Crypto.ManualEdit); err != nil {
				errs = append(errs, fmt.Errorf("%s.crypto.manual_edit: %w", prefix, err))
			}
		}
	}
	return errs
}

func IsExternalSecretMethod(method string) bool {
	switch method {
	case "1password", "op", "bitwarden", "bw":
		return true
	default:
		return false
	}
}

func ValidateManualEdit(value string) error {
	switch value {
	case "decrypted", "encrypted", "none":
		return nil
	case "":
		return fmt.Errorf("must be one of decrypted, encrypted, none")
	default:
		return fmt.Errorf("must be one of decrypted, encrypted, none")
	}
}

func ValidateRepoPath(pattern string) error {
	if strings.HasPrefix(pattern, "/") {
		return fmt.Errorf("absolute paths are not allowed; use a path relative to the repository root")
	}
	return validatePattern(pattern)
}

func ValidateDecryptedPath(pattern string) error {
	return validatePattern(pattern)
}

func validatePattern(pattern string) error {
	if strings.Contains(pattern, "\\") {
		return fmt.Errorf("use forward slashes from the repository root, not backslashes")
	}
	if strings.Contains(pattern, "\x00") {
		return fmt.Errorf("NUL byte is not allowed")
	}
	if strings.Contains(pattern, "[") || strings.Contains(pattern, "]") {
		if _, err := filepath.Match(pattern, "sample"); err != nil {
			return fmt.Errorf("invalid wildcard pattern: %v", err)
		}
	}
	for i, segment := range strings.Split(pattern, "/") {
		if i == 0 && segment == "" && strings.HasPrefix(pattern, "/") {
			continue
		}
		if segment == "" {
			return fmt.Errorf("empty path segment is not allowed")
		}
	}
	return nil
}
