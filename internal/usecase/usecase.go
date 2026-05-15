package usecase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tom-miy/ai-sensitive-files/internal/domain"
)

type GeneratedFile struct {
	Path    string
	Content []byte
}

type CheckResult struct {
	Errors   []error
	Warnings []string
}

func BuildGeneratedFiles(policy domain.Policy) []GeneratedFile {
	ai := pathsFor(policy, func(a domain.Action) bool { return a.AIIgnore })
	git := pathsFor(policy, func(a domain.Action) bool { return a.GitIgnore })
	encrypt := encryptionTargets(policy)
	decrypt := decryptionTargets(policy)
	secretSources := secretSourceTargets(policy)
	deny := map[string][]string{"denyRead": ai}
	denyJSON, _ := json.MarshalIndent(deny, "", "  ")
	return []GeneratedFile{
		{Path: ".aiignore", Content: ignoreContent("ai-sensitive-files: AI ignore entries", ai)},
		{Path: ".cursorignore", Content: ignoreContent("ai-sensitive-files: Cursor ignore entries", ai)},
		{Path: ".copilotignore", Content: ignoreContent("ai-sensitive-files: Copilot ignore entries", ai)},
		{Path: ".gitignore.ai-sensitive-files", Content: ignoreContent("ai-sensitive-files: gitignore entries", git)},
		{Path: "generated/claude-code-deny-read.json", Content: append(denyJSON, '\n')},
		{Path: "generated/ai-agent-guidance.md", Content: guidanceContent(policy)},
		{Path: "generated/ai-sensitive-files.summary.md", Content: summaryContent(policy)},
		{Path: "generated/encryption-targets.txt", Content: ignoreContent("ai-sensitive-files: SOPS/age encryption targets", encrypt)},
		{Path: "generated/decryption-targets.txt", Content: ignoreContent("ai-sensitive-files: decrypted plaintext output targets", decrypt)},
		{Path: "generated/secret-sources.txt", Content: ignoreContent("ai-sensitive-files: external secret manager references", secretSources)},
		{Path: "generated/crypto-plan.md", Content: cryptoPlanContent(policy)},
	}
}

