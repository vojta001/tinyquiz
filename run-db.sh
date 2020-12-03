#! /usr/bin/env nix-shell
#! nix-shell -i bash db.nix

#TODO create the directories if they don't exists
#TODO set the path relative to script location, not relative to CWD
postgres -D .pg-data -k "$PWD/.pg-sockets" #-c listen_addresses='' # Goland does not support connecting over socket
