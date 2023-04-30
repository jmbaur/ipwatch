{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorSha256 = "sha256-8+cmreIfjMET+v6sttEDfYtlqdXcevcJFFNlLdr7o6c=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
