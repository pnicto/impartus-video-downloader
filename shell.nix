{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    steam-run-free # for running on NixOS binaries on NixOS
    ffmpeg_6-full
  ];
}