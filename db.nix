with
  import (builtins.fetchTarball {
    url = "https://github.com/nixos/nixpkgs/archive/4d373182597cff60b3a820affb5a73dd274e205b.tar.gz";
    sha256 = "1kvsnlsq1czgk085rllg1b427n6lv5w3a2z1igfc95n81jpvcf58";
  }) {};
let
in
  stdenv.mkDerivation rec {
    name = "tinyquiz-db-env";

    buildInputs = [ postgresql ];
  }
