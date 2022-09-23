#!/usr/bin/env bash

{

GH_README="https://github.com/fairDataSociety/fairOS-dfs#how-to-run-fairos-dfs"

dfs_has() {
  type "$1" > /dev/null 2>&1
}

dfs_echo() {
  command printf %s\\n "$*" 2>/dev/null
}

dfs_download() {
  if ! dfs_has "curl"; then
    dfs_echo "Error: you need to have wget installed and in your path. Use brew (mac) or apt (unix) to install curl"
    exit 1
  fi

  if ! dfs_has "wget"; then
    dfs_echo "Error: you need to have wget installed and in your path. Use brew (mac) or apt (unix) to install wget"
    exit 1
  fi

  if [[ "$2" == "latest" ]] ; then
      eval curl -s https://api.github.com/repos/fairDataSociety/fairOS-dfs/releases/latest \
| grep "$1" \
| cut -d : -f 2,3 \
| tr -d \" \
| wget -qi -
  else
      eval curl -s https://api.github.com/repos/fairDataSociety/fairOS-dfs/releases \
| grep "$2/$1" \
| cut -d : -f 2,3 \
| tr -d \" \
| wget -qi -
  fi
}

install_dfs() {
  VERSION="latest"
  if [[ "$1" != "" ]]; then
    VERSION="$1"
  fi

  BIN_NAME="dfs-"

  if [[ "$OSTYPE" == "linux-gnu" ]]; then
    DETECTED_OS="linux" # TODO (Test)
  elif [[ "$OSTYPE" == "darwin"* ]]; then
    DETECTED_OS="mac"
  elif [[ "$OSTYPE" == "cygwin" ]]; then
    DETECTED_OS="linux" # TODO (Test)
  elif [[ "$OSTYPE" == "msys" ]]; then
    DETECTED_OS="windows"
  elif [[ "$OSTYPE" == "win32" ]]; then
    DETECTED_OS="windows" # TODO (Test)
  elif [[ "$OSTYPE" == "freebsd"* ]]; then
    DETECTED_OS="linux" # TODO (Test)
  else
    dfs_echo "Error: unable to detect operating system. Please install manually by referring to $GH_README"
    exit 1
  fi

  dfs_echo VERSION

  ARCH=$(uname -m)

  echo "  /@@@@@@          /@@            /@@@@@@   /@@@@@@                /@@  /@@@@@@"
  echo " /@@__  @@        |__/           /@@__  @@ /@@__  @@              | @@ /@@__  @@"
  echo "| @@  \__//@@@@@@  /@@  /@@@@@@ | @@  \ @@| @@  \__/          /@@@@@@@| @@  \__//@@@@@@@"
  echo "| @@@@   |____  @@| @@ /@@__  @@| @@  | @@|  @@@@@@  /@@@@@@ /@@__  @@| @@@@   /@@_____/"
  echo "| @@_/    /@@@@@@@| @@| @@  \__/| @@  | @@ \____  @@|______/| @@  | @@| @@_/  |  @@@@@@"
  echo "| @@     /@@__  @@| @@| @@      | @@  | @@ /@@  \ @@        | @@  | @@| @@     \____  @@"
  echo "| @@    |  @@@@@@@| @@| @@      |  @@@@@@/|  @@@@@@/        |  @@@@@@@| @@     /@@@@@@@/"
  echo "|__/     \_______/|__/|__/       \______/  \______/          \_______/|__/    |_______/"

  echo "========== FairOs-dfs Installation =========="
  echo "Detected OS: $DETECTED_OS"
  echo "Detected Architecture: $ARCH"
  echo "Downloading Version: $VERSION"
  echo "============================================="

  if [[ "$ARCH" == "arm64" && $DETECTED_OS == "mac" ]]; then
    BIN_NAME="dfs_darwin_arm64"
  elif [[ "$ARCH" == "amd64" && $DETECTED_OS == "mac" ]]; then
    BIN_NAME="dfs_darwin_amd64"
  elif [[ "$ARCH" == "x86_64" && $DETECTED_OS == "windows" ]]; then
    BIN_NAME="dfs_windows_amd64.exe"
  elif [[ "$ARCH" == "arm64" && $DETECTED_OS == "linux" ]]; then
    BIN_NAME="dfs_linux_arm64"
  elif [[ "$ARCH" == "x86_64" && $DETECTED_OS == "linux" ]]; then
    BIN_NAME="dfs_linux_amd64"
  elif [[ "$ARCH" == "amd64" && $DETECTED_OS == "linux" ]]; then
    BIN_NAME="dfs_linux_amd64"
  else
    dfs_echo "Error: unable to detect architecture. Please install manually by referring to $GH_README"
    exit 1
  fi
  dfs_echo "Downloading $BIN_NAME"
  dfs_download $BIN_NAME "$VERSION"
}

install_dfs "$1"

}