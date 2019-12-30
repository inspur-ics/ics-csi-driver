#!/bin/sh
kubectl create secret generic ics-config-secret --from-file=../../example/cfg/icsphere-csi.conf -n kube-system
kubectl get secret ics-config-secret -n kube-system
