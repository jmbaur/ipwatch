{ lib, buildGoModule }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = lib.fileset.toSource {
    root = ./.;
    fileset = lib.fileset.unions [
      ./go.mod
      ./go.sum
      ./cmd
      ./ipwatch
    ];
  };
  vendorHash = "sha256-SG/XqEXKTqmWbfe4H9+yVDp4YKicm2fP+tVV5bCtIpk=";
  ldflags = [
    "-s"
    "-w"
  ];
  CGO_ENABLED = 0;
  meta.mainProgram = "ipwatch";
}
