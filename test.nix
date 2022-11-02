{ nixosTest, module, ... }:
nixosTest {
  name = "ipwatch-nixos-test";
  nodes.machine = { config, ... }: {
    imports = [ module ];
    boot.kernelModules = [ "dummy" ];
    services.ipwatch = {
      enable = true;
      extraArgs = [ "-debug" ];
      interfaces = [ "dummy0" ];
      hooks = [ "internal:echo" ];
    };
  };

  testScript = ''
    machine.succeed("ip link add dummy0 type dummy")
    machine.succeed("ip addr add 10.0.0.1/24 dev dummy0")
    machine.succeed("ip link set dummy0 up")
    machine.wait_for_unit("ipwatch.service")
    machine.succeed("ip addr del 10.0.0.1/24 dev dummy0")
    machine.wait_for_console_text("Deleting address from cache")
    machine.succeed("ip addr add 10.0.0.2/24 dev dummy0")
    machine.wait_for_console_text("Caching new address")
    machine.wait_for_console_text("New IP for [0-9]: 10.0.0.2") # match on any interface index
  '';
}
