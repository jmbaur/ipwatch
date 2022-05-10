{ buildGo118Module }:
buildGo118Module {
  pname = "ipwatch";
  version = "0.0.1";
  src = builtins.path { path = ./.; };
  vendorSha256 = "sha256-I5RnpjSYFGfWu4PU6P6GYbuMI/A2/QjjXzVjsY3/4a8=";
}
