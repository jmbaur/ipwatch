{
  description = "Run code on changes to network interfaces";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    pre-commit-hooks.inputs.nixpkgs.follows = "nixpkgs";
    pre-commit-hooks.url = "github:cachix/pre-commit-hooks.nix";
  };

  outputs = inputs: with inputs; let
    forAllSystems = cb: nixpkgs.lib.genAttrs [ "aarch64-linux" "x86_64-linux" ] (system: cb {
      inherit system;
      pkgs = import nixpkgs { inherit system; overlays = [ self.overlays.default ]; };
    });
  in
  {
    overlays.default = _: prev: { ipwatch = prev.callPackage ./. { }; };
    nixosModules.default = {
      nixpkgs.overlays = [ self.overlays.default ];
      imports = [ ./module.nix ];
    };
    devShells = forAllSystems ({ pkgs, system, ... }: {
      default = self.devShells.${system}.ci.overrideAttrs (old: {
        inherit (pre-commit-hooks.lib.${system}.run {
          src = ./.;
          hooks = {
            nixpkgs-fmt.enable = true;
            govet.enable = true;
            revive.enable = true;
            gofmt = {
              enable = true;
              entry = "${pkgs.ipwatch.go}/bin/gofmt -w";
              types = [ "go" ];
            };
          };
        }) shellHook;
      });
      ci = pkgs.mkShell {
        buildInputs = with pkgs; [ just go-tools nix-prefetch ];
        inherit (pkgs.ipwatch) CGO_ENABLED nativeBuildInputs;
      };
    });
    packages = forAllSystems ({ pkgs, ... }: {
      default = pkgs.ipwatch;
      test = pkgs.callPackage ./test.nix { module = self.nixosModules.default; };
    });
    apps = forAllSystems ({ pkgs, ... }: {
      default = { type = "app"; program = "${pkgs.ipwatch}/bin/ipwatch"; };
    });
  };
}
