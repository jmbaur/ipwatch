{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorSha256 = "sha256-FN9CvNWawnb4bP1jz0jjlCc7LwCfznQGczEt0Y1Rl5g=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
