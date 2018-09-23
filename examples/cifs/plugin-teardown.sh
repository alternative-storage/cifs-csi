#!/bin/bash

deployment_base="${1}"

if [[ -z $deployment_base ]]; then
	deployment_base="../../deploy/cifs/kubernetes"
fi

cd "$deployment_base" || exit 1

objects=(csi-cifsplugin-attacher csi-cifsplugin-provisioner csi-cifsplugin csi-attacher-rbac csi-provisioner-rbac csi-nodeplugin-rbac)

for obj in ${objects[@]}; do
	kubectl delete -f "./$obj.yaml"
done
