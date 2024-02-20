#!/bin/bash

# shellcheck disable=SC2164
cd /root/GoPath/src/Toyou-CSI/deploy/kubernetes/realease/v1-0-1/

kubectl delete -f csi-toyou-iscsi-driver.yaml

kubectl apply -f csi-toyou-iscsi-driver.yaml