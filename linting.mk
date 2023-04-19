### linting options below
### lint: for local environment
### lint-ci: simulates the CI linting outcome on a local environment

TMP_DIR = /tmp/sarunask

LINT_VERSION = 1.52.2

LINT_DIR = $(TMP_DIR)/golangci-lint/$(LINT_VERSION)
LINT_BIN = $(LINT_DIR)/golangci-lint

CMD = golangci-lint run --allow-parallel-runners -c .golangci.yml
LINT_ARCHIVE = golangci-lint-$(LINT_VERSION)-$(OSNAME)-$(ARCH).tar.gz
LINT_ARCHIVE_DEST = $(TMP_DIR)/$(LINT_ARCHIVE)

# Run this on local machine.
# It downloads a version of golangci-lint and execute it locally.
.PHONY: lint
lint: $(LINT_BIN)
	$(LINT_DIR)/$(CMD)

# install a local golangci-lint if not found.
$(LINT_BIN):
	curl -L --create-dirs \
		https://github.com/golangci/golangci-lint/releases/download/v$(LINT_VERSION)/$(LINT_ARCHIVE) \
		--output $(LINT_ARCHIVE_DEST)
	mkdir -p $(LINT_DIR)/
	tar -xf $(LINT_ARCHIVE_DEST) --strip-components=1 -C $(LINT_DIR)/
	chmod +x $(LINT_BIN)
	rm -f $(LINT_ARCHIVE_DEST)
