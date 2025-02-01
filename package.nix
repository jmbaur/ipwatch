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
  vendorHash = "sha256-G5KL6w+U7ZJ16SD675cnX6K4UQAzCo5CQK5JwawpKLM=";
  ldflags = [
    "-s"
    "-w"
  ];
  env.CGO_ENABLED = 0;
  meta.mainProgram = "ipwatch";
}
