_: {
  default = _: super: {
    ipwatch = super.buildGoModule {
      pname = "ipwatch";
      version = "0.0.1";
      CGO_ENABLED = 0;
      src = ./.;
      vendorSha256 = "sha256-ySpXrn/zrIleW5Mkuw+Q8kubiM8+erb6UHhA83w4wyw=";
    };
  };

}
