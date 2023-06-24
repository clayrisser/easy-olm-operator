{ nixpkgs ? import <nixpkgs> {} }:

nixpkgs.stdenv.mkDerivation rec {
  name = "packages";
  buildInputs = [
    nixpkgs.cloc
    nixpkgs.docker-compose
    nixpkgs.gnumake42
    nixpkgs.gnused
    nixpkgs.jq
    nixpkgs.operator-sdk
    nixpkgs.yq-go
  ];
}
