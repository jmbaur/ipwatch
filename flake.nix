{
  description = "Run code on changes to network interfaces";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "nixpkgs/nixos-unstable";
    pre-commit-hooks.inputs.nixpkgs.follows = "nixpkgs";
    pre-commit-hooks.url = "github:cachix/pre-commit-hooks.nix";
  };

  outputs = inputs: with inputs; {
    overlays.default = _: prev: { ipwatch = prev.callPackage ./. { }; };
    nixosModules.default = {
      nixpkgs.overlays = [ self.overlays.default ];
      imports = [ ./module.nix ];
    };
  } // flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        inherit system;
        overlays = [ self.overlays.default ];
      };
      preCommitCheck = pre-commit-hooks.lib.${system}.run {
        src = ./.;
        hooks = {
          nixpkgs-fmt.enable = true;
          govet.enable = true;
          gofmt = {
            enable = true;
            entry = "${pkgs.ipwatch.go}/bin/gofmt -w";
            types = [ "go" ];
          };
        };
      };
    in
    {
      devShells.default = pkgs.mkShell {
        buildInputs = with pkgs; [ just go-tools ];
        inherit (preCommitCheck) shellHook;
        inherit (pkgs.ipwatch)
          CGO_ENABLED
          nativeBuildInputs;
      };
      packages.default = pkgs.ipwatch;
      packages.test = pkgs.callPackage ./test.nix { inherit inputs; };
      apps.default = { type = "app"; program = "${pkgs.ipwatch}/bin/ipwatch"; };
    });
}
