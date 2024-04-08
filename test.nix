{ nixosTest, module, ... }:
nixosTest {
  name = "ipwatch-nixos-test";
  nodes.machine =
    { ... }:
    {
      imports = [ module ];
      boot.kernelModules = [ "dummy" ];
      services.ipwatch = {
        enable = true;
        extraArgs = [
          "-debug"
          "-4"
        ];
        interfaces = [
          "dummy0" # manual
          "eth0" # DHCP
        ];
        hooks = [ "internal:echo" ];
      };
      networking.dhcpcd.denyInterfaces = [ "dummy0" ];
    };
  testScript = builtins.readFile ./test.py;
}
