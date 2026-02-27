.PHONY: setup check-commit-msg

## setup: configure git to use the project's .githooks directory.
setup:
	git config core.hooksPath .githooks

## check-commit-msg: validate a commit message from stdin (useful in CI).
check-commit-msg:
	@tmp=$$(mktemp) && trap 'rm -f "$$tmp"' EXIT && \
	cat > "$$tmp" && \
	.githooks/commit-msg "$$tmp"
