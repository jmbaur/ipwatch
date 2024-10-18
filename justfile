help:
	@just --list

update:
	#!/usr/bin/env bash
	go get -u all
	go mod tidy
	export NIX_PATH="nixpkgs=$(nix flake prefetch nixpkgs --json | jq --raw-output '.storePath')"
	newvendorHash="$(nix-prefetch \
	 "{ sha256 }: ((import <nixpkgs> {}).callPackage ./. {}).goModules.overrideAttrs (_: { vendorHash = sha256; })")"
	sed -i "s|vendorHash.*|vendorHash = \"$newvendorHash\";|" default.nix
