{
  description = "Build Golang";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
  flake-utils.lib.eachDefaultSystem (system: 
        let pkgs = import nixpkgs { inherit system; };
      in
      rec {
        devShell = pkgs.mkShell {
          buildInputs = [
            pkgs.go
          ];
        };
        defaultPackage = pkgs.buildGoModule {
          pname = "Hochwasser";
          name = "Hochwasser";
          src = ./.;
          vendorSha256 = "sha256-wc52mzV8cs4X1pQHIDcwh2oZiPLLUJwGN/nU9HJvZpQ=";
        };
        apps.default = { type = "app"; program = "${defaultPackage}/bin/Hochwasser"; };
      });
}
