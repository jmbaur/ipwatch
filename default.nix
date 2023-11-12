{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorSha256 = "sha256-y/8VrSnlVdAg56JxX/MhSxg9KeM0tIATJ+cQDzI/P1w=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
