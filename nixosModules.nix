inputs: with inputs;
{
  default = { config, lib, pkgs, utils, ... }:
    let
      cfg = config.services.ipwatch;
      deps = map (iface: "sys-subsystem-net-devices-${utils.escapeSystemdPath iface}.device") cfg.interfaces;
    in
    with lib;
    {
      options.services.ipwatch = {
        enable = mkEnableOption "Enable ipwatch service";
        scripts = mkOption {
          type = types.listOf types.path;
          description = ''
            Scripts to run after receiving a new IP address.
          '';
        };
        interfaces = lib.mkOption {
          type = types.listOf types.str;
          default = [ ];
          description = ''
            Interfaces to listen for changes on.
          '';
        };
      };

      config = mkIf cfg.enable {
        nixpkgs.overlays = [ self.overlays.default ];
        systemd.services.ipwatch = {
          enable = true;
          description = "ipwatch";
          serviceConfig = {
            DynamicUser = "yes";
            Type = "simple";
            ExecStart = "${pkgs.ipwatch}/bin/ipwatch ${lib.concatMapStringsSep " " (iface: "-interface ${iface}") cfg.interfaces} ${lib.concatMapStringsSep " " (script: "-script ${script}") cfg.scripts}";
          };
          wantedBy = [ "multi-user.target" ] ++ deps;
          wants = [ "network.target" ];
          bindsTo = deps;
          after = deps;
          before = [ "network.target" ];
        };
      };
    };
}
