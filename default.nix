{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorHash = "sha256-nJuL2rNUCWYebVCLx7nbZ+I27QkiUrmCm7BZ7et6x8o=";
  ldflags = [
    "-s"
    "-w"
  ];
  CGO_ENABLED = 0;
}
