# テストとカバレッジ関連のターゲット
.PHONY: test test-verbose coverage coverage-html clean

# カバレッジファイルの出力先
COVERAGE_FILE := coverage.out
# HTMLレポートの出力先
HTML_REPORT := coverage.html

# すべてのパッケージのテストを実行
test:
	@echo "すべてのパッケージのテストを実行します..."
	@go test ./... -coverprofile=$(COVERAGE_FILE)
	@echo "テスト完了"

# 詳細なテスト結果を表示
test-verbose:
	@echo "詳細なテスト結果を表示します..."
	@go test -v ./... -coverprofile=$(COVERAGE_FILE)
	@echo "テスト完了"

# カバレッジ情報を表示
coverage: test
	@echo "カバレッジ情報を表示します..."
	@go tool cover -func=$(COVERAGE_FILE)

# カバレッジレポートをHTML形式で表示
coverage-html: test
	@echo "カバレッジレポートをHTML形式で生成します..."
	@go tool cover -html=$(COVERAGE_FILE) -o $(HTML_REPORT)
	@echo "HTMLレポートが $(HTML_REPORT) に生成されました"
	@open $(HTML_REPORT)

# 生成されたファイルを削除
clean:
	@echo "生成されたファイルを削除します..."
	@rm -f $(COVERAGE_FILE) $(HTML_REPORT)
	@echo "クリーンアップ完了"