{
  description = "Tinyquiz – an open source online quiz platform";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-20.09";

  outputs = { self, nixpkgs }:
  let
    pkgs = nixpkgs.legacyPackages.x86_64-linux;
  in
    {
      defaultPackage.x86_64-linux = self.packages.x86_64-linux.tinyquiz;
      packages.x86_64-linux.tinyquiz = pkgs.hello; # TODO
      packages.x86_64-linux.dev = pkgs.writeShellScriptBin "dev" ''
        echo "This dev script must be run from the project root, otherwise unexpected behavior might occur."
        read -p "Are you in the right directory and shall I continue? (y/n): " ack

        if [ "$ack" != y ]; then exit 1; fi

        unset GOROOT # Use the one bundled into the binary. I don't currently know who sets this to the wrong one

        '${pkgs.findutils}/bin/find' cmd pkg ui/html | '${pkgs.entr}/bin/entr' -dr '${pkgs.go_1_14}/bin/go' run ./cmd/web
      '';
      packages.x86_64-linux.devDb = pkgs.writeShellScriptBin "devDb" ''
        '${pkgs.postgresql}/bin/postgres' -D .pg-data -k "$PWD/.pg-sockets" #-c listen_addresses="" # Goland does not support connecting over socket
      '';
    };
}
