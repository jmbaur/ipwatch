{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorHash = "sha256-qwOSYN5zSsE7LU4eSkAZqCViCZAY1y2NbQjGJQEUYgQ=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
