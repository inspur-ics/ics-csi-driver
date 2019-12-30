#!/bin/sh
export CSI_ENDPOINT=unix:///var/lib/csi/sockets/pluginproxy/csi.sock
export X_CSI_MODE="controller"
export VSPHERE_CSI_CONFIG="/etc/cloud/csi-vsphere.conf"
