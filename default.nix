{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorHash = "sha256-NauVmQ0rQsaYgTStLs/PtR8Tu7O9FohkugCylrCq/Ek=";
  ldflags = [
    "-s"
    "-w"
  ];
  CGO_ENABLED = 0;
}
