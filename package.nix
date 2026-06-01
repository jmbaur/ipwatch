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
  vendorHash = "sha256-nHi9Ec8pdEMBMM0GtCKh2g/Z3x46aklgwUHHs2EIgts=";
  ldflags = [
    "-s"
    "-w"
  ];
  env.CGO_ENABLED = 0;
  meta.mainProgram = "ipwatch";
}
