{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorSha256 = "sha256-0SGFE9sgYuBWHgITKHR8AwO05R0IKCcpCK0TH3ynmEQ=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
