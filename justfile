help:
	@just --list

build:
	go build -o $out/ipwatch {{justfile_directory()}}/cmd/ipwatch

run *ARGS:
	go run ./cmd/ipwatch -hook {{justfile_directory()}}/test-hook.sh -debug -filter=!IsLoopback {{ARGS}}

update:
	#!/usr/bin/env bash
	go get -u all
	go mod tidy
	export NIX_PATH="nixpkgs=$(nix flake prefetch nixpkgs --json | jq --raw-output '.storePath')"
	newvendorHash="$(nix-prefetch \
	 "{ sha256 }: ((import <nixpkgs> {}).callPackage ./. {}).goModules.overrideAttrs (_: { vendorHash = sha256; })")"
	sed -i "s|vendorHash.*|vendorHash = \"$newvendorHash\";|" default.nix
