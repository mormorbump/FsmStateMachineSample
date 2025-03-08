package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// ワークスペース設定
var rootLevelScripts = []string{"scripts"}

func main() {
	// ワークスペース一覧を取得
	workspaces, err := getWorkspaces()
	if err != nil {
		fmt.Printf("警告: ワークスペース情報の取得に失敗しました: %v\n", err)
		workspaces = []string{} // 空の配列で続行
	}

	// Git でステージングされたファイル一覧を取得
	changedFiles, err := getGitStagedFiles()
	if err != nil {
		fmt.Printf("エラー: Gitの変更ファイル取得に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// 変更されたファイルからワークスペースとスクリプトを特定
	changedPaths := make(map[string]bool)
	for _, file := range changedFiles {
		if file == "" {
			continue
		}

		// ワークスペースの変更をチェック
		for _, workspace := range workspaces {
			if strings.HasPrefix(file, workspace+"/") {
				changedPaths[workspace] = true
				break
			}
		}

		// ルートレベルのスクリプト変更をチェック
		for _, scriptDir := range rootLevelScripts {
			if strings.HasPrefix(file, scriptDir+"/") {
				changedPaths[scriptDir] = true
				break
			}
		}

		// ルートレベルの .go ファイルをチェック
		if strings.HasSuffix(file, ".go") && !strings.Contains(file, "/") {
			changedPaths["root"] = true
		}
	}

	// 何も変更がない場合は終了
	if len(changedPaths) == 0 {
		fmt.Println("No relevant changes detected. Skipping checks.")
		os.Exit(0)
	}

	// 変更されたパスを表示
	paths := make([]string, 0, len(changedPaths))
	for path := range changedPaths {
		paths = append(paths, path)
	}
	fmt.Printf("Changed paths: %v\n", paths)

	// フォーマットチェック
	fmt.Println("\n📝 Running format check...")
	// Goファイルが存在するディレクトリを特定
	goFiles, err := findGoFiles()
	if err != nil {
		fmt.Printf("警告: Goファイルの検索に失敗しました: %v\n", err)
	}

	if len(goFiles) > 0 {
		goDirs := getGoDirs(goFiles)
		for _, dir := range goDirs {
			fmt.Printf("Formatting Go files in %s\n", dir)
			if err := runCommand("go", "fmt", dir); err != nil {
				fmt.Println("❌ Format check failed")
				os.Exit(1)
			}
		}
		fmt.Println("✅ Format check passed")
	} else {
		fmt.Println("⚠️ No Go files found to format")
	}

	// 静的解析チェック
	fmt.Println("\n🔍 Running vet check...")
	if len(goFiles) > 0 {
		goDirs := getGoDirs(goFiles)
		for _, dir := range goDirs {
			fmt.Printf("Vetting Go files in %s\n", dir)
			if err := runCommand("go", "vet", dir); err != nil {
				fmt.Println("❌ Vet check failed")
				os.Exit(1)
			}
		}
		fmt.Println("✅ Vet check passed")
	} else {
		fmt.Println("⚠️ No Go files found to vet")
	}

	// リントチェック（golangci-lintがインストールされている場合）
	if commandExists("golangci-lint") {
		fmt.Println("\n🔍 Running lint check...")
		if len(goFiles) > 0 {
			goDirs := getGoDirs(goFiles)
			for _, dir := range goDirs {
				fmt.Printf("Linting Go files in %s\n", dir)
				if err := runCommand("golangci-lint", "run", dir); err != nil {
					fmt.Println("❌ Lint check failed")
					os.Exit(1)
				}
			}
			fmt.Println("✅ Lint check passed")
		} else {
			fmt.Println("⚠️ No Go files found to lint")
		}
	}

	// 変更されたワークスペース/スクリプトに対してテストを実行
	if len(goFiles) > 0 {
		for path := range changedPaths {
			if path == "root" || path == "scripts" {
				continue // ルートとscriptsのテストはスキップ
			}

			// パスにGoファイルが含まれているか確認
			hasGoFiles := false
			for _, file := range goFiles {
				if strings.HasPrefix(file, "./"+path+"/") {
					hasGoFiles = true
					break
				}
			}

			if !hasGoFiles {
				fmt.Printf("\n⚠️ No Go files found in %s, skipping tests\n", path)
				continue
			}

			fmt.Printf("\n🧪 Running tests for %s...\n", path)

			testPath := "./" + path + "/..."
			if err := runCommand("go", "test", "-v", testPath); err != nil {
				fmt.Printf("❌ Tests failed for %s\n", path)
				os.Exit(1)
			} else {
				fmt.Printf("✅ Tests passed for %s\n", path)
			}
		}
	} else {
		fmt.Println("\n⚠️ No Go files found to test")
	}

	// 依存関係チェック
	for _, arg := range os.Args[1:] {
		if arg == "--check-deps" {
			fmt.Println("\n📦 Running dependency check...")
			// go.modファイルが存在するか確認
			if _, err := os.Stat("go.mod"); err == nil {
				if err := runCommand("go", "mod", "verify"); err != nil {
					fmt.Println("❌ Dependency check failed")
					os.Exit(1)
				} else {
					fmt.Println("✅ Dependency check passed")
				}
			} else {
				fmt.Println("⚠️ No go.mod file found, skipping dependency check")
			}
			break
		}
	}

	fmt.Println("\n✅ All checks passed successfully!")
}

// getWorkspaces はgo.modファイルからモジュール名を取得し、ワークスペースを推測します
func getWorkspaces() ([]string, error) {
	// カレントディレクトリを表示
	pwd, err := os.Getwd()
	if err == nil {
		fmt.Printf("【デバッグ】getWorkspaces実行時のカレントディレクトリ: %s\n", pwd)
	}

	// go.modファイルを読み込む
	content, err := ioutil.ReadFile("go.mod")
	if err != nil {
		fmt.Printf("【デバッグ】go.modファイル読み込みエラー: %v\n", err)
		return nil, err
	}

	fmt.Println("【デバッグ】go.modファイルが正常に読み込まれました")

	// モジュール名を取得
	scanner := bufio.NewScanner(bytes.NewReader(content))
	moduleName := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "module ") {
			moduleName = line[7:]
			fmt.Printf("【デバッグ】モジュール名: %s\n", moduleName)
			break
		}
	}

	if moduleName == "" {
		fmt.Println("【デバッグ】モジュール名が見つかりませんでした")
	}

	// プロジェクト内のディレクトリを検索
	entries, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Printf("【デバッグ】ディレクトリ一覧取得エラー: %v\n", err)
		return nil, err
	}

	fmt.Println("【デバッグ】検出されたディレクトリ:")
	var workspaces []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") && !strings.HasPrefix(entry.Name(), "_") {
			// 隠しディレクトリや特殊ディレクトリを除外
			dirName := entry.Name()
			fmt.Printf("  - %s", dirName)

			// ディレクトリ内にGoファイルが存在するか確認
			hasGoFiles, err := directoryContainsGoFiles(dirName)
			if err != nil {
				fmt.Printf(" (エラー: %v)\n", err)
				continue
			}

			if hasGoFiles {
				workspaces = append(workspaces, dirName)
				fmt.Println(" (Goファイルを含むため、ワークスペースとして追加)")
			} else {
				// サブディレクトリにGoファイルがあるか確認
				hasGoFilesInSubdir, err := subdirectoryContainsGoFiles(dirName)
				if err != nil {
					fmt.Printf(" (サブディレクトリ確認エラー: %v)\n", err)
					continue
				}

				if hasGoFilesInSubdir {
					workspaces = append(workspaces, dirName)
					fmt.Println(" (サブディレクトリにGoファイルを含むため、ワークスペースとして追加)")
				} else {
					fmt.Println(" (Goファイルを含まないため、ワークスペースとして認識されません)")
				}
			}
		}
	}

	fmt.Printf("【デバッグ】検出されたワークスペース: %v\n", workspaces)
	return workspaces, nil
}

