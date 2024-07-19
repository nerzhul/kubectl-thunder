{ pkgs ? import <nixpkgs> { }, }:

pkgs.mkShell {
  LOCALE_ARCHIVE = "${pkgs.glibcLocales}/lib/locale/locale-archive";
  env.LANG = "C.UTF-8";
  env.LC_ALL = "C.UTF-8";

  packages = [
    pkgs.go
    pkgs.delve
    pkgs.git
    pkgs.kubectl
    pkgs.nixfmt
    pkgs.bash-completion
  ];
}
