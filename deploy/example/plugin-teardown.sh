#!/bin/bash

deployment_base="${1}"

if [[ -z $deployment_base ]]; then
	deployment_base="../deploy/kubernetes"
fi

cd "$deployment_base" || exit 1

objects=(csi-sider-iscsi-attacher-rbac csi-sider-iscsi-provisioner-rbac csi-xsky-iscsi-driver-rbac csi--sider-iscsi-attacher csi--sider-iscsi-provisioner csi-xsky-iscsi-driver
csi-sidecar-iscsi-resizer-rbac csi-sidecar-iscsi-resizer csi-sidecar-iscsi-snapshotter-rabc csi-sidecar-iscsi-snapshotter)

for obj in ${objects[@]}; do
	kubectl delete -f "./$obj.yaml"
done
