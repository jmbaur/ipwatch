{ buildGoModule, go-tools, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.1.6";
  CGO_ENABLED = 0;
  src = ./.;
  vendorSha256 = "sha256-AwH8pC0S4PSYKW6PacqkO8Hfd1YQ9WKnbFFO/zzy2Ow=";
  preCheck = "HOME=/tmp ${go-tools}/bin/staticcheck ./...";
}
