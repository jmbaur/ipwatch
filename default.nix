{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.0.2";
  CGO_ENABLED = 0;
  src = ./.;
  vendorSha256 = "sha256-upILiQkwVyg0xgpEjVMc71w8FZrOOWxwXO7Vp9R5rDM=";
}
