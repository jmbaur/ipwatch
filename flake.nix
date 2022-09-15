{
  description = "Run code on changes to network interfaces";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = inputs: with inputs; {
    overlays.default = final: prev: { ipwatch = prev.callPackage ./. { }; };
    nixosModules = import ./nixosModules.nix inputs;
  } // flake-utils.lib.eachSystem [ "x86_64-linux" "aarch64-linux" ] (system:
    let
      pkgs = import nixpkgs {
        inherit system;
        overlays = [ self.overlays.default ];
      };
    in
    {
      devShells.default = pkgs.mkShell {
        inherit (pkgs.ipwatch)
          CGO_ENABLED
          nativeBuildInputs;
      };
      packages.default = pkgs.ipwatch;
      apps.default = { type = "app"; program = "${pkgs.ipwatch}/bin/ipwatch"; };
    });
}
