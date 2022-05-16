{ config, lib, pkgs, utils, ... }:
let
  cfg = config.services.ipwatch;
  interfacesFlag = lib.concatSringsSep "," cfg.interfaces;
  deps = map (iface: "sys-subsystem-net-devices-${utils.escapeSystemdPath iface}.device") cfg.interfaces;
in
with lib;
{
  options.services.ipwatch = {
    enable = mkEnableOption "Enable ipwatch service";
    hookScript = mkOption {
      type = types.path;
      description = ''
        The path to an executable/script to run after receiving a new IP
        address.
      '';
    };
    interfaces = lib.mkOption {
      type = types.listOf types.str;
      default = [ ];
      description = ''
        The interfaces to listen for changes on.
      '';
    };
  };

  config = mkIf cfg.enable {
    systemd.services.ipwatch = {
      enable = true;
      description = "ipwatch";
      serviceConfig = {
        DynamicUser = "yes";
        Type = "simple";
        ExecStart = "${pkgs.ipwatch}/bin/ipwatch -hook-script ${cfg.hookScript} -interfaces ${interfacesFlag}";
      };
      wantedBy = [ "multi-user.target" ] ++ deps;
      wants = [ "network.target" ];
      bindsTo = deps;
      after = deps;
      before = [ "network.target" ];
    };
  };
}
