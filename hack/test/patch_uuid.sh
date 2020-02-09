#!/bin/sh

provider_prefix="ics://"
product_uuid=$(cat /sys/class/dmi/id/product_uuid)
nodename=$(kubectl get node --no-headers | awk '{print $1}' | grep $(hostname))

if [ ! -z $nodename ]; then
    kubectl patch node $nodename -p '{"spec":{"providerID":"'$provider_prefix$product_uuid'"}}'
    kubectl describe node $nodename | grep -i providerID
else
    echo "k8s node not found"
fi

