{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorHash = "sha256-jISUHi4DdP6WtToL01jkhkPlNtcC4EfNWB86hSzrBV8=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
