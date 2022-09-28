{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.0.2";
  CGO_ENABLED = 0;
  src = ./.;
  vendorSha256 = "sha256-gyozyeTSR2XshVwHO9oaRZ5VHD1aE44/VGMXzYWmJWE=";
}
