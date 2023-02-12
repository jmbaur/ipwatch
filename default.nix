{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.2";
  src = ./.;
  vendorSha256 = "sha256-9owX2rJZxfUWOS5Pt1fWlEH/i74754bKKyBrTmURFT0=";
  ldflags = [ "-s" "-w" ];
}
