{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorHash = "sha256-Amh1uf4xziSErgk9psW0LRM3le1tT+5PTcKSABkixv4=";
  ldflags = [
    "-s"
    "-w"
  ];
  CGO_ENABLED = 0;
}
