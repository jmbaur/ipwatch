{ nixosTest, module, ... }:
nixosTest {
  name = "ipwatch-nixos-test";
  nodes.machine = { config, ... }: {
    imports = [ module ];
    boot.kernelModules = [ "dummy" ];
    services.ipwatch = {
      enable = true;
      extraArgs = [ "-debug" "-4" ];
      interfaces = [
        "dummy0" # manual
        "eth0" # DHCP
      ];
      hooks = [ "internal:echo" ];
    };
    networking.dhcpcd.denyInterfaces = [ "dummy0" ];
  };

  testScript = ''
    start_all()

    machine.succeed("ip link add dummy0 type dummy")

    # ipwatch will start automatically
    machine.wait_for_unit("ipwatch.service")

    machine.succeed("ip addr add 10.0.0.1/24 dev dummy0")
    machine.succeed("ip link set dummy0 up")

    # manual
    machine.succeed("ip addr del 10.0.0.1/24 dev dummy0")
    machine.wait_for_console_text("Deleting address from cache")
    machine.succeed("ip addr add 10.0.0.2/24 dev dummy0")
    machine.wait_for_console_text("Caching new address")
    machine.wait_for_console_text("New IP for 4: 10.0.0.2")

    # dhcp
    machine.succeed("dhcpcd --rebind eth0")
    machine.wait_for_console_text("eth0: rebinding lease")
    machine.wait_for_console_text("New addr was found in cache, skipping hooks")
  '';
}
