{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorSha256 = "sha256-ETm0XlwkH1fuLYe2qt0XoAirql6nmdxokxmCllOupz0=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
