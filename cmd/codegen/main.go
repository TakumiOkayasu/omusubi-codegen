package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/TakumiOkayasu/omusubi-codegen/internal/generator"
	"github.com/TakumiOkayasu/omusubi-codegen/internal/model"
	"github.com/TakumiOkayasu/omusubi-codegen/internal/parser"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "omusubi-codegen",
		Short: "Omusubi プラットフォーム コードジェネレーター",
		Long: `Omusubi 組み込みフレームワーク用のコード生成ツールです。
インターフェース定義から C++ 実装のスケルトンを自動生成します。`,
		Version: fmt.Sprintf("%s (commit: %s, built at: %s)", version, commit, date),
	}

	cmd.AddCommand(newGenerateCmd())
	cmd.AddCommand(newParseCmd())

	return cmd
}

func newGenerateCmd() *cobra.Command {
	var (
		repoPath      string
		className     string
		outputDir     string
		templateDir   string
		useLegacyName bool
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "抽象基底クラスから C++ 実装を生成",
		Long: `omusubi リポジトリ内の抽象基底クラスから純粋仮想関数を
オーバーライドした C++ 実装ファイル (.hpp/.cpp) を生成します。

このコマンドは以下の処理を行います:
1. 指定されたリポジトリ内の抽象基底クラスを検索
2. すべての純粋仮想メソッドを抽出
3. 派生クラス名の入力を促す
4. 空のメソッド本体を持つヘッダーとソースファイルを生成`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(repoPath, className, outputDir, templateDir, useLegacyName)
		},
	}

	cmd.Flags().StringVarP(&repoPath, "repo", "r", "",
		"omusubi リポジトリへのパス（省略可、未指定時は自動検出）\n"+
			"例: --repo /path/to/omusubi/include")

	cmd.Flags().StringVarP(&className, "class", "c", "",
		"派生クラス名のプレフィックス（省略可、未指定時はプロンプト表示）\n"+
			"例: --class M5Stack（M5StackXxx クラスを生成）")

	cmd.Flags().StringVarP(&outputDir, "output", "o", ".",
		"生成ファイル (.hpp/.cpp) の出力ディレクトリ\n"+
			"例: --output ./src")

	cmd.Flags().StringVarP(&templateDir, "templates", "t", "internal/generator/templates",
		"テンプレートディレクトリ（上級者向け）\n"+
			"デフォルト: 組み込みテンプレート")

	cmd.Flags().BoolVar(&useLegacyName, "legacy-name", false,
		"アルファ版サポート用のレガシー 'pre-omusubi' ディレクトリ名を使用\n"+
			"'omusubi' の代わりに 'pre-omusubi' を検索\n"+
			"このフラグは正式リリース後に削除予定")

	return cmd
}

