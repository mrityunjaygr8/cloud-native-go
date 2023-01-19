{ pkgs ? import <nixpkgs> { }  }:
pkgs.mkShell {
  # nativeBuildInputs is usually what you want -- tools you need to run
  nativeBuildInputs = [
    pkgs.go
    pkgs.fish
    pkgs.go-migrate
    pkgs.go-task
    (pkgs.callPackage ./go-jet.nix {})
  ];
}
