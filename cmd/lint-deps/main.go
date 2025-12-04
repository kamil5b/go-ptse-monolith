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

	// First pass: check modules
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

				// Rule: /module/** cannot import from /app/** and /infrastructure/**
				if strings.Contains(importPath, "internal/app/") || strings.Contains(importPath, "internal/infrastructure/") {
					relPath, _ := filepath.Rel(rootDir, path)
					violation := fmt.Sprintf(
						"❌ %s imports %s (modules cannot import from /app or /infrastructure!)",
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

	// Second pass: enforce /app import rules
	projectRoot := filepath.Dir(filepath.Dir(rootDir))
	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		totalFiles++
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			fmt.Printf("Warning: Could not parse %s: %v\n", path, err)
			return nil
		}
		relPath, _ := filepath.Rel(projectRoot, path)
		isAppFile := strings.Contains(relPath, "internal/app/")
		isBootstrapFile := strings.Contains(relPath, "cmd/bootstrap/")
		for _, imp := range f.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if !strings.Contains(importPath, "go-modular-monolith") {
				continue
			}
			// Rule: /app/** can import from anywhere except /cmd/bootstrap
			if isAppFile && strings.Contains(importPath, "cmd/bootstrap/") {
				violation := fmt.Sprintf(
					"❌ %s imports %s (/app cannot import from /cmd/bootstrap)",
					relPath,
					importPath,
				)
				violations = append(violations, violation)
				if verbose {
					fmt.Println(violation)
				}
			}
			// Rule: Only /cmd/bootstrap or /app/** can import from /app/**
			if strings.Contains(importPath, "internal/app/") && !(isAppFile || isBootstrapFile) {
				violation := fmt.Sprintf(
					"❌ %s imports %s (only /app/** or /cmd/bootstrap can import from /app)",
					relPath,
					importPath,
				)
				violations = append(violations, violation)
				if verbose {
					fmt.Println(violation)
				}
			}
			// Third pass: enforce module layer import rules
			modulesRoot := filepath.Join(filepath.Dir(rootDir), "modules")
			err = filepath.Walk(modulesRoot, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() || !strings.HasSuffix(path, ".go") {
					return nil
				}
				totalFiles++
				fset := token.NewFileSet()
				f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
				if err != nil {
					fmt.Printf("Warning: Could not parse %s: %v\n", path, err)
					return nil
				}
				relPath, _ := filepath.Rel(modulesRoot, path)
				// Determine module and layer
				parts := strings.Split(relPath, string(filepath.Separator))
				if len(parts) < 3 {
					return nil
				}
				moduleName := parts[0]
				layer := parts[1]
				for _, imp := range f.Imports {
					importPath := strings.Trim(imp.Path.Value, `"`)
					if !strings.Contains(importPath, "go-modular-monolith") {
						continue
					}
					// Domain layer
					if layer == "domain" {
						if strings.Contains(importPath, "/modules/") && !strings.Contains(importPath, "/modules/"+moduleName+"/domain") {
							violation := fmt.Sprintf(
								"❌ %s (domain) imports %s (domain can only import shared)",
								relPath,
								importPath,
							)
							violations = append(violations, violation)
							if verbose {
								fmt.Println(violation)
							}
						}
					}
					// Handler layer
					if layer == "handler" {
						if strings.Contains(importPath, "/modules/") && !strings.Contains(importPath, "/modules/"+moduleName+"/domain") {
							violation := fmt.Sprintf(
								"❌ %s (handler) imports %s (handler can only import own domain/shared)",
								relPath,
								importPath,
							)
							violations = append(violations, violation)
							if verbose {
								fmt.Println(violation)
							}
						}
					}
					// Service layer
					if layer == "service" {
						if strings.Contains(importPath, "/modules/") && !strings.Contains(importPath, "/modules/"+moduleName+"/domain") && !strings.Contains(importPath, "/modules/"+moduleName+"/acl") {
							violation := fmt.Sprintf(
								"❌ %s (service) imports %s (service can only import own domain/acl/shared)",
								relPath,
								importPath,
							)
							violations = append(violations, violation)
							if verbose {
								fmt.Println(violation)
							}
						}
					}
					// Repository layer
					if layer == "repository" {
						if strings.Contains(importPath, "/modules/") && !strings.Contains(importPath, "/modules/"+moduleName+"/domain") {
							violation := fmt.Sprintf(
								"❌ %s (repository) imports %s (repository can only import own domain/shared)",
								relPath,
								importPath,
							)
							violations = append(violations, violation)
							if verbose {
								fmt.Println(violation)
							}
						}
					}
					// ACL layer
					if layer == "acl" {
						// ACL can import own domain and other module domains
						if strings.Contains(importPath, "/modules/") && !strings.Contains(importPath, "/modules/"+moduleName+"/domain") && !strings.Contains(importPath, "/domain") {
							violation := fmt.Sprintf(
								"❌ %s (acl) imports %s (acl can only import domains)",
								relPath,
								importPath,
							)
							violations = append(violations, violation)
							if verbose {
								fmt.Println(violation)
							}
						}
					}
				}
				return nil
			})
			if err != nil {
				fmt.Printf("Error walking modules for layer rules: %v\n", err)
			}

			// Fourth pass: enforce shared kernel import rules
			sharedRoot := filepath.Join(filepath.Dir(rootDir), "shared")
			err = filepath.Walk(sharedRoot, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() || !strings.HasSuffix(path, ".go") {
					return nil
				}
				totalFiles++
				fset := token.NewFileSet()
				f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
				if err != nil {
					fmt.Printf("Warning: Could not parse %s: %v\n", path, err)
					return nil
				}
				relPath, _ := filepath.Rel(sharedRoot, path)
				for _, imp := range f.Imports {
					importPath := strings.Trim(imp.Path.Value, `"`)
					if strings.Contains(importPath, "/modules/") {
						violation := fmt.Sprintf(
							"❌ %s (shared) imports %s (shared can only import stdlib/external)",
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
				fmt.Printf("Error walking shared for import rules: %v\n", err)
			}

			// Fifth pass: check for cyclic dependencies (simple check)
			// (For brevity, only check if a module imports another module that imports it back)
			// This is a basic check, not a full graph cycle detection
			moduleImports := make(map[string]map[string]bool)
			err = filepath.Walk(modulesRoot, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() || !strings.HasSuffix(path, ".go") {
					return nil
				}
				relPath, _ := filepath.Rel(modulesRoot, path)
				parts := strings.Split(relPath, string(filepath.Separator))
				if len(parts) < 2 {
					return nil
				}
				moduleName := parts[0]
				if _, ok := moduleImports[moduleName]; !ok {
					moduleImports[moduleName] = make(map[string]bool)
				}
				fset := token.NewFileSet()
				f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
				if err != nil {
					return nil
				}
				for _, imp := range f.Imports {
					importPath := strings.Trim(imp.Path.Value, `"`)
					for _, otherModule := range modules {
						if otherModule != moduleName && strings.Contains(importPath, "/modules/"+otherModule+"/") {
							moduleImports[moduleName][otherModule] = true
						}
					}
				}
				return nil
			})
			if err != nil {
				fmt.Printf("Error walking modules for cyclic check: %v\n", err)
			}
			for m1, imports := range moduleImports {
				for m2 := range imports {
					if moduleImports[m2][m1] {
						violation := fmt.Sprintf(
							"❌ Cyclic dependency: %s <-> %s",
							m1,
							m2,
						)
						violations = append(violations, violation)
						if verbose {
							fmt.Println(violation)
						}
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking project for /app import rules: %v\n", err)
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
		fmt.Println("\n💡 Linting Rules:")
		fmt.Println("   1. /module/** cannot import from /app/** or /infrastructure/**")
		fmt.Println("   2. /app/** can import from anywhere except /cmd/bootstrap")
		fmt.Println("   3. Only /cmd/bootstrap or /app/** can import from /app/**")
		fmt.Println("   4. No cross-module dependencies within /modules")
		fmt.Println("   5. Use internal/modules/<module>/domain for module-specific types")
		fmt.Println("   6. Use internal/shared/* for shared types (events, errors, context)")
		fmt.Println("   7. Use Anti-Corruption Layer (ACL) for cross-module communication")
		os.Exit(1)
	}

	fmt.Println("\n✅ All dependency linting rules passed!")
}
