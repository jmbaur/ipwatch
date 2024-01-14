{ buildGoModule, ... }:
buildGoModule {
  pname = "ipwatch";
  version = "0.2.3";
  src = ./.;
  vendorHash = "sha256-wkan+bfn4mczH6H0nU9qYFAjf+CfXm6TcnJ3Ex8/5o4=";
  ldflags = [ "-s" "-w" ];
  CGO_ENABLED = 0;
}
