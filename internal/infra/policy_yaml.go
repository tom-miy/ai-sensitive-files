package infra

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/tom-miy/ai-sensitive-files/internal/domain"
)

func LoadPolicy(path string) (domain.Policy, error) {
	f, err := os.Open(path)
	if err != nil {
		return domain.Policy{}, err
	}
	defer f.Close()

	var policy domain.Policy
	var current *domain.SensitiveFile
	inSensitiveFiles := false
	inAction := false
	inCrypto := false
	pendingList := ""

	scanner := bufio.NewScanner(f)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		raw := stripComment(scanner.Text())
		if strings.TrimSpace(raw) == "" {
			continue
		}
		trimmed := strings.TrimSpace(raw)
		indent := leadingSpaces(raw)
		if pendingList != "" && current != nil && indent > 2 && strings.HasPrefix(trimmed, "- ") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			if err := appendListValue(current, pendingList, value); err != nil {
				return domain.Policy{}, fmt.Errorf("line %d: %w", lineNo, err)
			}
			continue
		}
		switch {
		case trimmed == "sensitive_files:":
			inSensitiveFiles = true
			inAction = false
			inCrypto = false
			pendingList = ""
		case strings.HasPrefix(trimmed, "- "):
			if !inSensitiveFiles {
				return domain.Policy{}, fmt.Errorf("line %d: list item outside sensitive_files", lineNo)
			}
			policy.SensitiveFiles = append(policy.SensitiveFiles, domain.SensitiveFile{})
			current = &policy.SensitiveFiles[len(policy.SensitiveFiles)-1]
			inAction = false
			inCrypto = false
			pendingList = ""
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			if rest != "" {
				key, value, ok := splitKeyValue(rest)
				if !ok {
					return domain.Policy{}, fmt.Errorf("line %d: expected key: value", lineNo)
				}
				if err := assignField(current, key, value); err != nil {
					return domain.Policy{}, fmt.Errorf("line %d: %w", lineNo, err)
				}
			}
		case current != nil:
			key, value, ok := splitKeyValue(trimmed)
			if !ok {
				return domain.Policy{}, fmt.Errorf("line %d: expected key: value", lineNo)
			}
			if key == "action" && value == "" {
				inAction = true
				inCrypto = false
				pendingList = ""
				current.HasAction = true
				continue
			}
			if key == "crypto" && value == "" {
				inAction = false
				inCrypto = true
				pendingList = ""
				current.HasCrypto = true
				continue
			}
			if key == "tags" && value == "" {
				inAction = false
				inCrypto = false
				pendingList = "tags"
				continue
			}
			if inAction {
				pendingList = ""
				if err := assignAction(current, key, value); err != nil {
					return domain.Policy{}, fmt.Errorf("line %d: %w", lineNo, err)
				}
				continue
			}
			if inCrypto {
				if key == "recipients" && value == "" {
					pendingList = "recipients"
					continue
				}
				pendingList = ""
				if err := assignCrypto(current, key, value); err != nil {
					return domain.Policy{}, fmt.Errorf("line %d: %w", lineNo, err)
				}
				continue
			}
			pendingList = ""
			if err := assignField(current, key, value); err != nil {
				return domain.Policy{}, fmt.Errorf("line %d: %w", lineNo, err)
			}
		default:
			return domain.Policy{}, fmt.Errorf("line %d: expected sensitive_files", lineNo)
		}
	}
	if err := scanner.Err(); err != nil {
		return domain.Policy{}, err
	}
	return policy, nil
}

func leadingSpaces(line string) int {
	count := 0
	for _, r := range line {
		if r != ' ' {
			break
		}
		count++
	}
	return count
}

func stripComment(line string) string {
	inQuote := false
	for i, r := range line {
		if r == '"' {
			inQuote = !inQuote
		}
		if r == '#' && !inQuote {
			return line[:i]
		}
	}
	return line
}

func splitKeyValue(s string) (string, string, bool) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), true
}

func assignField(item *domain.SensitiveFile, key, value string) error {
	switch key {
	case "path":
		item.Path = unquote(value)
	case "encrypted_path":
		item.EncryptedPath = unquote(value)
	case "decrypted_path":
		item.DecryptedPath = unquote(value)
	case "reason":
		item.Reason = unquote(value)
	case "tags":
		item.Tags = parseList(value)
	case "action":
		item.HasAction = true
		return fmt.Errorf("action must be a mapping")
	case "crypto":
		item.HasCrypto = true
		return fmt.Errorf("crypto must be a mapping")
	default:
		return fmt.Errorf("unknown field %q", key)
	}
	return nil
}

func assignCrypto(item *domain.SensitiveFile, key, value string) error {
	switch key {
	case "method":
		item.Crypto.Method = unquote(value)
	case "secret_ref":
		item.Crypto.SecretRef = unquote(value)
	case "recipients":
		item.Crypto.Recipients = parseList(value)
	case "encrypt_command":
		item.Crypto.EncryptCommand = unquote(value)
	case "decrypt_command":
		item.Crypto.DecryptCommand = unquote(value)
	case "manual_edit":
		item.Crypto.ManualEdit = unquote(value)
	default:
		return fmt.Errorf("unknown crypto field %q", key)
	}
	return nil
}

func assignAction(item *domain.SensitiveFile, key, value string) error {
	b, err := strconv.ParseBool(unquote(value))
	if err != nil {
		return fmt.Errorf("action.%s must be true or false", key)
	}
	switch key {
	case "aiignore":
		item.Action.AIIgnore = b
	case "gitignore":
		item.Action.GitIgnore = b
	case "encrypt":
		item.Action.Encrypt = b
	case "commit_block":
		item.Action.CommitBlock = b
	default:
		return fmt.Errorf("unknown action %q", key)
	}
	return nil
}

func appendListValue(item *domain.SensitiveFile, key, value string) error {
	switch key {
	case "tags":
		item.Tags = append(item.Tags, unquote(value))
	case "recipients":
		item.Crypto.Recipients = append(item.Crypto.Recipients, unquote(value))
	default:
		return fmt.Errorf("unknown list field %q", key)
	}
	return nil
}

func parseList(value string) []string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	if strings.TrimSpace(value) == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(value, ",") {
		out = append(out, unquote(strings.TrimSpace(part)))
	}
	return out
}

func unquote(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
		return value[1 : len(value)-1]
	}
	return value
}
