{
  description = "Tinyquiz – an open source online quiz platform";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
  let
    pkgs = nixpkgs.legacyPackages.x86_64-linux;
    buildCmd = name:
      (pkgs.buildGoPackage {
        pname = "tinyquiz-${name}";
        version = "dev";
        src = ./.;
        goDeps = ./deps.nix;
        preBuild = "go generate vkane.cz/tinyquiz/...";
        checkPhase = "go test -race vkane.cz/tinyquiz/...";
        doCheck = true;
        goPackagePath = "vkane.cz/tinyquiz";
        subPackages = [ "cmd/${name}" ];
      });
  in
    {
      defaultPackage.x86_64-linux = self.packages.x86_64-linux.tinyquiz-web;
      packages.x86_64-linux.tinyquiz-web = buildCmd "web";
      packages.x86_64-linux.dev = pkgs.writeShellScriptBin "dev" ''
        echo "This dev script must be run from the project root, otherwise unexpected behavior might occur."
        read -p "Are you in the right directory and shall I continue? (y/n): " ack

        if [ "$ack" != y ]; then exit 1; fi

        unset GOROOT # Use the one bundled into the binary. I don't currently know who sets this to the wrong one

        '${pkgs.findutils}/bin/find' cmd pkg ui | '${pkgs.entr}/bin/entr' -dr '${pkgs.go}/bin/go' run ./cmd/web
      '';
      packages.x86_64-linux.devDb = pkgs.writeShellScriptBin "devDb" ''
        '${pkgs.postgresql}/bin/postgres' -D .pg-data -k "$PWD/.pg-sockets" -c log_statement=all #-c listen_addresses="" # Goland does not support connecting over socket
      '';
      nixosModules.web = { config, lib, pkgs, ... }:
      let
        cfg = config.services.tinyquiz;
      in
        {
          options = {
            services.tinyquiz = {
              enable = lib.mkEnableOption "tinyquiz";

              postgresql.enable = lib.mkOption {
                type = lib.types.bool;
                default = true;
                description = ''
                  Whether the set up the postgresql database for tinyquiz automatically.
                '';
              };
              config = lib.mkOption {
                type = lib.types.attrsOf lib.types.str;
              };
            };
          };
          config = lib.mkIf cfg.enable {
            users.groups.tinyquiz = {};
            users.users.tinyquiz = {
              description = "Tinyquiz service user";
              group = "tinyquiz";
              isSystemUser = true;
            };
            systemd.services.tinyquiz = {
              description = "Tinyquiz service";
              wantedBy = [ "multi-user.target" ];
              after = [ "network.target" (lib.mkIf cfg.postgresql.enable "postgresql.service") ];

              serviceConfig = {
                ExecStart = "${self.packages.x86_64-linux.tinyquiz-web}/bin/web";
                User = "tinyquiz";
              };
              environment = cfg.config;
            };
            services.tinyquiz.config = if cfg.postgresql.enable then {
              TINYQUIZ_PG_HOST = "/run/postgresql";
            } else {};
            services.postgresql = lib.mkIf cfg.postgresql.enable {
              enable = true;
              ensureDatabases = [ "tinyquiz" ];
              ensureUsers = [
                {
                  name = "tinyquiz";
                  ensurePermissions = {
                    "DATABASE \"tinyquiz\"" = "ALL PRIVILEGES";
                  };
                }
              ];
            };
          };
        };
      checks.x86_64-linux.module = (import "${nixpkgs}/nixos/tests/make-test-python.nix" ({ pkgs, ... }: {
        system = "x86_64-linux";
        machine = { config, pkgs, ... }:
        {
          require = [
            self.nixosModules.web
          ];
          services.tinyquiz = {
            enable = true;
            config = {
              TINYQUIZ_LISTEN = ":8080";
            };
          };
          systemd.services.postgresql.serviceConfig.TimeoutSec = pkgs.lib.mkForce "3600"; # wait for postgresql even on very slow machines. Not using "infinity" is just a safeguard
        };
        testScript = ''
          machine.start()
          machine.wait_for_unit("default.target")
          machine.wait_for_open_port(8080)
          machine.succeed("curl --fail http://localhost:8080")
        '';
      })) { system = "x86_64-linux"; inherit pkgs; };
    };
}
