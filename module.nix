{ config, lib, pkgs, ... }:
let
  cfg = config.services.ipwatch;
in
{
  options.services.ipwatch = {
    enable = lib.mkEnableOption "Enable ipwatch service";
    exe = lib.mkOption {
      type = lib.types.path;
      description = ''
        The path to an executable to run
      '';
    };
    iface = lib.mkOption {
      type = lib.types.str;
      default = "";
      description = ''
        The interface to listen for changes on
      '';
    };
  };

  config = lib.mkIf cfg.enable {
    users.users.ipwatch = {
      isSystemUser = true;
      group = "ipwatch";
    };
    users.groups.ipwatch = { };

    systemd.services.ipwatch = {
      enable = true;
      description = "ipwatch";
      serviceConfig = {
        Type = "simple";
        User = users.users.ipwatch.name;
        Group = users.groups.ipwatch.name;
        ExecStart = "${pkgs.ipwatch}/bin/ipwatch -exe ${cfg.exe}${lib.optionalString (cfg.iface != "") " -iface ${cfg.iface}"}";
      };
      bindsTo = lib.mkIf (cfg.iface != "") [ "sys-subsystem-net-devices-${cfg.iface}.device" ];
      wantedBy = [ "multi-user.target" ];
    };
  };
}
