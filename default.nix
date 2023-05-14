{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorSha256 = "sha256-D67uCv2DCGRHz7qxOvSmpz0re63BoBNS28G7kOUlXPs=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
