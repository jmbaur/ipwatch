{ buildGo118Module }:
buildGo118Module {
  pname = "ipwatch";
  version = "0.0.1";
  CGO_ENABLED = 0;
  src = builtins.path { path = ./.; };
  vendorSha256 = "sha256-8lgxk2eO60USN6i8v+/2pKHlYnNtoyuCoOrKhbtMpb0=";
}
