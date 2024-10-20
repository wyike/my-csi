# my-csi


This is a sample CSI plugin implemented on vSphere cloud native storage. 
It's only a minimum implementation to make basic things work and learn to write an own CSI plugin skeleton.

#### Reference
- How to write a CSI plugin from scrach:
  - video series: https://www.youtube.com/watch?v=OIpX7WkJzOg&list=PLh4KH3LtJvRSAQsRLNLMDu6hd1uh6ZMoR (many thinks to Vivek)

- Also refer some code specific for vsphere from:
  - https://github.com/vmware/govmomi
  - https://github.com/kubernetes-sigs/vsphere-csi-driver/tree/master

#### Usage
- You have a kubernetes cluster
- Preparations
  - Edit manifests/deploy/controller-plugin.yaml to contain your VC Host, VC username, VC password
  - image wyike/my-csi:1.1.0 is on dockerhub. If you want to make yours and pull from your registry, please download source code and run:
    - ```make docker-image  IMAGE_REGISTRY=<your-registry-store-path> IMAGE_NAME=<your-image-name> IMAGE_VERSION=<your-version>```
    - then replace csi image in manifests/deploy/controller-plugin.yaml and manifests/deploy/node-plugin.yaml
- Apply yamls in manifests/deploy to have controller-plugin and node-plugin deployed
- Deploy storageclass with my-csi provisioner: manifests/test/sc.yaml
- Deploy PVC: manifests/test/pvc.yaml
- Deploy a Pod: manifests/test/pod.yaml
- Check a PV can be created automatically and Pod can start successfully
- Delete Pod
- Delete PVC
- Pod and PVC, PV can be deleted successfully
Details please refer to [demo](https://github.com/wyike/my-csi/blob/main/docs/Demo.md)