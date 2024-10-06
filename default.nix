{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorHash = "sha256-VG1ZVwO78KUqyQpQw8hbjJm6AxvudoiNQbjYevd+qj8=";
  ldflags = [
    "-s"
    "-w"
  ];
  CGO_ENABLED = 0;
}
