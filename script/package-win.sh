#!/bin/bash

APP_NAME="image-sorter"
EXE_NAME="${APP_NAME}.exe"
VERSION="$(cat VERSION)-$(git rev-parse --short HEAD)"

BIN_SRC_DIR="out/windows"
MSYS_DIR="/c/msys64/mingw64"
DEPLOY_DIR_WIN="./deploy/windows"
DEPLOY_DIR_WIN_BIN="${DEPLOY_DIR_WIN}/bin"
DEPLOY_DIR_WIN_LIB="${DEPLOY_DIR_WIN}/lib"
DEPLOY_DIR_WIN_ASSETS="${DEPLOY_DIR_WIN_BIN}/assets"
ZIP_BASE_DIR="./deploy/windows"
ZIP_DIR_NAME="${APP_NAME}"
ZIP_FILE_NAME="${TAR_DIR_NAME}-${VERSION}.zip"

# Copy binary the
cp "${BIN_SRC_DIR}/${EXE_NAME}" ${DEPLOY_DIR_WIN_BIN}/${EXE_NAME}

# Copy DLLs for Windows
mkdir -p "${DEPLOY_DIR_WIN_BIN}"
ldd "${DEPLOY_DIR_WIN_BIN}/image-sorter.exe" | sed -n 's/\([^ ]*\) => \(\/.*msys64.*\.dll\).*/\2/p' | sort | xargs cp -t ${DEPLOY_DIR_WIN_BIN}
cp "${MSYS_DIR}/bin/gdbus.exe" ${DEPLOY_DIR_WIN_BIN}
cp "${MSYS_DIR}/bin/libbz2-1.dll" ${DEPLOY_DIR_WIN_BIN}
cp "${MSYS_DIR}/bin/libcroco-0.6-3.dll" ${DEPLOY_DIR_WIN_BIN}
cp "${MSYS_DIR}/bin/libexpat-1.dll" ${DEPLOY_DIR_WIN_BIN}
cp "${MSYS_DIR}/bin/libgcc_s_seh-1.dll" ${DEPLOY_DIR_WIN_BIN}
cp "${MSYS_DIR}/bin/libiconv-2.dll" ${DEPLOY_DIR_WIN_BIN}
cp "${MSYS_DIR}/bin/libintl-8.dll" ${DEPLOY_DIR_WIN_BIN}
cp "${MSYS_DIR}/bin/libjpeg-8.dll" ${DEPLOY_DIR_WIN_BIN}
cp "${MSYS_DIR}/bin/liblzma-5.dll" ${DEPLOY_DIR_WIN_BIN}
cp "${MSYS_DIR}/bin/libpcre-1.dll" ${DEPLOY_DIR_WIN_BIN}
cp "${MSYS_DIR}/bin/librsvg-2-2.dll" ${DEPLOY_DIR_WIN_BIN}
cp "${MSYS_DIR}/bin/libstdc++-6.dll" ${DEPLOY_DIR_WIN_BIN}
cp "${MSYS_DIR}/bin/libturbojpeg.dll" ${DEPLOY_DIR_WIN_BIN}
cp "${MSYS_DIR}/bin/libwinpthread-1.dll" ${DEPLOY_DIR_WIN_BIN}
cp "${MSYS_DIR}/bin/libxml2-2.dll" ${DEPLOY_DIR_WIN_BIN}
cp "${MSYS_DIR}/bin/zlib1.dll" ${DEPLOY_DIR_WIN_BIN}

# Copy Pixbuf loaders
mkdir -p "${DEPLOY_DIR_WIN_LIB}/gdk-pixbuf-2.0/2.10.0/loaders"
cp ${MSYS_DIR}/lib/gdk-pixbuf-2.0/2.10.0/loaders/libpixbufloader-svg* "${DEPLOY_DIR_WIN_LIB}/gdk-pixbuf-2.0/2.10.0/loaders"
cp ${MSYS_DIR}/lib/gdk-pixbuf-2.0/2.10.0/loaders/libpixbufloader-svg* "${DEPLOY_DIR_WIN_LIB}/gdk-pixbuf-2.0/2.10.0/loaders"
cp ${MSYS_DIR}/lib/gdk-pixbuf-2.0/2.10.0/loaders/libpixbufloader-png* "${DEPLOY_DIR_WIN_LIB}/gdk-pixbuf-2.0/2.10.0/loaders"

# Copy gio modules
mkdir -p "${DEPLOY_DIR_WIN_LIB}/gio/modules"
cp ${MSYS_DIR}/lib/gio/modules/*.dll "${DEPLOY_DIR_WIN_LIB}/gio/modules"

# General assets
cp style.css ${DEPLOY_DIR_WIN_BIN}/style.css
cp main-view.glade ${DEPLOY_DIR_WIN_BIN}/main-view.glade

# Icon
mkdir -p "${DEPLOY_DIR_WIN_ASSETS}"
cp assets/icon-*.png ${DEPLOY_DIR_WIN_ASSETS}


cd ${ZIP_BASE_DIR}
echo "Create zip..."
zip -r  ${ZIP_FILE_NAME} ${ZIP_DIR_NAME}
mv ${ZIP_FILE_NAME} ../artifacts/