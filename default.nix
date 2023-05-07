{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorSha256 = "sha256-s9wx0g2nIURiTz4IJbD6VgIwYv/Vigp2RTdy57U0kxU=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
