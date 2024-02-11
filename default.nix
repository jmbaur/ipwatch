{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorHash = "sha256-CEFT5jRm0Ybf2WeK7PyLGv2M/pofD0KaLPI/uI+Tozs=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
