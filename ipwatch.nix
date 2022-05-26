{ buildGo118Module }:
buildGo118Module {
  pname = "ipwatch";
  version = "0.0.1";
  CGO_ENABLED = 0;
  src = builtins.path { path = ./.; };
  vendorSha256 = "sha256-gR2BwrHd7UcvaKIOD3LtqMvFSYxbxUq3LXm0IiHDFr8=";
}
