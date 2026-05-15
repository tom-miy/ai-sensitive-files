package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/tom-miy/ai-sensitive-files/internal/domain"
	"github.com/tom-miy/ai-sensitive-files/internal/infra"
	"github.com/tom-miy/ai-sensitive-files/internal/usecase"
)

func Run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "validate":
		return validate(args[1:])
	case "generate":
		return generate(args[1:])
	case "check":
		return check(args[1:])
	case "list":
		return list(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		usage()
		return 2
	}
}

func validate(args []string) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	config := fs.String("config", ".ai-sensitive-files/sensitive-files.yaml", "policy YAML path")
	jsonOut := fs.Bool("json", false, "print JSON output")
	if fs.Parse(args) != nil {
		return 2
	}
	p, ok := loadAndValidate(*config, *jsonOut)
	if !ok {
		return 1
	}
	if *jsonOut {
		printJSON(map[string]any{"ok": true, "entries": len(p.SensitiveFiles)})
	} else {
		fmt.Printf("policy valid: %s (%d sensitive files)\n", *config, len(p.SensitiveFiles))
	}
	return 0
}

func generate(args []string) int {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	config := fs.String("config", ".ai-sensitive-files/sensitive-files.yaml", "policy YAML path")
	out := fs.String("out", ".", "output repository root")
	force := fs.Bool("force", false, "backup and overwrite existing generated files")
	jsonOut := fs.Bool("json", false, "print JSON output")
	if fs.Parse(args) != nil {
		return 2
	}
	p, ok := loadAndValidate(*config, *jsonOut)
	if !ok {
		return 1
	}
	files := usecase.BuildGeneratedFiles(p)
	if err := usecase.WriteGeneratedFiles(*out, files, *force); err != nil {
		printErr(err, *jsonOut)
		return 1
	}
	if *jsonOut {
		var paths []string
		for _, file := range files {
			paths = append(paths, file.Path)
		}
		printJSON(map[string]any{"ok": true, "generated": paths})
	} else {
		fmt.Printf("generated %d files under %s\n", len(files), *out)
		for _, file := range files {
			fmt.Printf("- %s\n", file.Path)
		}
	}
	return 0
}

func check(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	config := fs.String("config", ".ai-sensitive-files/sensitive-files.yaml", "policy YAML path")
	repo := fs.String("repo", ".", "repository root to check")
	jsonOut := fs.Bool("json", false, "print JSON output")
	if fs.Parse(args) != nil {
		return 2
	}
	p, ok := loadAndValidate(*config, *jsonOut)
	if !ok {
		return 1
	}
	result := usecase.Check(p, *repo)
	if len(result.Errors) > 0 {
		if *jsonOut {
			printJSON(map[string]any{"ok": false, "errors": errStrings(result.Errors), "warnings": result.Warnings})
		} else {
			fmt.Println("sensitive file check failed")
			for _, err := range result.Errors {
				fmt.Printf("- %s\n", err)
			}
			printWarnings(result.Warnings)
		}
		return 1
	}
	if *jsonOut {
		printJSON(map[string]any{"ok": true, "warnings": result.Warnings})
	} else {
		fmt.Println("sensitive file check passed")
		printWarnings(result.Warnings)
	}
	return 0
}

func list(args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	config := fs.String("config", ".ai-sensitive-files/sensitive-files.yaml", "policy YAML path")
	jsonOut := fs.Bool("json", false, "print JSON output")
	if fs.Parse(args) != nil {
		return 2
	}
	p, ok := loadAndValidate(*config, *jsonOut)
	if !ok {
		return 1
	}
	if *jsonOut {
		printJSON(p)
	} else {
		fmt.Print(usecase.List(p))
	}
	return 0
}

func loadAndValidate(config string, jsonOut bool) (domain.Policy, bool) {
	p, err := infra.LoadPolicy(config)
	if err != nil {
		printErr(err, jsonOut)
		return domain.Policy{}, false
	}
	errs := domain.ValidatePolicy(p)
	if len(errs) > 0 {
		if jsonOut {
			printJSON(map[string]any{"ok": false, "errors": errStrings(errs)})
		} else {
			fmt.Fprintln(os.Stderr, "policy invalid")
			for _, err := range errs {
				fmt.Fprintf(os.Stderr, "- %s\n", err)
			}
		}
		return domain.Policy{}, false
	}
	return p, true
}

func printErr(err error, jsonOut bool) {
	if jsonOut {
		printJSON(map[string]any{"ok": false, "error": err.Error()})
		return
	}
	fmt.Fprintln(os.Stderr, err)
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func errStrings(errs []error) []string {
	out := make([]string, 0, len(errs))
	for _, err := range errs {
		out = append(out, err.Error())
	}
	return out
}

func printWarnings(warnings []string) {
	if len(warnings) == 0 {
		return
	}
	fmt.Println("warnings")
	for _, warning := range warnings {
		fmt.Printf("- %s\n", warning)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: ai-sensitive-files <validate|generate|check|list> [options]")
}
