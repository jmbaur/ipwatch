{
  config,
  lib,
  pkgs,
  ...
}:
let
  cfg = config.services.ipwatch;
in
{
  options.services.ipwatch = with lib; {
    enable = mkEnableOption "Enable ipwatch service";
    package = mkPackageOption pkgs "ipwatch" { };
    hooks = mkOption {
      type = types.attrsOf (
        types.submodule (
          { name, ... }:
          {
            options = {
              interface = mkOption {
                type = types.str;
                default = name;
                description = ''
                  Interface to listen for changes on.
                '';
              };
              filters = mkOption {
                type = types.listOf types.str;
                default = [ ];
                description = ''
                  Filters to apply on new IP addresses that will conditionally
                  run hooks.
                '';
              };
              program = mkOption {
                type = types.path;
                description = ''
                  Program to run when filter passes for changes to ''${interface}.
                '';
              };
            };
          }
        )
      );
      default = { };
      description = ''
        Hooks to run after receiving a new IP address.
      '';
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.ipwatch = {
      enable = true;
      description = "ipwatch (https://github.com/jmbaur/ipwatch)";
      before = [ "network-pre.target" ];
      wantedBy = [ "multi-user.target" ];
      serviceConfig = {
        Type = "notify";
        ExecStart = lib.escapeShellArgs (
          [ (lib.getExe cfg.package) ]
          ++ lib.flatten (
            map (hook: "-hook=${hook.interface}:${lib.concatStringsSep "," hook.filters}:${hook.program}") (
              lib.attrValues cfg.hooks
            )
          )
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
          "AF_UNIX" # needed for notify support
        ];
        RestrictNamespaces = true;
        RestrictRealtime = true;
        RestrictSUIDSGID = true;
        SystemCallArchitectures = "native";
      };
    };
  };
}
