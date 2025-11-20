package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/TakumiOkayasu/omusubi-platform-codegen/internal/generator"
	"github.com/TakumiOkayasu/omusubi-platform-codegen/internal/parser"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "codegen",
		Short: "Omusubi Platform Code Generator",
		Long: `A code generation tool for the Omusubi embedded framework.
Automatically generates C++ implementation skeletons, tests, and documentation
from interface definitions.`,
		Version: fmt.Sprintf("%s (commit: %s, built at: %s)", version, commit, date),
	}

	cmd.AddCommand(newGenerateCmd())
	cmd.AddCommand(newParseCmd())

	return cmd
}

func newGenerateCmd() *cobra.Command {
	var (
		repoPath     string
		baseClass    string
		className    string
		outputDir    string
		templateDir  string
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate C++ implementation from abstract base class",
		Long: `Generate C++ implementation (.hpp and .cpp) files by overriding
pure virtual functions from an abstract base class found in the pre-omusubi repository.

The command will:
1. Search for the abstract base class in the specified repository
2. Extract all pure virtual methods
3. Prompt for the derived class name
4. Generate header and source files with empty method bodies`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(repoPath, baseClass, className, outputDir, templateDir)
		},
	}

	cmd.Flags().StringVarP(&repoPath, "repo", "r", "", "Path to pre-omusubi repository (required)")
	cmd.Flags().StringVarP(&baseClass, "base", "b", "", "Base class name to search for (optional, will prompt if not provided)")
	cmd.Flags().StringVarP(&className, "class", "c", "", "Derived class name (optional, will prompt if not provided)")
	cmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for generated files")
	cmd.Flags().StringVarP(&templateDir, "templates", "t", "internal/template/templates", "Template directory")
	cmd.MarkFlagRequired("repo")

	return cmd
}

func runGenerate(repoPath, baseClass, className, outputDir, templateDir string) error {
	// Create parser
	p := parser.New()

	// If base class not provided, list available abstract classes
	if baseClass == "" {
		fmt.Println("Searching for abstract classes in repository...")
		fileInfos, err := p.ParseDirectory(repoPath)
		if err != nil {
			return fmt.Errorf("failed to parse repository: %w", err)
		}

		if len(fileInfos) == 0 {
			return fmt.Errorf("no abstract classes found in %s", repoPath)
		}

		fmt.Println("\nAvailable abstract classes:")
		for _, fileInfo := range fileInfos {
			for _, classInfo := range fileInfo.Classes {
				if classInfo.IsAbstract {
					fullName := classInfo.Name
					if classInfo.Namespace != "" {
						fullName = classInfo.Namespace + "::" + classInfo.Name
					}
					fmt.Printf("  - %s (from %s)\n", fullName, fileInfo.Path)
				}
			}
		}

		fmt.Print("\nEnter base class name: ")
		baseClass = readLine()
		if baseClass == "" {
			return fmt.Errorf("base class name is required")
		}
	}

	// Find the abstract class
	fmt.Printf("Searching for abstract class '%s'...\n", baseClass)
	classInfo, err := p.FindAbstractClass(repoPath, baseClass)
	if err != nil {
		return err
	}

	fmt.Printf("Found abstract class: %s\n", classInfo.Name)
	if classInfo.Namespace != "" {
		fmt.Printf("Namespace: %s\n", classInfo.Namespace)
	}

	// Count pure virtual methods
	pureVirtualCount := 0
	for _, method := range classInfo.Methods {
		if method.IsPureVirtual {
			pureVirtualCount++
		}
	}
	fmt.Printf("Pure virtual methods: %d\n", pureVirtualCount)

	// Get derived class name
	if className == "" {
		fmt.Print("\nEnter derived class name: ")
		className = readLine()
		if className == "" {
			return fmt.Errorf("derived class name is required")
		}
	}

	// Create generator
	gen := generator.New(generator.Config{
		TemplateDir: templateDir,
		OutputDir:   outputDir,
	})

	// Generate files
	fmt.Printf("\nGenerating files for %s...\n", className)
	if err := gen.GenerateImplementation(classInfo, className); err != nil {
		return fmt.Errorf("failed to generate implementation: %w", err)
	}

	fmt.Printf("\n✓ Code generation completed successfully!\n")
	fmt.Printf("Generated files in: %s\n", outputDir)
	return nil
}

func readLine() string {
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func newParseCmd() *cobra.Command {
	var (
		repoPath string
		verbose  bool
	)

	cmd := &cobra.Command{
		Use:   "parse",
		Short: "Parse pre-omusubi repository and list abstract classes",
		Long: `Parse all header files in the pre-omusubi repository and display
information about abstract classes and their pure virtual methods.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runParse(repoPath, verbose)
		},
	}

	cmd.Flags().StringVarP(&repoPath, "repo", "r", "", "Path to pre-omusubi repository (required)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed method information")
	cmd.MarkFlagRequired("repo")

	return cmd
}

func runParse(repoPath string, verbose bool) error {
	p := parser.New()

	fmt.Printf("Parsing repository: %s\n", repoPath)
	fileInfos, err := p.ParseDirectory(repoPath)
	if err != nil {
		return fmt.Errorf("failed to parse repository: %w", err)
	}

	if len(fileInfos) == 0 {
		fmt.Println("No abstract classes found.")
		return nil
	}

	fmt.Printf("\nFound %d file(s) with abstract classes:\n\n", len(fileInfos))

	for _, fileInfo := range fileInfos {
		fmt.Printf("File: %s\n", fileInfo.Path)
		if fileInfo.Namespace != "" {
			fmt.Printf("Namespace: %s\n", fileInfo.Namespace)
		}

		for _, classInfo := range fileInfo.Classes {
			fmt.Printf("\n  Class: %s", classInfo.Name)
			if len(classInfo.BaseClasses) > 0 {
				fmt.Printf(" (extends: %s)", strings.Join(classInfo.BaseClasses, ", "))
			}
			fmt.Println()

			pureVirtualCount := 0
			for _, method := range classInfo.Methods {
				if method.IsPureVirtual {
					pureVirtualCount++
				}
			}

			fmt.Printf("  Pure virtual methods: %d\n", pureVirtualCount)

			if verbose && pureVirtualCount > 0 {
				fmt.Println("  Methods:")
				for _, method := range classInfo.Methods {
					if method.IsPureVirtual {
						fmt.Printf("    - %s %s(", method.ReturnType, method.Name)
						for i, param := range method.Parameters {
							if i > 0 {
								fmt.Print(", ")
							}
							fmt.Printf("%s %s", param.Type, param.Name)
						}
						fmt.Print(")")
						if method.IsConst {
							fmt.Print(" const")
						}
						fmt.Println()
					}
				}
			}
		}
		fmt.Println()
	}

	return nil
}