func runGenerate(repoPath, className, outputDir, templateDir string, useLegacyName bool) error {
	// Create parser
	p := parser.New()
	if p == nil {
		return fmt.Errorf("パーサーの初期化に失敗しました: tree-sitter C++言語の取得に失敗")
	}

	// Auto-detect workspace if paths not provided
	var coreLibPath string
	if repoPath == "" {
		libNameDisplay := "omusubi"
		if useLegacyName {
			libNameDisplay = "pre-omusubi (alpha)"
		}
		fmt.Printf("%s ワークスペースを自動検出中...\n", libNameDisplay)
		detectedCore, _, err := parser.DetectWorkspace(".", useLegacyName)
		if err != nil {
			return fmt.Errorf("ワークスペースの自動検出に失敗しました: %w\n--repo を明示的に指定してください", err)
		}
		coreLibPath = detectedCore
		fmt.Printf("✓ コアライブラリを検出: %s\n", coreLibPath)
		repoPath = coreLibPath + "/include"
	}

	fmt.Println("リポジトリ内の抽象クラスを検索中...")
	fileInfos, err := p.ParseDirectory(repoPath)
	if err != nil {
		return fmt.Errorf("リポジトリの解析に失敗しました: %w", err)
	}

	if len(fileInfos) == 0 {
		return fmt.Errorf("%s に抽象クラスが見つかりませんでした", repoPath)
	}

	// Collect all abstract classes
	type ClassOption struct {
		ClassInfo *model.ClassInfo
		Name      string
		FullName  string
		FilePath  string
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
		return fmt.Errorf("%s に抽象クラスが見つかりませんでした", repoPath)
	}

	// Ask if user wants to select all
	var selectAllResponse string
	selectAllPrompt := &survey.Select{
		Message: fmt.Sprintf("%d 個の抽象クラスが見つかりました。すべて選択しますか？", len(classOptions)),
		Options: []string{"はい - すべてのクラスを選択", "いいえ - 個別に選択"},
		Default: "いいえ - 個別に選択",
	}
	if err := survey.AskOne(selectAllPrompt, &selectAllResponse); err != nil {
		return fmt.Errorf("ユーザー入力の取得に失敗しました: %w", err)
	}

	selectAllClasses := strings.HasPrefix(selectAllResponse, "はい")

	var selectedIndices []int

	if selectAllClasses {
		// Select all classes
		for i := range classOptions {
			selectedIndices = append(selectedIndices, i)
		}
		fmt.Printf("\n✓ %d 個のクラスをすべて選択しました\n", len(selectedIndices))
	} else {
		// Create options for survey
		var options []string
		for _, opt := range classOptions {
			options = append(options, fmt.Sprintf("%s（%s より）", opt.FullName, opt.FilePath))
		}

		// Multi-select prompt
		prompt := &survey.MultiSelect{
			Message: "実装する抽象クラスを選択してください:",
			Options: options,
			Help:    "矢印キーで移動、スペースで選択/解除、エンターで確定",
		}

		if err := survey.AskOne(prompt, &selectedIndices); err != nil {
			return fmt.Errorf("ユーザー選択の取得に失敗しました: %w", err)
		}

		if len(selectedIndices) == 0 {
			return fmt.Errorf("クラスが選択されていません")
		}
	}

	// Get derived class name prefix(DeviceName) if not provided
	classPrefix := className
	if classPrefix == "" {
		fmt.Print("\n派生クラス名のプレフィックスを入力してください（<デバイス名><基底クラス名> として使用）: ")
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
	fmt.Printf("\n実装ファイルを生成中...\n")
	successCount := 0
	for _, idx := range selectedIndices {
		selectedClass := classOptions[idx]
		// Capitalize first letter of base class name to ensure proper CamelCase
		baseName := selectedClass.Name
		if len(baseName) > 0 {
			baseName = strings.ToUpper(baseName[:1]) + baseName[1:]
		}
		derivedName := classPrefix + baseName

		fmt.Printf("\n[%d/%d] %s を %s から生成中...\n",
			successCount+1, len(selectedIndices), derivedName, selectedClass.FullName)

		if err := gen.GenerateImplementation(selectedClass.ClassInfo, derivedName); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s の生成に失敗しました: %v\n", derivedName, err)
			continue
		}

		fmt.Printf("  ✓ %s.hpp と %s.cpp を生成しました\n", derivedName, derivedName)
		successCount++
	}

	// Output summary and helpful information
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("✓ コード生成が完了しました！\n")
	fmt.Printf("生成成功: %d/%d クラス\n", successCount, len(selectedIndices))

	// Get absolute path for output directory
	absOutputDir, _ := filepath.Abs(outputDir)
	fmt.Printf("出力ディレクトリ: %s\n", absOutputDir)

	// Print helpful information for PlatformIO project setup
	fmt.Printf("\n" + strings.Repeat("-", 60) + "\n")
	fmt.Printf("【PlatformIO プロジェクトのセットアップ】\n\n")
	fmt.Printf("1. プロジェクトを初期化:\n")
	fmt.Printf("   pio project init --board <ボード名>\n")
	fmt.Printf("   例: pio project init --board m5stack-core-esp32\n\n")

	fmt.Printf("2. platformio.ini に以下を追加:\n")
	fmt.Printf("   [env:<環境名>]\n")
	fmt.Printf("   platform = espressif32\n")
	fmt.Printf("   board = <ボード名>\n")
	fmt.Printf("   framework = arduino\n")
	if coreLibPath != "" {
		relCorePath, _ := filepath.Rel(".", coreLibPath)
		fmt.Printf("   lib_deps = \n")
		fmt.Printf("       %s\n", relCorePath)
	}
	fmt.Printf("   build_flags = \n")
	fmt.Printf("       -I include\n")
	if coreLibPath != "" {
		fmt.Printf("       -I %s/include\n", coreLibPath)
	}

	fmt.Printf("\n✓ 生成されたファイルは既にPlatformIO構造で配置済み:\n")
	fmt.Printf("   - ヘッダー (.hpp/.h): %s/include/\n", absOutputDir)
	fmt.Printf("   - ソース (.cpp): %s/src/\n", absOutputDir)

	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")

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
		Short: "omusubi リポジトリを解析して抽象クラスを一覧表示",
		Long: `omusubi リポジトリ内のすべてのヘッダーファイルを解析し、
抽象クラスとその純粋仮想メソッドの情報を表示します。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runParse(repoPath, verbose)
		},
	}

	cmd.Flags().StringVarP(&repoPath, "repo", "r", "", "omusubi リポジトリへのパス（必須）")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "詳細なメソッド情報を表示")
	_ = cmd.MarkFlagRequired("repo")

	return cmd
}

func runParse(repoPath string, verbose bool) error {
	p := parser.New()
	if p == nil {
		return fmt.Errorf("パーサーの初期化に失敗しました: tree-sitter C++言語の取得に失敗")
	}

	fmt.Printf("リポジトリを解析中: %s\n", repoPath)
	fileInfos, err := p.ParseDirectory(repoPath)
	if err != nil {
		return fmt.Errorf("リポジトリの解析に失敗しました: %w", err)
	}

	if len(fileInfos) == 0 {
		fmt.Println("抽象クラスが見つかりませんでした。")
		return nil
	}

	fmt.Printf("\n抽象クラスを含む %d 個のファイルが見つかりました:\n\n", len(fileInfos))

	for _, fileInfo := range fileInfos {
		fmt.Printf("ファイル: %s\n", fileInfo.Path)
		if fileInfo.Namespace != "" {
			fmt.Printf("名前空間: %s\n", fileInfo.Namespace)
		}

		for _, classInfo := range fileInfo.Classes {
			fmt.Printf("\n  クラス: %s", classInfo.Name)
			if len(classInfo.BaseClasses) > 0 {
				fmt.Printf("（継承元: %s）", strings.Join(classInfo.BaseClasses, ", "))
			}
			fmt.Println()

			pureVirtualCount := 0
			for _, method := range classInfo.Methods {
				if method.IsPureVirtual {
					pureVirtualCount++
				}
			}

			fmt.Printf("  純粋仮想メソッド数: %d\n", pureVirtualCount)

			if verbose && pureVirtualCount > 0 {
				fmt.Println("  メソッド:")
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
