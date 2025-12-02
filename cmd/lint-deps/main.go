package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// DependencyLinter checks for cross-module imports in a modular monolith
// Usage: go run cmd/lint-deps/main.go -root ./internal/modules

var (
	rootDir string
	verbose bool
)

func init() {
	flag.StringVar(&rootDir, "root", "./internal/modules", "Root directory of modules")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
}

func main() {
	flag.Parse()

	// Define modules that should be isolated
	modules := []string{"auth", "product", "user"}

	// Allowed shared imports
	allowedShared := []string{
		"go-modular-monolith/internal/shared",
		"go-modular-monolith/pkg",
	}

	violations := []string{}
	totalFiles := 0

	for _, module := range modules {
		modulePath := filepath.Join(rootDir, module)

		err := filepath.Walk(modulePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}

			totalFiles++

			// Parse the Go file
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
			if err != nil {
				fmt.Printf("Warning: Could not parse %s: %v\n", path, err)
				return nil
			}

			// Check each import
			for _, imp := range f.Imports {
				importPath := strings.Trim(imp.Path.Value, `"`)

				// Skip standard library and external imports
				if !strings.Contains(importPath, "go-modular-monolith") {
					continue
				}

				// Check if it's a shared import (allowed)
				isShared := false
				for _, shared := range allowedShared {
					if strings.HasPrefix(importPath, shared) {
						isShared = true
						break
					}
				}
				if isShared {
					continue
				}

				// ACL (Anti-Corruption Layer) folders are allowed to import other modules
				// This is by design - ACL is the designated translation layer
				relPath, _ := filepath.Rel(rootDir, path)
				isACLFile := strings.Contains(relPath, "/acl/")
				if isACLFile {
					continue
				}

				// Check for cross-module imports
				for _, otherModule := range modules {
					if otherModule == module {
						continue
					}

					crossModulePattern := fmt.Sprintf("internal/modules/%s", otherModule)
					if strings.Contains(importPath, crossModulePattern) {
						violation := fmt.Sprintf(
							"❌ %s imports %s (cross-module dependency!)",
							relPath,
							importPath,
						)
						violations = append(violations, violation)

						if verbose {
							fmt.Println(violation)
						}
					}
				}

				// Check for old domain imports (deprecated)
				if strings.Contains(importPath, "internal/domain/") {
					relPath, _ := filepath.Rel(rootDir, path)
					violation := fmt.Sprintf(
						"⚠️  %s imports deprecated path %s (use internal/modules/<module>/domain instead)",
						relPath,
						importPath,
					)
					violations = append(violations, violation)

					if verbose {
						fmt.Println(violation)
					}
				}
			}

			return nil
		})

		if err != nil {
			fmt.Printf("Error walking module %s: %v\n", module, err)
		}
	}

	// Print summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("Dependency Lint Results\n")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Modules checked: %v\n", modules)
	fmt.Printf("Files scanned: %d\n", totalFiles)
	fmt.Printf("Violations found: %d\n", len(violations))
	fmt.Println(strings.Repeat("-", 60))

	if len(violations) > 0 {
		fmt.Println("\nViolations:")
		for _, v := range violations {
			fmt.Println(v)
		}
		fmt.Println("\n💡 To fix:")
		fmt.Println("   1. Use internal/modules/<module>/domain for module-specific types")
		fmt.Println("   2. Use internal/shared/* for shared types (events, errors, context)")
		fmt.Println("   3. Use Anti-Corruption Layer (ACL) for cross-module communication")
		os.Exit(1)
	}

	fmt.Println("\n✅ No cross-module dependencies found!")
}
