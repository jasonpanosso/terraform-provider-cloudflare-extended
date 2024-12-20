{
  description = "devshell flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];

      forEachSupportedSystem = f: nixpkgs.lib.genAttrs supportedSystems (system: f {
        pkgs = import nixpkgs { inherit system; config.allowUnfree = true; };
      });
    in
    {
      devShells = forEachSupportedSystem
        ({ pkgs }: {
          default = pkgs.mkShell {
            packages = [
              pkgs.terraform
              pkgs.go
              pkgs.gnumake
              pkgs.tfplugindocs
              pkgs.goreleaser
            ];
          };
        });
    };
}
