# Directories for miscellaneous files for the local environment
ROOT_DIR=$(abspath $(dir $(lastword $(MAKEFILE_LIST))))
LOCAL_DIR=$(ROOT_DIR)/.local
LOCAL_BIN_DIR=$(LOCAL_DIR)/bin

# Go packages for the tools
PKG_golangci_lint=github.com/golangci/golangci-lint/cmd/golangci-lint
PKG_gotestsum=gotest.tools/gotestsum

.PHONY: run
run:
	@echo "${COLOR_GREEN}Running the application...${COLOR_RESET}"
	@PORT=9191 go run ./cmd/coda

.PHONY: test
test: ${LOCAL_BIN_DIR}
	@echo "${COLOR_GREEN}Running tests...${COLOR_RESET}"
	@GOBIN=${LOCAL_BIN_DIR} go install ${PKG_gotestsum}
	@${LOCAL_BIN_DIR}/gotestsum ${GOTESTSUM_ARGS} -- ${GO_TEST_FLAGS}  -coverprofile="coverage.txt" -covermode=atomic ./.../...

.PHONY: open-coverage
open-coverage:
	@go tool cover -html=coverage.txt

.PHONY: lint
lint: ${LOCAL_BIN_DIR}
	@echo "${COLOR_GREEN}Running linter...${COLOR_RESET}"
	@GOBIN=${LOCAL_BIN_DIR} go install ${PKG_golangci_lint}
	@${LOCAL_BIN_DIR}/golangci-lint run --fix

.PHONY: $(LOCAL_BIN_DIR)
$(LOCAL_BIN_DIR):
	@mkdir -p $(LOCAL_BIN_DIR)
