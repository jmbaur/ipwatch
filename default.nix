{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorHash = "sha256-CRsFyl9ZYZ+/6DzjBQJW9jzflGrAyd+k9RsaWcNsCdI=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
