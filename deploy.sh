#!/usr/bin/env sh

#TODO add UI=1 as option!
set -Eeou pipefail

echo "Uninstalling helm release"
helm uninstall --namespace=chaos-testing chaos-mesh || true &

echo "Running make generate"
make generate > /dev/null

echo "Running make yaml"
make yaml > /dev/null

echo "Running make (this may take a long time)"
UI=1 make > /dev/null

echo "Running make docker-push"
make docker-push > /dev/null


echo "Applying Manifests"
kubectl apply -f manifests

echo "Installing helm release"
helm install chaos-mesh helm/chaos-mesh --namespace=chaos-testing --set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock --set dashboard.create=true

echo "Applying clusterrole (temporary fix)"
kubectl apply -f clusterrole.yaml

#TODO this must wait for the dashbaord to be running
echo "Port forward for chaos-dashboard"
kubectl port-forward -n chaos-testing svc/chaos-dashboard 2333:2333 &
