{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    steam-run-free # for running non NixOS binaries on NixOS
    ffmpeg_6-full
  ];
}
