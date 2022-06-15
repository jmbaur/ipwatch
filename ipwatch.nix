{ buildGo118Module, lib }:
buildGo118Module {
  pname = "ipwatch";
  version = "0.0.1";
  CGO_ENABLED = 0;
  src = builtins.path { path = ./.; };
  vendorSha256 = "sha256-i4m3Jxny1ibC9ul3lqzGt6e6oZQ8IsEr3absXFwNwvs=";
}
