{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.2";
  src = ./.;
  vendorSha256 = "sha256-J9BWF6mPyC6LvuCWYlwqyspbNUWkmQ3BeVQzSzAYkRM=";
  ldflags = [ "-s" "-w" ];
}
