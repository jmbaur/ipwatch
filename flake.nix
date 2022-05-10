{
  description = "Run code on changes to network interfaces";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }: {
    nixosModules.default = import ./module.nix;
    overlays.default = final: prev: {
      ipwatch = prev.callPackage ./ipwatch.nix { };
    };
  } //
  flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        overlays = [ self.overlays.default ];
        inherit system;
      };
    in
    {
      devShells.default = pkgs.mkShell {
        CGO_ENABLED = 0;
        buildInputs = with pkgs; [ fd entr go_1_18 ];
      };
      packages.default = pkgs.ipwatch;
      apps.default = flake-utils.lib.mkApp {
        drv = pkgs.ipwatch;
        name = "ipwatch";
      };
    });
}
