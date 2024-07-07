{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorHash = "sha256-rtVXJTfSZh+TZwPX59W0TtD51GNLMoghIboyDIdoJOw=";
  ldflags = [
    "-s"
    "-w"
  ];
  CGO_ENABLED = 0;
}
