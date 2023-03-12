{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorSha256 = "sha256-B6u8kX2w6I8vWgWj4UkpA76RiLHBDaGaBBalkjKnzig=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
