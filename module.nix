{
  config,
  lib,
  pkgs,
  utils,
  ...
}:
let
  cfg = config.services.ipwatch;
  deps = map (
    iface: "sys-subsystem-net-devices-${utils.escapeSystemdPath iface}.device"
  ) cfg.interfaces;
in
{
  options.services.ipwatch = with lib; {
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
      type = types.listOf types.str;
      default = [ "internal:echo" ];
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
      default = null;
      description = ''
        File to use to set the environment for the ipwatch systemd service.
        This environment will be propagated to each hook.
      '';
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.ipwatch = {
      enable = true;
      description = "ipwatch (https://github.com/jmbaur/ipwatch)";
      serviceConfig = {
        EnvironmentFile = lib.mkIf (cfg.environmentFile != null) cfg.environmentFile;
        ExecStart = lib.escapeShellArgs (
          [ (lib.getExe cfg.package) ]
          ++ lib.flatten (
            (map (iface: "-interface=${iface}") cfg.interfaces)
            ++ (map (hook: "-hook=${hook}") cfg.hooks)
            ++ (map (filter: "-filter=${filter}") cfg.filters)
          )
          ++ cfg.extraArgs
        );

        CapabilityBoundingSet = [ ];
        DeviceAllow = [ ];
        DynamicUser = true;
        LockPersonality = true;
        MemoryDenyWriteExecute = true;
        NoNewPrivileges = true;
        PrivateDevices = true;
        ProtectClock = true;
        ProtectControlGroups = true;
        ProtectHome = true;
        ProtectHostname = true;
        ProtectKernelLogs = true;
        ProtectKernelModules = true;
        ProtectKernelTunables = true;
        ProtectSystem = "strict";
        RemoveIPC = true;
        RestrictAddressFamilies = [
          "AF_NETLINK"
          "AF_INET"
          "AF_INET6"
        ];
        RestrictNamespaces = true;
        RestrictRealtime = true;
        RestrictSUIDSGID = true;
        SystemCallArchitectures = "native";
      };
      wantedBy = [ "multi-user.target" ];
      before = [ "network-pre.target" ];
      wants = deps;
      after = deps;
    };
  };
}
