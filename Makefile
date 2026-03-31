.PHONY: build

build:
	goreleaser build --skip=validate --clean --snapshot