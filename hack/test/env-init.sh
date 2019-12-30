#!/bin/sh
export CSI_ENDPOINT=unix:///var/lib/csi/sockets/pluginproxy/csi.sock
export X_CSI_MODE="controller"
export ICSPHERE_CSI_CONFIG="/etc/ics/icsphere-csi.conf"
