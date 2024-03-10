{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorHash = "sha256-w+VTfCBgnbZ8YzYnp2AnQihU8M5OiO/4v0ZVX9UZm8s=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