func WriteGeneratedFiles(outDir string, files []GeneratedFile, force bool) error {
	for _, file := range files {
		target := filepath.Join(outDir, file.Path)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if existing, err := os.ReadFile(target); err == nil && !bytes.Equal(existing, file.Content) {
			if !force {
				return fmt.Errorf("%s exists and differs; rerun with --force to create a .bak and overwrite", target)
			}
			if err := os.WriteFile(target+".bak", existing, 0o644); err != nil {
				return err
			}
		}
		if err := os.WriteFile(target, file.Content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func Check(policy domain.Policy, repoDir string) CheckResult {
	var result CheckResult
	tracked, gitErr := gitTracked(repoDir)
	reportedGitErr := false
	for _, item := range policy.SensitiveFiles {
		if item.Action.CommitBlock {
			if gitErr != nil {
				if !reportedGitErr {
					result.Errors = append(result.Errors, fmt.Errorf("git tracked file check failed: %w", gitErr))
					reportedGitErr = true
				}
				continue
			}
			for _, pattern := range sensitivePatterns(item) {
				for _, path := range matchingPaths(pattern, tracked) {
					result.Errors = append(result.Errors, fmt.Errorf("commit_block path is git tracked: %s", path))
				}
			}
		}
		if item.Action.Encrypt {
			for _, path := range expandExisting(repoDir, plaintextPath(item)) {
				if isPlaintext(path) && item.Crypto.ManualEdit != "decrypted" {
					result.Errors = append(result.Errors, fmt.Errorf("decrypted plaintext exists but crypto.manual_edit is %s: %s", item.Crypto.ManualEdit, rel(repoDir, path)))
				}
			}
			result = appendSyncResult(result, checkCryptoSync(repoDir, item))
		}
	}
	result.Errors = append(result.Errors, checkGitIgnore(policy, repoDir)...)
	for _, file := range BuildGeneratedFiles(policy) {
		target := filepath.Join(repoDir, file.Path)
		existing, err := os.ReadFile(target)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("generated file missing or unreadable: %s", file.Path))
			continue
		}
		if !bytes.Equal(existing, file.Content) {
			result.Errors = append(result.Errors, fmt.Errorf("generated file is out of sync: %s", file.Path))
		}
	}
	return result
}

func List(policy domain.Policy) string {
	var b strings.Builder
	b.WriteString("Sensitive Files\n\n")
	for _, item := range policy.SensitiveFiles {
		b.WriteString("- " + item.Path + "\n")
		b.WriteString("  reason: " + item.Reason + "\n")
		if item.EncryptedPath != "" {
			b.WriteString("  encrypted_path: " + item.EncryptedPath + "\n")
		}
		if item.DecryptedPath != "" {
			b.WriteString("  decrypted_path: " + item.DecryptedPath + "\n")
		}
		if item.HasCrypto {
			b.WriteString("  crypto: " + item.Crypto.Method + ", manual_edit=" + item.Crypto.ManualEdit + "\n")
			if item.Crypto.SecretRef != "" {
				b.WriteString("  secret_ref: " + item.Crypto.SecretRef + "\n")
			}
		}
		b.WriteString("  actions: " + strings.Join(item.Action.Names(), ", ") + "\n\n")
	}
	return b.String()
}

func pathsFor(policy domain.Policy, ok func(domain.Action) bool) []string {
	seen := map[string]bool{}
	for _, item := range policy.SensitiveFiles {
		if ok(item.Action) {
			seen[item.Path] = true
			if item.DecryptedPath != "" && isRepoRelativePath(item.DecryptedPath) {
				seen[item.DecryptedPath] = true
			}
		}
	}
	var paths []string
	for path := range seen {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func encryptionTargets(policy domain.Policy) []string {
	seen := map[string]bool{}
	for _, item := range policy.SensitiveFiles {
		if item.Action.Encrypt && item.EncryptedPath != "" {
			seen[item.EncryptedPath] = true
		}
	}
	return sortedKeys(seen)
}

func decryptionTargets(policy domain.Policy) []string {
	seen := map[string]bool{}
	for _, item := range policy.SensitiveFiles {
		if item.Action.Encrypt {
			target := item.DecryptedPath
			if target == "" {
				target = item.Path
			}
			source := secretOrEncryptedSource(item)
			seen[source+" -> "+target] = true
		}
	}
	return sortedKeys(seen)
}

func secretSourceTargets(policy domain.Policy) []string {
	seen := map[string]bool{}
	for _, item := range policy.SensitiveFiles {
		if item.Action.Encrypt && item.Crypto.SecretRef != "" {
			seen[item.Crypto.Method+": "+item.Crypto.SecretRef+" -> "+plaintextPath(item)] = true
		}
	}
	return sortedKeys(seen)
}

func sortedKeys(seen map[string]bool) []string {
	var paths []string
	for path := range seen {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func ignoreContent(title string, paths []string) []byte {
	var b strings.Builder
	b.WriteString("# Generated by ai-sensitive-files. Edit .ai-sensitive-files/sensitive-files.yaml, then regenerate.\n")
	b.WriteString("# " + title + "\n\n")
	for _, path := range paths {
		b.WriteString(path + "\n")
	}
	return []byte(b.String())
}

func summaryContent(policy domain.Policy) []byte {
	var b strings.Builder
	b.WriteString("# AI Sensitive Files Summary\n\n")
	b.WriteString("Generated from policy. Do not edit this file by hand.\n\n")
	for _, item := range policy.SensitiveFiles {
		b.WriteString("- `" + item.Path + "`: " + item.Reason + " (" + strings.Join(item.Action.Names(), ", ") + ")")
		if item.DecryptedPath != "" {
			b.WriteString("; decrypted path: `" + item.DecryptedPath + "`")
		}
		if item.EncryptedPath != "" {
			b.WriteString("; encrypted path: `" + item.EncryptedPath + "`")
		}
		if item.HasCrypto {
			b.WriteString("; crypto: `" + item.Crypto.Method + "`, manual edit: `" + item.Crypto.ManualEdit + "`")
			if item.Crypto.SecretRef != "" {
				b.WriteString("; secret ref: `" + item.Crypto.SecretRef + "`")
			}
		}
		b.WriteString("\n")
	}
	return []byte(b.String())
}

func guidanceContent(policy domain.Policy) []byte {
	var b strings.Builder
	b.WriteString("# AI Agent Guidance\n\n")
	b.WriteString("Generated from `.ai-sensitive-files/sensitive-files.yaml`. Do not edit this file by hand.\n\n")
	b.WriteString("AI coding agents and editor assistants should not read, summarize, index, or paste the following paths into prompts:\n\n")
	for _, path := range pathsFor(policy, func(a domain.Action) bool { return a.AIIgnore }) {
		b.WriteString("- `" + path + "`\n")
	}
	b.WriteString("\nBefore committing, run:\n\n")
	b.WriteString("```bash\n")
	b.WriteString("ai-sensitive-files check --config .ai-sensitive-files/sensitive-files.yaml\n")
	b.WriteString("```\n")
	return []byte(b.String())
}

func cryptoPlanContent(policy domain.Policy) []byte {
	var b strings.Builder
	b.WriteString("# AI Sensitive Files Crypto Plan\n\n")
	b.WriteString("Generated from policy. Review commands before running them.\n\n")
	for _, item := range policy.SensitiveFiles {
		if !item.Action.Encrypt {
			continue
		}
		b.WriteString("## " + item.Path + "\n\n")
		b.WriteString("- method: `" + item.Crypto.Method + "`\n")
		if item.EncryptedPath != "" {
			b.WriteString("- encrypted_path: `" + item.EncryptedPath + "`\n")
		}
		if item.Crypto.SecretRef != "" {
			b.WriteString("- secret_ref: `" + item.Crypto.SecretRef + "`\n")
		}
		if len(item.Crypto.Recipients) > 0 {
			b.WriteString("- recipients:\n")
			for _, recipient := range item.Crypto.Recipients {
				b.WriteString("  - `" + recipient + "`\n")
			}
		}
		b.WriteString("- decrypted_path: `" + plaintextPath(item) + "`\n")
		b.WriteString("- manual_edit: `" + item.Crypto.ManualEdit + "`\n")
		if item.Crypto.EncryptCommand != "" {
			b.WriteString("- encrypt: `" + renderCryptoCommand(item.Crypto.EncryptCommand, item) + "`\n")
		}
		b.WriteString("- decrypt: `" + renderCryptoCommand(item.Crypto.DecryptCommand, item) + "`\n\n")
	}
	return []byte(b.String())
}

func renderCryptoCommand(command string, item domain.SensitiveFile) string {
	replacer := strings.NewReplacer(
		"{path}", item.Path,
		"{encrypted_path}", encryptedPath(item),
		"{decrypted_path}", plaintextPath(item),
		"{secret_ref}", item.Crypto.SecretRef,
		"{age_recipients}", ageRecipientsArgs(item.Crypto.Recipients),
	)
	return replacer.Replace(command)
}

func ageRecipientsArgs(recipients []string) string {
	var args []string
	for _, recipient := range recipients {
		args = append(args, "--recipient "+recipient)
	}
	return strings.Join(args, " ")
}

func checkGitIgnore(policy domain.Policy, repoDir string) []error {
	required := pathsFor(policy, func(a domain.Action) bool { return a.GitIgnore })
	if len(required) == 0 {
		return nil
	}
	lines, err := readIgnoreLines(filepath.Join(repoDir, ".gitignore"))
	if err != nil {
		return []error{fmt.Errorf(".gitignore is missing; merge entries from .gitignore.ai-sensitive-files")}
	}
	var errs []error
	for _, path := range required {
		if !lines[path] {
			errs = append(errs, fmt.Errorf(".gitignore is missing generated entry: %s", path))
		}
	}
	return errs
}

func readIgnoreLines(path string) (map[string]bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := map[string]bool{}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines[line] = true
	}
	return lines, nil
}

func plaintextPath(item domain.SensitiveFile) string {
	if item.DecryptedPath != "" {
		return item.DecryptedPath
	}
	return item.Path
}

func encryptedPath(item domain.SensitiveFile) string {
	if item.EncryptedPath != "" {
		return item.EncryptedPath
	}
	return item.Path
}

func appendSyncResult(result CheckResult, sync CheckResult) CheckResult {
	result.Errors = append(result.Errors, sync.Errors...)
	result.Warnings = append(result.Warnings, sync.Warnings...)
	return result
}

func checkCryptoSync(repoDir string, item domain.SensitiveFile) CheckResult {
	if item.EncryptedPath == "" {
		return CheckResult{}
	}
	if strings.Contains(plaintextPath(item), "*") || strings.Contains(encryptedPath(item), "*") {
		return CheckResult{}
	}
	plain := filepath.Join(repoDir, plaintextPath(item))
	encrypted := filepath.Join(repoDir, encryptedPath(item))
	plainInfo, plainErr := os.Stat(plain)
	encryptedInfo, encryptedErr := os.Stat(encrypted)
	if plainErr == nil && os.IsNotExist(encryptedErr) {
		return missingEncryptedResult(item)
	}
	if os.IsNotExist(plainErr) && encryptedErr == nil {
		return CheckResult{Warnings: []string{fmt.Sprintf("encrypted path exists but decrypted path is missing; run decrypt_command before editing locally: %s -> %s", encryptedPath(item), plaintextPath(item))}}
	}
	if plainErr != nil || encryptedErr != nil {
		return CheckResult{}
	}
	if plainInfo.ModTime().After(encryptedInfo.ModTime()) {
		switch item.Crypto.ManualEdit {
		case "decrypted":
			return CheckResult{Errors: []error{fmt.Errorf("decrypted path is newer than encrypted path; run encrypt_command before commit: %s -> %s", plaintextPath(item), encryptedPath(item))}}
		case "none":
			return CheckResult{Errors: []error{fmt.Errorf("decrypted path changed but crypto.manual_edit is none; update encrypted file or discard plaintext: %s", plaintextPath(item))}}
		default:
			return CheckResult{Errors: []error{fmt.Errorf("decrypted path is newer than encrypted path but manual edits should target %s: %s", item.Crypto.ManualEdit, plaintextPath(item))}}
		}
	}
	if encryptedInfo.ModTime().After(plainInfo.ModTime()) {
		return CheckResult{Warnings: []string{fmt.Sprintf("encrypted path is newer than decrypted path; run decrypt_command before editing locally: %s -> %s", encryptedPath(item), plaintextPath(item))}}
	}
	return CheckResult{}
}

func missingEncryptedResult(item domain.SensitiveFile) CheckResult {
	switch item.Crypto.ManualEdit {
	case "decrypted":
		return CheckResult{Errors: []error{fmt.Errorf("decrypted path exists but encrypted path is missing; run encrypt_command before commit: %s -> %s", plaintextPath(item), encryptedPath(item))}}
	case "none":
		return CheckResult{Errors: []error{fmt.Errorf("decrypted path exists but crypto.manual_edit is none and encrypted path is missing: %s -> %s", plaintextPath(item), encryptedPath(item))}}
	default:
		return CheckResult{Errors: []error{fmt.Errorf("decrypted path exists but encrypted path is missing and manual edits should target %s: %s -> %s", item.Crypto.ManualEdit, plaintextPath(item), encryptedPath(item))}}
	}
}

func secretOrEncryptedSource(item domain.SensitiveFile) string {
	if item.Crypto.SecretRef != "" {
		return item.Crypto.Method + ":" + item.Crypto.SecretRef
	}
	return encryptedPath(item)
}

func sensitivePatterns(item domain.SensitiveFile) []string {
	seen := map[string]bool{item.Path: true}
	if item.DecryptedPath != "" && isRepoRelativePath(item.DecryptedPath) {
		seen[item.DecryptedPath] = true
	}
	return sortedKeys(seen)
}

func gitTracked(repoDir string) ([]string, error) {
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			paths = append(paths, line)
		}
	}
	return paths, nil
}

func matchingPaths(pattern string, paths []string) []string {
	var matches []string
	for _, path := range paths {
		if match(pattern, path) {
			matches = append(matches, path)
		}
	}
	return matches
}

func expandExisting(root, pattern string) []string {
	if !isRepoRelativePath(pattern) {
		path := filepath.Clean(pattern)
		if !filepath.IsAbs(path) {
			path = filepath.Clean(filepath.Join(root, pattern))
		}
		if strings.Contains(path, "*") {
			return nil
		}
		if _, err := os.Stat(path); err == nil {
			return []string{path}
		}
		return nil
	}
	var matches []string
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			if d != nil && d.IsDir() && (d.Name() == ".git" || d.Name() == "vendor") {
				return filepath.SkipDir
			}
			return nil
		}
		r := rel(root, path)
		if match(pattern, r) {
			matches = append(matches, path)
		}
		return nil
	})
	return matches
}

func isRepoRelativePath(path string) bool {
	path = filepath.ToSlash(filepath.Clean(path))
	return path != ".." && !strings.HasPrefix(path, "../") && !filepath.IsAbs(path)
}

func match(pattern, path string) bool {
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)
	if pattern == path {
		return true
	}
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**") + "/"
		return strings.HasPrefix(path, prefix)
	}
	ok, _ := filepath.Match(pattern, path)
	if ok {
		return true
	}
	if strings.Contains(pattern, "**/") {
		return match(strings.ReplaceAll(pattern, "**/", ""), path)
	}
	return false
}

func isPlaintext(path string) bool {
	name := filepath.Base(path)
	return !strings.HasSuffix(name, ".enc") && !strings.HasSuffix(name, ".encrypted") && !strings.Contains(name, ".sops.")
}

func rel(root, path string) string {
	r, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(r)
}