// directoryContainsGoFiles はディレクトリ内にGoファイルが存在するか確認します
func directoryContainsGoFiles(dirName string) (bool, error) {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		return false, err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".go") {
			return true, nil
		}
	}

	return false, nil
}

// subdirectoryContainsGoFiles はサブディレクトリにGoファイルが存在するか確認します
func subdirectoryContainsGoFiles(dirName string) (bool, error) {
	// findコマンドを使用してサブディレクトリ内のGoファイルを検索
	cmd := exec.Command("find", dirName, "-name", "*.go", "-type", "f", "-not", "-path", "*/\\.*")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return len(output) > 0, nil
}

// getGitStagedFiles はGitでステージングされたファイル一覧を取得します
func getGitStagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	files := strings.Split(string(output), "\n")

	// デバッグ用：ステージングされたファイル一覧を表示
	fmt.Println("【デバッグ】ステージングされたファイル一覧:")
	for _, file := range files {
		if file != "" {
			fmt.Printf("  %s\n", file)
		}
	}

	return files, nil
}

// runCommand は指定されたコマンドを実行します
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// commandExists は指定されたコマンドが存在するかチェックします
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// findGoFiles はプロジェクト内のGoファイルを検索します
func findGoFiles() ([]string, error) {
	fmt.Println("検索コマンド: find . -name \"*.go\" -type f -not -path \"*/\\.*\"")
	cmd := exec.Command("find", ".", "-name", "*.go", "-type", "f", "-not", "-path", "*/\\.*")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	files := strings.Split(string(output), "\n")
	var goFiles []string
	for _, file := range files {
		if file != "" {
			goFiles = append(goFiles, file)
		}
	}

	fmt.Println("検索結果:")
	if len(goFiles) == 0 {
		fmt.Println("  Goファイルが見つかりませんでした")
	} else {
		for _, file := range goFiles {
			fmt.Printf("  %s\n", file)
		}
	}

	// カレントディレクトリも表示
	pwd, err := os.Getwd()
	if err == nil {
		fmt.Printf("カレントディレクトリ: %s\n", pwd)
	}

	return goFiles, nil
}

// getGoDirs はGoファイルが存在するディレクトリのリストを返します
func getGoDirs(goFiles []string) []string {
	dirMap := make(map[string]bool)
	for _, file := range goFiles {
		dir := "."
		if lastSlash := strings.LastIndex(file, "/"); lastSlash != -1 {
			dir = file[:lastSlash]
		}
		dirMap[dir] = true
	}

	dirs := make([]string, 0, len(dirMap))
	for dir := range dirMap {
		dirs = append(dirs, dir)
	}
	return dirs
}
