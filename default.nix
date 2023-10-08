{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorSha256 = "sha256-H/rQkrZRH/zHNt69e1k1Y9PcvXUVUy6JPwPY4NYWsSY=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
