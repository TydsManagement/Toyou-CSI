# Toyou CSI Drivers

## Overview
Toyou CSI drivers provide storage container storage interface for Toyou Tyds Storage Platform.

The Toyou CSI plugins establish an interface between a CSI-enabled Container Orchestrator (CO) and the Toyou cluster.

The Toyou-CSI (Toyou Tyds Storage Platform) CSI plugins currently support the following features:
- Block storage
  - The Tyds Block CSI plugin supports block storage, and it can be used in ReadWriteOnce mode. It allows dynamic volume creation and mounting into workloads.

- [ ] File storage
  - This feature is currently under development.
- [ ] Object storage
  - This feature is currently under development.



## Deployment

To deploy the Toyou CSI drivers, follow the steps below. Please note that the version number in the path (`v1-0-1`) may vary. Replace it with the actual version you are deploying.

1. Create a new namespace for the Toyou CSI drivers:
   ```
   kubectl create namespace toyou-csi
   ```

   This command will create a new namespace called `toyou-csi` to isolate the Toyou CSI drivers from other system components.

2. Change to the directory containing the CSI YAML files:
   ```
   cd deploy/kubernetes/release/v1-0-1
   ```

3. Apply the YAML files using `kubectl`:
   ```
   kubectl apply -f .
   ```

   This command will apply all the YAML files in the current directory, configuring the necessary resources for the Toyou CSI drivers within the `toyou-csi` namespace.

4. Verify the deployment:
   ```
   kubectl get pods -n toyou-csi
   ```

   Use the above command to check the status of the Toyou CSI driver pods running in the `toyou-csi` namespace. You should see the `toyou-csi-controller` and `toyou-csi-node` pods listed.

5. (Optional) Customize the deployment:

    - If you need to modify any configuration parameters, refer to the documentation or comments within the YAML files to make the necessary changes before applying them.
    - Additionally, you can refer to the Toyou CSI driver documentation for any advanced configuration or customization options.


## Usage

To use the Toyou CSI drivers, you can follow the steps below. Please note that the examples provided assume that you have already deployed the Toyou CSI drivers as described in the deployment instructions.

1. Change to the directory containing the example YAML files:
   ```
   cd deploy/example
   ```

2. Apply the example PersistentVolumeClaim (PVC):
   ```
   kubectl apply -f pvc.yaml
   ```

   This command will create a PersistentVolumeClaim (PVC) object, which represents a request for storage resources from the Toyou CSI drivers.

3. Apply the example Pod:
   ```
   kubectl apply -f pod.yaml
   ```

   This command will create a Pod object that uses the previously created PVC to request storage from the Toyou CSI drivers.

4. Verify the deployment:
   ```
   kubectl get pods
   ```

   Use the above command to check the status of the Pod. Once the Pod is in the "Running" state, it indicates that the storage has been successfully provisioned and attached to the Pod.

5. (Optional) Customize the deployment:

    - If you need to modify any configuration parameters or resources in the example YAML files, you can edit the files before applying them with `kubectl apply`.
    - Additionally, you can refer to the Toyou CSI driver documentation for advanced usage and configuration options.


## Building from Source
You can build a binary using the following command:
```
make disk-container
```

## Versioning and Compatibility Policy
The Toyou CSI drivers adhere to Semantic Versioning.

Minor versions align with the minor version number of the latest Kubernetes release. Version 0.y is compatible with Kubernetes version 1.y. Unless specified otherwise in the compatibility matrix below, CSI 0.y is also compatible with the two most recent Kubernetes minor versions before 1.y. Older versions may work, but without guarantee. For example, CSI 0.99 would be compatible with Kubernetes versions 1.99, 1.98, and 1.97.

### Compatibility Matrix
| CSI Version | Supported Kubernetes Versions |
|-------------|-------------------------------|
| 0.29        | 1.29,1.28,1.27                |

## Project Status
The Toyou CSI drivers are currently in the Beta phase. While they reliably perform their essential functions, missing features and bugs may still be encountered.