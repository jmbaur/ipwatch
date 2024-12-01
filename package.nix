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
  vendorHash = "sha256-9GHmJc0xBaSwG1tMuVKGK8bETRrWEuoenS145QAV85U=";
  ldflags = [
    "-s"
    "-w"
  ];
  CGO_ENABLED = 0;
  meta.mainProgram = "ipwatch";
}
