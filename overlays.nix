_: {
  default = _: super: {
    ipwatch = super.buildGoModule {
      pname = "ipwatch";
      version = "0.0.1";
      CGO_ENABLED = 0;
      src = ./.;
      vendorSha256 = "sha256-hhcxhhdKwVZH/VIK9zHmrPaEZ1XpeNkQdWjxEvyA8ZQ=";
    };
  };

}
