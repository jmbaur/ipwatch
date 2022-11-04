{ buildGoModule, CGO_ENABLED ? 0, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.1.9";
  src = ./.;
  vendorSha256 = "sha256-AwH8pC0S4PSYKW6PacqkO8Hfd1YQ9WKnbFFO/zzy2Ow=";
  ldflags = [ "-s" "-w" ];
  inherit CGO_ENABLED;
}
