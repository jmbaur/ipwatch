{
  description = "Run code on changes to network interfaces";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = inputs: with inputs; {
    overlays = import ./overlays.nix inputs;
    nixosModules = import ./nixosModules.nix inputs;
  } //
  flake-utils.lib.eachSystem [ "x86_64-linux" "aarch64-linux" ] (system:
    let pkgs = import nixpkgs { inherit system; overlays = [ self.overlays.default ]; }; in
    {
      packages.default = pkgs.ipwatch;
      apps.default = flake-utils.lib.mkApp { drv = pkgs.ipwatch; name = "ipwatch"; };
    }) //
  flake-utils.lib.eachDefaultSystem (system:
    let pkgs = import nixpkgs { inherit system; }; in
    {
      devShells.default = pkgs.mkShell {
        CGO_ENABLED = 0;
        buildInputs = with pkgs; [ go-tools go ];
      };
    });
}
