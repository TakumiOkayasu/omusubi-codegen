package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/TakumiOkayasu/omusubi-platform-codegen/internal/generator"
	"github.com/TakumiOkayasu/omusubi-platform-codegen/internal/model"
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

	fmt.Println("Searching for abstract classes in repository...")
	fileInfos, err := p.ParseDirectory(repoPath)
	if err != nil {
		return fmt.Errorf("failed to parse repository: %w", err)
	}

	if len(fileInfos) == 0 {
		return fmt.Errorf("no abstract classes found in %s", repoPath)
	}

	// Collect all abstract classes
	type ClassOption struct {
		Name      string
		FullName  string
		FilePath  string
		ClassInfo *model.ClassInfo
	}

	var classOptions []ClassOption
	for _, fileInfo := range fileInfos {
		for _, classInfo := range fileInfo.Classes {
			if classInfo.IsAbstract {
				fullName := classInfo.Name
				if classInfo.Namespace != "" {
					fullName = classInfo.Namespace + "::" + classInfo.Name
				}

				classOptions = append(classOptions, ClassOption{
					Name:      classInfo.Name,
					FullName:  fullName,
					FilePath:  fileInfo.Path,
					ClassInfo: &classInfo,
				})
			}
		}
	}

	if len(classOptions) == 0 {
		return fmt.Errorf("no abstract classes found in %s", repoPath)
	}

	// Ask if user wants to select all
	var selectAllResponse string
	selectAllPrompt := &survey.Select{
		Message: fmt.Sprintf("Found %d abstract classes. Select all?", len(classOptions)),
		Options: []string{"Yes - Select all classes", "No - Choose individually"},
		Default: "No - Choose individually",
	}
	if err := survey.AskOne(selectAllPrompt, &selectAllResponse); err != nil {
		return fmt.Errorf("failed to get user input: %w", err)
	}

	selectAllClasses := strings.HasPrefix(selectAllResponse, "Yes")

	var selectedIndices []int

	if selectAllClasses {
		// Select all classes
		for i := range classOptions {
			selectedIndices = append(selectedIndices, i)
		}
		fmt.Printf("\n✓ All %d classes selected\n", len(selectedIndices))
	} else {
		// Create options for survey
		var options []string
		for _, opt := range classOptions {
			options = append(options, fmt.Sprintf("%s (from %s)", opt.FullName, opt.FilePath))
		}

		// Multi-select prompt
		prompt := &survey.MultiSelect{
			Message: "Select abstract classes to implement:",
			Options: options,
			Help:    "Use arrow keys to move, space to select/deselect, enter to confirm",
		}

		if err := survey.AskOne(prompt, &selectedIndices); err != nil {
			return fmt.Errorf("failed to get user selection: %w", err)
		}

		if len(selectedIndices) == 0 {
			return fmt.Errorf("no classes selected")
		}
	}

	// Get derived class name prefix(DeviceName) if not provided
	classPrefix := className
	if classPrefix == "" {
		fmt.Print("\nEnter derived class name prefix (will be used as: <DeviceName><BaseClassName>): ")
		classPrefix = readLine()
		if classPrefix == "" {
			classPrefix = "My"
		}
	}

	// Create generator
	gen := generator.New(generator.Config{
		TemplateDir: templateDir,
		OutputDir:   outputDir,
	})

	// Generate files for each selected class
	fmt.Printf("\nGenerating files...\n")
	successCount := 0
	for _, idx := range selectedIndices {
		selectedClass := classOptions[idx]
		derivedName := classPrefix + selectedClass.Name

		fmt.Printf("\n[%d/%d] Generating %s from %s...\n",
			successCount+1, len(selectedIndices), derivedName, selectedClass.FullName)

		if err := gen.GenerateImplementation(selectedClass.ClassInfo, derivedName); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ Failed to generate %s: %v\n", derivedName, err)
			continue
		}

		fmt.Printf("  ✓ Generated %s.hpp and %s.cpp\n", derivedName, derivedName)
		successCount++
	}

	fmt.Printf("\n✓ Code generation completed!\n")
	fmt.Printf("Successfully generated: %d/%d classes\n", successCount, len(selectedIndices))
	fmt.Printf("Output directory: %s\n", outputDir)
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
