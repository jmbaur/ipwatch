{ testers }:

testers.runNixOSTest {
  name = "ipwatch-nixos-test";

  nodes.machine =
    { lib, pkgs, ... }:
    {
      boot.kernelModules = [ "dummy" ];
      systemd.services.test-ipwatch = {
        wantedBy = [ "multi-user.target" ];
        path = [
          pkgs.jq
          pkgs.ipwatch
        ];
        script = ''
          ipwatch -hook dummy0:Is4 | while read -r json_line; do
            printf "NEW ADDRESS IS %s\n" $(echo "$json_line" | jq -r '.address')
          done
        '';
      };
      networking.useNetworkd = true;
      systemd.network.networks."10-dummy" = {
        name = "dummy0";
        linkConfig.Unmanaged = true;
      };
    };
  testScript = ''
    machine.wait_for_unit("test-ipwatch.service")

    machine.succeed("ip link add dummy0 type dummy")
    machine.succeed("ip link set dummy0 up")
    machine.succeed("ip addr add 10.0.0.1/24 dev dummy0")
    machine.wait_until_succeeds("journalctl -u test-ipwatch.service | grep 'NEW ADDRESS IS 10.0.0.1'")
    machine.succeed("ip addr del 10.0.0.1/24 dev dummy0")
    machine.succeed("ip addr add 10.0.0.2/24 dev dummy0")
    machine.wait_until_succeeds("journalctl -u test-ipwatch.service | grep 'NEW ADDRESS IS 10.0.0.2'")
  '';
}
