{ buildGoModule
, writeShellScriptBin
, ...
}:

let
  drv = buildGoModule {
    pname = "ipwatch";
    version = "0.0.1";
    CGO_ENABLED = 0;
    src = ./.;
    vendorSha256 = "sha256-k5QsE5vrXIStx21onC7E0mMRhqvaZ74twx/IrIkdAgQ=";
    passthru.update = writeShellScriptBin "update" ''
      if [[ $(${drv.go}/bin/go get -u all 2>&1) != "" ]]; then
        sed -i 's/vendorSha256\ =.*;/vendorSha256="sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";/' default.nix
        ${drv.go}/bin/go mod tidy
      fi
    '';
  };
in
drv
