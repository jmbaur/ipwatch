{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.0.2";
  CGO_ENABLED = 0;
  src = ./.;
  vendorSha256 = "sha256-AwH8pC0S4PSYKW6PacqkO8Hfd1YQ9WKnbFFO/zzy2Ow=";
}
