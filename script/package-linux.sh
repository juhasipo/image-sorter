#!/bin/bash

set -e

VERSION="0.1-1"

# Source dirs
BIN_SRC_DIR="out/linux"
RESOURCE_SRC_DIR="."

# Destination dirs
DEPLOY_BASE_DIR="deploy/linux"
DEPLOY_DIR="${DEPLOY_BASE_DIR}/image-sorter_${VERSION}"
DEB_CONTROL_DIR="${DEPLOY_DIR}/DEBIAN"
BIN_DST_DIR="${DEPLOY_DIR}/usr/local/bin"
RESOURCE_DEST_DIR="${DEPLOY_DIR}/usr/local/bin"

# Create directory structure for deployment
echo "Create directories..."
mkdir -p ${DEPLOY_DIR}/usr/local/bin
mkdir -p ${DEPLOY_DIR}/DEBIAN

# Copy binaries and resources
echo "Copy binaries and resources..."
cp script/control-template ${DEB_CONTROL_DIR}/control
cp ${BIN_SRC_DIR}/image-sorter ${BIN_DST_DIR}
cp ${RESOURCE_SRC_DIR}/main-view.glade ${RESOURCE_DEST_DIR}
cp ${RESOURCE_SRC_DIR}/style.css ${RESOURCE_DEST_DIR}

# Create DEB package
echo "Create DEB..."
cd ${DEPLOY_BASE_DIR}
dpkg-deb --build image-sorter_${VERSION}
