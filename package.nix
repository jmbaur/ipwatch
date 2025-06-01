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
  vendorHash = "sha256-mlqf9WRoNNOBX2mAvLH/gr9Vi1kA6JTu2YRYN4CwM5E=";
  ldflags = [
    "-s"
    "-w"
  ];
  env.CGO_ENABLED = 0;
  meta.mainProgram = "ipwatch";
}
