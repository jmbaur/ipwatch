# vim: ft=make

help:
	@just --list

build:
	go build -o $out/ipwatch ./cmd/ipwatch

run:
	go run ./cmd/ipwatch -debug -filter=!IsLoopback

check: build
	go test ./...
	revive -set_exit_status=1 ./...
	staticcheck ./...

update:
	#!/usr/bin/env bash
	go get -u all
	go mod tidy
	export NIX_PATH="nixpkgs=$(nix flake prefetch nixpkgs --json | jq --raw-output '.storePath')"
	newvendorSha256="$(nix-prefetch \
	 "{ sha256 }: ((import <nixpkgs> {}).callPackage ./. {}).goModules.overrideAttrs (_: { vendorSha256 = sha256; })")"
	sed -i "s|vendorSha256.*|vendorSha256 = \"$newvendorSha256\";|" default.nix
