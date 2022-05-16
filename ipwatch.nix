{ buildGo118Module, lib }:
buildGo118Module {
  pname = "ipwatch";
  version = "0.0.1";
  CGO_ENABLED = 0;
  src = builtins.path { path = ./.; };
  vendorSha256 = "sha256-U+faweP3KX06n50IDtz2YP9aKuzivX5KfXhcXJujzOU=";
}
