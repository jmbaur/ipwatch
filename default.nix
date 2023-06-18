{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorSha256 = "sha256-fPFINgKVYK56ex/LcRRbzJYI7SZuU3CXlVUJSVLwlsg=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
