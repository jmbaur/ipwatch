{
  description = "Run code on changes to network interfaces";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    git-hooks.inputs.nixpkgs.follows = "nixpkgs";
    git-hooks.url = "github:cachix/git-hooks.nix";
  };

  outputs =
    {
      self,
      nixpkgs,
      git-hooks,
    }@inputs:
    {
      overlays.default = final: _: { ipwatch = final.callPackage ./package.nix { }; };
      nixosModules.default = {
        nixpkgs.overlays = [ self.overlays.default ];
        imports = [ ./module.nix ];
      };
      legacyPackages =
        nixpkgs.lib.genAttrs
          [
            "aarch64-linux"
            "x86_64-linux"
          ]
          (
            system:
            import nixpkgs {
              inherit system;
              overlays = [ self.overlays.default ];
            }
          );
      devShells = nixpkgs.lib.mapAttrs (system: pkgs: {
        default = pkgs.mkShell {
          inputsFrom = [ pkgs.ipwatch ];
          inherit
            (git-hooks.lib.${system}.run {
              src = ./.;
              hooks.gofmt.enable = true;
              hooks.govet.enable = true;
              hooks.nixfmt-rfc-style.enable = true;
              hooks.revive.enable = true;
              hooks.staticcheck.enable = true;
            })
            shellHook
            ;
        };
      }) self.legacyPackages;
      apps = nixpkgs.lib.mapAttrs (_: pkgs: {
        updateDependencies = {
          type = "app";
          program = toString (
            pkgs.writeShellScript "update-dependencies" ''
              ${pkgs.lib.getExe pkgs.ipwatch.go} get -u all
              ${pkgs.lib.getExe pkgs.ipwatch.go} mod tidy
              export NIX_PATH="nixpkgs=$(nix flake prefetch nixpkgs --json | jq --raw-output '.storePath')"
              newvendorHash=$(nix build --impure --expr 'with import <nixpkgs> {}; (callPackage ./package.nix {}).goModules.overrideAttrs (_: {outputHash = ""; outputHashAlgo = "sha256";})' 2>&1 | grep 'got: ' | cut -d':' -f2 | xargs)
              if [[ -z $newvendorHash ]]; then
              	echo "failed to fetch new vendor hash"
              	exit 1
              fi
              sed -i "s|vendorHash.*|vendorHash = \"$newvendorHash\";|" package.nix
            ''
          );
        };
      }) self.legacyPackages;
      checks = nixpkgs.lib.mapAttrs (_: pkgs: {
        default = pkgs.callPackage ./test.nix { inherit inputs; };
      }) self.legacyPackages;
    };
}
