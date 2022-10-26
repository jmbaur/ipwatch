build:
	go build -o $out/ipwatch ./cmd/ipwatch

check: build
	staticcheck ./...
	go test ./...

update:
	#!/usr/bin/env bash
	go get -u all
	export NIX_PATH="nixpkgs=$(nix flake prefetch nixpkgs --json | jq --raw-output '.storePath')"
	newvendorSha256="$(nix-prefetch \
		 "{ sha256 }: let pkgs = import <nixpkgs> {}; in (pkgs.callPackage ./. {}).go-modules.overrideAttrs (_: { vendorSha256 = sha256; })")"
	sed -i "s|vendorSha256.*|vendorSha256 = \"$newvendorSha256\";|" default.nix

run:
	go run ./cmd/ipwatch -debug -4 -hook=internal:echo -filter=!IsLoopback
