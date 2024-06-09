{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorHash = "sha256-fC9pv2iStQeWoLaojGk4QNKgtESeOfz+SMVRCm5QjA0=";
  ldflags = [
    "-s"
    "-w"
  ];
  CGO_ENABLED = 0;
}
