#!/bin/bash

VERSION="0.1-1"
DEPLOY_DIR="deploy/linux/image-sorter_${VERSION}"

DEB_CONTROL_DIR="${DEPLOY_DIR}/DEBIAN"
DEB_BIN_DIR="${DEPLOY_DIR}/usr/local/bin"
DEB_RESOURCE_DIR="${DEPLOY_DIR}/usr/local/bin"

mkdir -p ${DEPLOY_DIR}/usr/local/bin
mkdir -p ${DEPLOY_DIR}/DEBIAN

cp script/control-template ${DEB_CONTROL_DIR}
cp main-view.glade ${RESOURCE_DEB_DIR}
cp image-sorter ${DEB_BIN_DIR}
cp style.css ${DEB_RESOURCE_DIR}

dpkg -i image-sorter_${VERSION}.deb
