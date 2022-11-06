{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.0";
  src = ./.;
  vendorSha256 = "sha256-AwH8pC0S4PSYKW6PacqkO8Hfd1YQ9WKnbFFO/zzy2Ow=";
  ldflags = [ "-s" "-w" ];
}
