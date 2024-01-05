#!/bin/bash

deployment_base="${1:-./}"

cd "$deployment_base" || exit 1

objects=(
    csi-sidecar-iscsi-attacher-rbac
    csi-sidecar-iscsi-provisioner-rbac
    csi-toyou-iscsi-driver-rbac
    csi-sidecar-iscsi-attacher
    csi-sidecar-iscsi-provisioner
    csi-toyou-iscsi-driver
    csi-sidecar-iscsi-resizer-rbac
    csi-sidecar-iscsi-resizer
    setup-csi-snapshotter
    setup-snapshot-controller
)

for obj in "${objects[@]}"; do
    kubectl delete -f "./${obj}.yaml"
done