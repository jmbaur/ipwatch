{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.2";
  src = ./.;
  vendorSha256 = "sha256-/1nrU34QmFKU6e+R88ZnC27NYemPL/+GzzDG8gcsbk4=";
  ldflags = [ "-s" "-w" ];
}
