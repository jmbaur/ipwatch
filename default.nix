{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorSha256 = "sha256-sShCqSee3VtC6D76IgQTgTF+T+q5SScgvQ6uq3vnlbw=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
