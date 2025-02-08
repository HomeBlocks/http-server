-include .env

install_utils:
	@echo "install golangci-lint"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "install gci"
	go install github.com/daixiang0/gci@latest
	@echo "install mockery"
	@go install github.com/vektra/mockery/v2@v2.46.3


lint_fix:
	golangci-lint run --fix --out-format colored-line-number

lint:
	golangci-lint run --config .golang-ci.yml ./...

gci_fix:
	gci write .

test:
	go test -cover -race -coverpkg=./... -coverprofile=.testCoverage.txt.tmp ./...; \
	echo "Test coverage profile created"; \
	cat .testCoverage.txt.tmp | grep -v -E "mocks/|mock_|main.go|$GO_COVERAGE_EXCLUDE_PATTERN" > .testCoverage.txt; \
	echo "Coverage filtered"; \
	go tool cover -func .testCoverage.txt | tee .testCoverageSummary.txt; \
	echo "Coverage summary generated"

.PHONY: mocks
mocks: delete-mocks
	@mockery --all --output=./mocks

delete-mocks:
	@find ./app -name 'mock_*' -delete