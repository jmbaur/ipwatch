{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorSha256 = "sha256-3Gq2m6eRdtb3ouV3jBsTbm7JVygvJEfb6+UWgMWxOyI=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
