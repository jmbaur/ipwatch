{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorSha256 = "";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
