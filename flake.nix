{
  description = "Run code on changes to network interfaces";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
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
    in
    {
      devShells.default = pkgs.mkShell {
        buildInputs = [ pkgs.just ];
        inherit (pkgs.ipwatch)
          CGO_ENABLED
          nativeBuildInputs;
      };
      packages.default = pkgs.ipwatch;
      packages.test = pkgs.callPackage ./test.nix { inherit inputs; };
      apps.default = { type = "app"; program = "${pkgs.ipwatch}/bin/ipwatch"; };
    });
}
