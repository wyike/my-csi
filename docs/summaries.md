### draft (in progress)

csi controller plugin and node plugin functionalities are implemented in one code registry, we call it a csi driver in general

![volume](images/plugins.jpeg)


csi-controller
- creates a volume
- publish the volume to worker node

csi-node-plugin
- stage the volume to staging dir on the node vm
- publish (bind mount) the staged volume to pod dir


![volume](images/diagram.jpeg)
The workflow also differs on different types of volumes

![volume](images/diff-volumes.jpeg)




