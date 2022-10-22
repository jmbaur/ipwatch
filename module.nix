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
    hooks = mkOption {
      type = types.listOf types.string;
      description = ''
        Hooks to run after receiving a new IP address.
      '';
    };
    interfaces = lib.mkOption {
      type = types.listOf types.str;
      default = [ ];
      description = ''
        Interfaces to listen for changes on.
      '';
    };
    filters = lib.mkOption {
      type = types.listOf types.str;
      default = [ ];
      description = ''
        Filters to apply on new IP addresses that will conditionally run hooks.
      '';
    };
    environmentFile = lib.mkOption {
      type = types.nullOr types.path;
      description = ''
        File to use to set the environment for hooks that need it.
      '';
    };
  };

  config = mkIf cfg.enable {
    systemd.services.ipwatch = {
      enable = true;
      description = "ipwatch (https://github.com/jmbaur/ipwatch)";
      serviceConfig = {
        DynamicUser = true;
        ProtectHome = true;
        ProtectSystem = true;
        EnvironmentFile = mkIf (cfg.environmentFile != null) cfg.environmentFile;
        ExecStart = lib.escapeShellArgs ([ "${cfg.package}/bin/ipwatch" ] ++
          lib.flatten (
            (map (iface: "-interface=${iface}") cfg.interfaces) ++
              (map (hook: "-hook=${hook}") cfg.hooks) ++
              (map (filter: "-filter=${filter}") cfg.filters)
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
