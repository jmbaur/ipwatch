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
    }:
    {
      overlays.default = _: prev: { ipwatch = prev.callPackage ./. { }; };
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
        default = self.devShells.${system}.ci.overrideAttrs (old: {
          inherit
            (git-hooks.lib.${system}.run {
              src = ./.;
              hooks.nixfmt.enable = true;
              hooks.nixfmt.package = pkgs.nixfmt-rfc-style;
              hooks.govet.enable = true;
              hooks.revive.enable = true;
              hooks.gofmt.enable = true;
            })
            shellHook
            ;
        });
        ci = pkgs.mkShell {
          inputsFrom = [ pkgs.ipwatch ];
          buildInputs = with pkgs; [
            go-tools
            just
            nix-prefetch
            revive
          ];
        };
      }) self.legacyPackages;
      packages = nixpkgs.lib.mapAttrs (_: pkgs: {
        default = pkgs.ipwatch;
        test = pkgs.callPackage ./test.nix { module = self.nixosModules.default; };
      }) self.legacyPackages;
    };
}
