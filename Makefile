-include .env

install_utils:
	@echo "install golangci-lint"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint_fix:
	golangci-lint run --fix --out-format colored-line-number

lint:
	golangci-lint run --config .golang-ci.yml ./...

test:
	go test -cover -race -coverpkg=./... -coverprofile=.testCoverage.txt.tmp ./...; \
	echo "Test coverage profile created"; \
	cat .testCoverage.txt.tmp | grep -v -E "mocks/|mock_|main.go|$GO_COVERAGE_EXCLUDE_PATTERN" > .testCoverage.txt; \
	echo "Coverage filtered"; \
	go tool cover -func .testCoverage.txt | tee .testCoverageSummary.txt; \
	echo "Coverage summary generated"