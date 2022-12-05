{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.2";
  src = ./.;
  vendorSha256 = "sha256-ZSHpGx6PG7DgFxB6pndGGAF2ysl8OuybhEO5z1ky7Ck=";
  ldflags = [ "-s" "-w" ];
}
