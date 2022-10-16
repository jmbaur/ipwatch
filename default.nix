{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.0.2";
  CGO_ENABLED = 0;
  src = ./.;
  vendorSha256 = "sha256-3c1mgws2KhFSHSbgpn+QmEOCt5aGgUklQtR5xTgWToE=";
}
