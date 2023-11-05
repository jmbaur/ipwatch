{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorSha256 = "sha256-8O0dIKayTfJ5W5vcdLA8sXgQxUw5hD0Ud8AIzR2mr5E=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
