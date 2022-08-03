{ buildGoModule, lib }:
buildGoModule {
  pname = "ipwatch";
  version = "0.0.1";
  CGO_ENABLED = 0;
  src = ./.;
  vendorSha256 = "sha256-6jHBEG7GZ8GXGER0P+sTv3AMcm+RsILpfpDJo64hMpg=";
}
