{
  description = "Run code on changes to network interfaces";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = inputs: with inputs; {
    nixosModules.default = import ./module.nix;
    overlays.default = final: prev: {
      ipwatch = prev.callPackage ./ipwatch.nix { };
    };
  } //
  flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        inherit system;
        overlays = [ self.overlays.default ];
      };
    in
    {
      devShells.default = pkgs.mkShell {
        CGO_ENABLED = 0;
        buildInputs = with pkgs; [ go-tools go ];
      };
      packages.default = pkgs.ipwatch;
      apps.default = flake-utils.lib.mkApp {
        drv = pkgs.ipwatch;
        name = "ipwatch";
      };
    });
}
