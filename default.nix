{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorSha256 = "sha256-nFjE1VCGs6B4khNFZcNkz+5Zfn8SGL3wO9Qfh4yeHuM=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
