#!/bin/sh
provider_prefix="ics://"
system_uuid=$(dmidecode -s system-uuid)
nodename=$(kubectl get node --no-headers | awk '{print $1}' | grep $(hostname))

if [ ! -z $nodename ]; then
    kubectl patch node $nodename -p '{"spec":{"providerID":"'$provider_prefix$system_uuid'"}}'
    kubectl describe node $nodename | egrep "Name:|ProviderID:"
else
    echo "k8s node not found"
fi

