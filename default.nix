{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorSha256 = "sha256-IZXgjGAx2Xyz53vJ/UwzfGyiyD3pMbDeoOamb5qUusI=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
