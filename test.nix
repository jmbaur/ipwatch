{
  inputs,
  lib,
  testers,
}:

testers.runNixOSTest {
  name = "ipwatch-nixos-test";

  extraBaseModules.imports = [ inputs.self.nixosModules.default ];

  node.pkgs = lib.mkForce null;

  nodes.machine =
    { lib, pkgs, ... }:
    {
      boot.kernelModules = [ "dummy" ];
      services.ipwatch = {
        enable = true;
        hooks = lib.genAttrs [ "dummy0" ] (_: {
          program = pkgs.writeShellScript "on-change" ''
            echo "NEW ADDRESS IS $ADDR"
          '';
          filters = [ "Is4" ];
        });
      };
      networking.useNetworkd = true;
      systemd.network.networks."10-dummy" = {
        name = "dummy0";
        linkConfig.Unmanaged = true;
      };
    };
  testScript = ''
    machine.wait_for_unit("ipwatch.service")

    machine.succeed("ip link add dummy0 type dummy")
    machine.succeed("ip link set dummy0 up")
    machine.succeed("ip addr add 10.0.0.1/24 dev dummy0")
    machine.wait_until_succeeds("journalctl -u ipwatch.service | grep 'NEW ADDRESS IS 10.0.0.1'")
    machine.succeed("ip addr del 10.0.0.1/24 dev dummy0")
    machine.succeed("ip addr add 10.0.0.2/24 dev dummy0")
    machine.wait_until_succeeds("journalctl -u ipwatch.service | grep 'NEW ADDRESS IS 10.0.0.2'")
  '';
}
