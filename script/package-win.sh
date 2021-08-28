#!/bin/bash

set -eux

APP_NAME="image-sorter"
EXE_NAME="${APP_NAME}.exe"
VERSION="$(cat VERSION)-$(git rev-parse --short HEAD)"

BIN_SRC_DIR="out/windows"
MSYS_DIR="/c/msys64/mingw64"
DEPLOY_DIR_WIN="./deploy/windows"
DEPLOY_DIR_WIN_BIN="${DEPLOY_DIR_WIN}"
ZIP_BASE_DIR="./deploy/windows"
ZIP_FILE_NAME="${APP_NAME}-${VERSION}.zip"
EXE_PATH="${DEPLOY_DIR_WIN_BIN}/image-sorter.exe"

# Copy binary the
mkdir -p "${DEPLOY_DIR_WIN_BIN}"
cp "${BIN_SRC_DIR}/${EXE_NAME}" ${DEPLOY_DIR_WIN_BIN}/${EXE_NAME}

# Copy DLLs for Windows
echo "Copy dependencies:"

ldd ${EXE_PATH} | sed -n 's/\([^ ]*\) => \(\/.*msys64.*\.dll\).*/\2/p' | sort | xargs cp -t ${DEPLOY_DIR_WIN_BIN}
ldd ${EXE_PATH} | sed -n 's/\([^ ]*\) => \(\/.*mingw64.*\.dll\).*/\2/p' | sort | xargs cp -t ${DEPLOY_DIR_WIN_BIN}

cd ${ZIP_BASE_DIR}
echo "Create zip..."
zip -r  ${ZIP_FILE_NAME} *
mv ${ZIP_FILE_NAME} ../artifacts/
