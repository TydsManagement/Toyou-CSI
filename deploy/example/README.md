## How to test toyou csi block driver with Kubernetes 1.13+ for basic functions

```bash
$ kubectl create -f storageclass.yaml
$ kubectl create -f pvc.yaml
$ kubectl create -f pod.yaml
```
## How to test toyou csi block driver with Kubernetes 1.13+ for snapshot
```bash
$ kubectl create -f snapshotstoreclass.yaml
$ kubectl create -f snapshot.yaml
$ kubectl create -f restore.yaml
```
## How to test toyou csi block driver with kubernetes 1.13+ for clone
```
```
## How to test toyou csi block driver with kubernetes 1.13+ for raw block
```
```

Other helper scripts:
* `logs.sh` output of the plugin
* `exec-bash.sh` logs into the plugin's container and runs bash
