{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorSha256 = "sha256-mCg6s2ygYiUdGvizvju9xDB6NrH3WAZkVocxlytsNro=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
