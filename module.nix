{ config, lib, pkgs, utils, ... }:
let
  cfg = config.services.ipwatch;
  deps = map (iface: "sys-subsystem-net-devices-${utils.escapeSystemdPath iface}.device") cfg.interfaces;
in
with lib;
{
  options.services.ipwatch = {
    enable = mkEnableOption "Enable ipwatch service";
    package = mkPackageOption pkgs "ipwatch" { };
    extraArgs = mkOption {
      type = types.listOf types.str;
      default = [ ];
      description = ''
        Extra arguments to be passed to ipwatch.
      '';
    };
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
    environmentFile = lib.mkOption {
      type = types.nullOr types.path;
      description = ''
        File to use to set the environment for scripts.
      '';
    };
  };

  config = mkIf cfg.enable {
    systemd.services.ipwatch = {
      enable = true;
      description = "ipwatch";
      serviceConfig = {
        DynamicUser = true;
        ProtectHome = true;
        ProtectSystem = true;
        EnvironmentFile = mkIf (cfg.environmentFile != null) cfg.environmentFile;
        ExecStart = lib.escapeShellArgs ([ "${cfg.package}/bin/ipwatch" ] ++
          lib.flatten (
            (map (iface: "-interface=${iface}") cfg.interfaces) ++
              (map (script: "-script=${script}") cfg.scripts)
          ) ++ cfg.extraArgs
        );

      };
      wantedBy = [ "multi-user.target" ] ++ deps;
      wants = [ "network.target" ];
      bindsTo = deps;
      after = deps;
      before = [ "network.target" ];
    };
  };
}
