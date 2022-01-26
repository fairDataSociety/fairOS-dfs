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

  eval curl -s https://api.github.com/repos/fairDataSociety/fairOS-dfs/releases/latest \
| grep "$1" \
| cut -d : -f 2,3 \
| tr -d \" \
| wget -qi -

}

install_dfs() {
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
  echo "====================================================="

  if [[ "$ARCH" == "arm64" && $DETECTED_OS == "mac" ]]; then
    BIN_NAME="dfs-darwin-amd64"
    dfs_echo $BIN_NAME
  elif [[ "$ARCH" == "x86_64" && $DETECTED_OS == "windows" ]]; then
    BIN_NAME="dfs-windows-amd64.exe"
    dfs_echo $BIN_NAME
  elif [[ "$ARCH" == "x86_32" && $DETECTED_OS == "windows" ]]; then
    BIN_NAME="dfs-windows-386.exe"
    dfs_echo $BIN_NAME
  elif [[ "$ARCH" == "arm64" && $DETECTED_OS == "linux" ]]; then
    BIN_NAME="dfs-linux-arm64.exe"
    dfs_echo $BIN_NAME
  elif [[ "$ARCH" == "x86_32" && $DETECTED_OS == "linux" ]]; then
    BIN_NAME="dfs-linux-386.exe"
    dfs_echo $BIN_NAME
  elif [[ "$ARCH" == "x86_64" && $DETECTED_OS == "linux" ]]; then
    BIN_NAME="dfs-linux-amd64.exe"
    dfs_echo $BIN_NAME
  elif [[ "$ARCH" == "amd64" && $DETECTED_OS == "linux" ]]; then
    BIN_NAME="dfs-linux-amd64.exe"
    dfs_echo $BIN_NAME
  else
    dfs_echo "Error: unable to detect architecture. Please install manually by referring to $GH_README"
    exit 1
  fi

  dfs_download $BIN_NAME
}

install_dfs

}