{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.2";
  src = ./.;
  vendorSha256 = "sha256-A0EH4QMh3odycO91vFFEp0BbWAoN3Tw14JD81ZHH5F8=";
  ldflags = [ "-s" "-w" ];
}
