help:
	@just --list

build:
	go build -o $out/ipwatch ./cmd/ipwatch

check: build
	go test ./...
	revive -set_exit_status=1 ./...
	staticcheck ./...

update:
	#!/usr/bin/env bash
	go get -u all
	export NIX_PATH="nixpkgs=$(nix flake prefetch nixpkgs --json | jq --raw-output '.storePath')"
	newvendorSha256="$(nix-prefetch \
	 "{ sha256 }: ((import <nixpkgs> {}).callPackage ./. {}).go-modules.overrideAttrs (_: { vendorSha256 = sha256; })")"
	sed -i "s|vendorSha256.*|vendorSha256 = \"$newvendorSha256\";|" default.nix

run:
	go run ./cmd/ipwatch -debug -4 -hook=internal:echo -filter=!IsLoopback
