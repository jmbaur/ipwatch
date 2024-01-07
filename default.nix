{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorHash = "sha256-MZCR6nfINbH/79xox/cwTypQUmUXEQYHNKwIbMYhUH8=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
