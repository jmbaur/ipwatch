{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorHash = "sha256-c8f9xiO36qwlruTFRSLZbj7WD88GjMZfOdpaSC1AJwY=";
  ldflags = [
    "-s"
    "-w"
  ];
  CGO_ENABLED = 0;
}
