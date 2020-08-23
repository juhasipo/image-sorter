#!/bin/bash

set -e
APP_NAME="image-sorter"
VERSION="$(cat VERSION)-$(git rev-parse --short HEAD)"

# Source dirs
BIN_SRC_DIR="out/linux"
RESOURCE_SRC_DIR="."

# Destination dirs
DEPLOY_BASE_DIR="deploy/linux"

DEB_DIR_NAME="${APP_NAME}_${VERSION}"
DEB_BASE_DIR="${DEPLOY_BASE_DIR}/deb"
DEB_DEPLOY_DIR="${DEB_BASE_DIR}/${DEB_DIR_NAME}"
DEB_CONTROL_DIR="${DEB_DEPLOY_DIR}/DEBIAN"
DEB_BIN_DST_DIR="${DEB_DEPLOY_DIR}/usr/local/bin"
DEB_RESOURCE_DST_DIR="${DEB_DEPLOY_DIR}/usr/local/bin"

TAR_DIR_NAME="${APP_NAME}"
TAR_FILE_NAME="${TAR_DIR_NAME}-${VERSION}.tar.gz"
TAR_BASE_DIR="${DEPLOY_BASE_DIR}/tar"
TAR_DEPLOY_DIR="${TAR_BASE_DIR}/${TAR_DIR_NAME}"
TAR_BIN_DST_DIR="${TAR_DEPLOY_DIR}"
TAR_RESOURCE_DST_DIR="${TAR_DEPLOY_DIR}"

# Create directory structure for deployment
echo "Create directories..."
mkdir -p ${DEPLOY_BASE_DIR}/artifacts
mkdir -p ${DEPLOY_BASE_DIR}/out
mkdir -p ${DEB_DEPLOY_DIR}/usr/local/bin
mkdir -p ${DEB_DEPLOY_DIR}/DEBIAN

mkdir -p ${TAR_BIN_DST_DIR}
mkdir -p ${TAR_RESOURCE_DST_DIR}

# Copy binaries and resources
echo "Copy binaries and resources..."
cp script/control-template ${DEB_CONTROL_DIR}/control
cp ${BIN_SRC_DIR}/image-sorter ${DEB_BIN_DST_DIR}/${APP_NAME}
cp ${RESOURCE_SRC_DIR}/main-view.glade ${DEB_RESOURCE_DST_DIR}
cp ${RESOURCE_SRC_DIR}/style.css ${DEB_RESOURCE_DST_DIR}

cp ${BIN_SRC_DIR}/image-sorter ${TAR_BIN_DST_DIR}/${APP_NAME}
cp ${RESOURCE_SRC_DIR}/main-view.glade ${TAR_RESOURCE_DST_DIR}
cp ${RESOURCE_SRC_DIR}/style.css ${TAR_RESOURCE_DST_DIR}
cp ${RESOURCE_SRC_DIR}/README.md ${TAR_RESOURCE_DST_DIR}

# Create DEB package
echo "Create DEB..."
cd ${DEB_BASE_DIR}
dpkg-deb --build ${DEB_DIR_NAME}

echo "Move to artifacts dir..."
mv *.deb ../artifacts/

cd ../../..
cd ${TAR_BASE_DIR}
echo "Create tar..."
tar -czvf ${TAR_FILE_NAME} ${TAR_DIR_NAME}
mv ${TAR_FILE_NAME} ../artifacts/
