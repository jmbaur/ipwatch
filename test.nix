{ nixosTest, module, ... }:
nixosTest {
  name = "ipwatch-nixos-test";
  nodes.machine =
    { lib, pkgs, ... }:
    {
      imports = [ module ];
      boot.kernelModules = [ "dummy" ];
      services.ipwatch = {
        enable = true;
        hooks = lib.genAttrs [ "dummy0" ] (_: {
          program = pkgs.writeShellScript "on-change" ''
            set -x
            if [[ $ADDR == "10.0.0.1" ]]; then
              touch /tmp/1
            elif [[ $ADDR == "10.0.0.2" ]]; then
              touch /tmp/2
            else
              touch /tmp/fail
            fi
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
    machine.wait_until_succeeds("test -e /tmp/systemd*ipwatch*/tmp/1")
    machine.succeed("ip addr del 10.0.0.1/24 dev dummy0")
    machine.succeed("ip addr add 10.0.0.2/24 dev dummy0")
    machine.wait_until_succeeds("test -e /tmp/systemd*ipwatch*/tmp/2")

    machine.succeed("! test -e /tmp/systemd*ipwatch*/tmp/fail")
  '';
}
