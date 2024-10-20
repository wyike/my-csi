# my-csi


## Simple test
### 1. deploy controller-plugin and node-plugin, using manifests/deploy
```
kubo@XMlgbMxLVEskU:~/my-csi$ kubectl apply -f manifests/deploy
deployment.apps/my-csi created
clusterrole.rbac.authorization.k8s.io/my-csi-cr created
clusterrolebinding.rbac.authorization.k8s.io/my-csi-crb created
daemonset.apps/node-plugin created
serviceaccount/my-csi-sa created

kubo@XMlgbMxLVEskU:~/my-csi$ kubectl get pod
NAME                      READY   STATUS    RESTARTS   AGE
my-csi-56ccfdc66d-wv9vr   3/3     Running   0          6s
node-plugin-554hr         2/2     Running   0          2s
node-plugin-5ftnj         2/2     Running   0          4s
node-plugin-7jj9q         2/2     Running   0          6s

# one more new driver is registered on each node
kubo@XMlgbMxLVEskU:~$ kubectl get csinode 
NAME                                          DRIVERS   AGE
wl-antrea-md-0-lndvf-74b758648bx5zbwk-thjkw   2         4d1h
wl-antrea-md-1-swl28-65d7b87ccdx69hjk-m7fgx   2         4d1h
wl-antrea-md-2-xnsqs-67d5b9747dxcwz96-78wpz   2         4d1h
```
### 2. create storage class, using manifests/test/sc.yaml
```
kubo@XMlgbMxLVEskU:~$ kubectl get sc my-csi
NAME     PROVISIONER   RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
my-csi   my-csi        Delete          Immediate           false                  3d23h
```
### 3. create a PVC, using mafenists/test/test/pvc.yaml
```
kubo@XMlgbMxLVEskU:~/my-csi$ kubectl apply -f manifests/test/pvc.yaml 
persistentvolumeclaim/myclaim created
kubo@XMlgbMxLVEskU:~/my-csi$ kubectl get pvc
NAME      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
myclaim   Bound    pvc-b7764996-1743-43a8-920f-676532db1aab   5Gi        RWO            my-csi         5s
kubo@XMlgbMxLVEskU:~/my-csi$ kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM             STORAGECLASS   REASON   AGE
pvc-b7764996-1743-43a8-920f-676532db1aab   5Gi        RWO            Delete           Bound    default/myclaim   my-csi   
```
A pv is created automatically.

### 4. create a Pod with the pvc
```
kubo@XMlgbMxLVEskU:~$ kubectl apply -f pod.yaml
pod/task-pv-pod created
```
volumeattachment is created automatically:
```
kubo@XMlgbMxLVEskU:~/my-csi$ kubectl get volumeattachment
NAME                                                                   ATTACHER   PV                                         NODE                                          ATTACHED   AGE
csi-696dde9a71ed8b1cbc64b786bb2c669ef035b535fc37694733f6a83784b64065   my-csi     pvc-b7764996-1743-43a8-920f-676532db1aab   wl-antrea-md-1-swl28-65d7b87ccdx69hjk-m7fgx   true       57s
kubo@XMlgbMxLVEskU:~/my-csi$ kubectl get volumeattachment -oyaml | grep diskUUID
      diskUUID: 6000c299f18b4235370bc9297e2eef60
```
pod events:
```
kubo@XMlgbMxLVEskU:~$ kubectl describe pod task-pv-pod
Name:             task-pv-pod
Namespace:        default
...
  Type    Reason                  Age   From                     Message
  ----    ------                  ----  ----                     -------
  Normal  Scheduled               79s   default-scheduler        Successfully assigned default/task-pv-pod to wl-antrea-md-1-swl28-65d7b87ccdx69hjk-m7fgx
  Normal  SuccessfulAttachVolume  78s   attachdetach-controller  AttachVolume.Attach succeeded for volume "pvc-b7764996-1743-43a8-920f-676532db1aab"
  Normal  Pulling                 68s   kubelet                  Pulling image "nginx:1.17"
  Normal  Pulled                  66s   kubelet                  Successfully pulled image "nginx:1.17" in 1.729296752s (1.729342237s including waiting)
  Normal  Created                 66s   kubelet                  Created container task-pv-container
  Normal  Started                 66s   kubelet                  Started container task-pv-container
```

### 5. logs in controller-plugin and node-plugin:
```
kubectl logs my-csi-fd49b8565-kthzz -c controller-plugin
...
CreateVolume of the controller service was called
Creating volume using the spec: ...
volumeCreateResult &{CnsVolumeOperationResult:{DynamicData:{} VolumeId:{DynamicData:{} Id:f902fee8-f463-4260-a67f-f21fc320b0cb} Fault:<nil>} Name:pvc-b7764996-1743-43a8-920f-676532db1aab PlacementResults:[{Datastore:Datastore:datastore-43 PlacementFaults:[]}]}
Volume created sucessfully. volumeId: f902fee8-f463-4260-a67f-f21fc320b0cb
ControllerPublishVolume of the controller service was called
Attaching volume using the spec: {DynamicData:{} VolumeId:{DynamicData:{} Id:f902fee8-f463-4260-a67f-f21fc320b0cb} Vm:VirtualMachine:vm-69}
Volume attached sucessfully. Disk UUID: 6000C299-f18b-4235-370b-c9297e2eef60
ControllerPublishVolume of the controller service was called
Attaching volume using the spec: {DynamicData:{} VolumeId:{DynamicData:{} Id:f902fee8-f463-4260-a67f-f21fc320b0cb} Vm:VirtualMachine:vm-69}
```

```
kubo@XMlgbMxLVEskU:~$ kubectl logs node-plugin-f8d2v -c node-plugin
...
NodeGetCapabilities of node serivce was called
NodeGetCapabilities of node serivce was called
NodeGetCapabilities of node serivce was called
NodeStageVolume was called
time="2023-11-19T03:20:00Z" level=info msg="attempting to mount disk" fsType=ext4 options="[defaults]" source=/dev/disk/by-id/wwn-0x6000c299f18b4235370bc9297e2eef60 target=/var/lib/kubelet/plugins/kubernetes.io/csi/my-csi/d08d403ad6988c61773651e8849be90a8a465b230f041172109bd947cc8efe0a/globalmount
time="2023-11-19T03:20:00Z" level=info msg="mount command" args="-t ext4 -o defaults /dev/disk/by-id/wwn-0x6000c299f18b4235370bc9297e2eef60 /var/lib/kubelet/plugins/kubernetes.io/csi/my-csi/d08d403ad6988c61773651e8849be90a8a465b230f041172109bd947cc8efe0a/globalmount" cmd=mount
time="2023-11-19T03:20:00Z" level=error msg="mount Failed" args="-t ext4 -o defaults /dev/disk/by-id/wwn-0x6000c299f18b4235370bc9297e2eef60 /var/lib/kubelet/plugins/kubernetes.io/csi/my-csi/d08d403ad6988c61773651e8849be90a8a465b230f041172109bd947cc8efe0a/globalmount" cmd=mount error="exit status 32" output="mount: /var/lib/kubelet/plugins/kubernetes.io/csi/my-csi/d08d403ad6988c61773651e8849be90a8a465b230f041172109bd947cc8efe0a/globalmount: wrong fs type, bad option, bad superblock on /dev/sdc, missing codepage or helper program, or other error.\n       dmesg(1) may have more information after failed mount system call.\n"
time="2023-11-19T03:20:00Z" level=info msg="checking if disk is formatted using lsblk" args="[-n -o FSTYPE /dev/disk/by-id/wwn-0x6000c299f18b4235370bc9297e2eef60]" disk=/dev/disk/by-id/wwn-0x6000c299f18b4235370bc9297e2eef60
time="2023-11-19T03:20:00Z" level=info msg="disk appears unformatted, attempting format" fsType=ext4 options="[defaults]" source=/dev/disk/by-id/wwn-0x6000c299f18b4235370bc9297e2eef60 target=/var/lib/kubelet/plugins/kubernetes.io/csi/my-csi/d08d403ad6988c61773651e8849be90a8a465b230f041172109bd947cc8efe0a/globalmount
time="2023-11-19T03:20:00Z" level=info msg="disk successfully formatted" fsType=ext4 options="[defaults]" source=/dev/disk/by-id/wwn-0x6000c299f18b4235370bc9297e2eef60 target=/var/lib/kubelet/plugins/kubernetes.io/csi/my-csi/d08d403ad6988c61773651e8849be90a8a465b230f041172109bd947cc8efe0a/globalmount
time="2023-11-19T03:20:00Z" level=info msg="mount command" args="-t ext4 -o defaults /dev/disk/by-id/wwn-0x6000c299f18b4235370bc9297e2eef60 /var/lib/kubelet/plugins/kubernetes.io/csi/my-csi/d08d403ad6988c61773651e8849be90a8a465b230f041172109bd947cc8efe0a/globalmount" cmd=mount
node staging mount volume: format and mount the volume ID f902fee8-f463-4260-a67f-f21fc320b0cbnodeStageBlockVolume: Format and mount the device /dev/disk/by-id/wwn-0x6000c299f18b4235370bc9297e2eef60 at /var/lib/kubelet/plugins/kubernetes.io/csi/my-csi/d08d403ad6988c61773651e8849be90a8a465b230f041172109bd947cc8efe0a/globalmount with mount flags []NodeGetCapabilities of node serivce was called
NodeGetCapabilities of node serivce was called
NodeGetCapabilities of node serivce was called
NodePublishVolume was called with source /var/lib/kubelet/plugins/kubernetes.io/csi/my-csi/d08d403ad6988c61773651e8849be90a8a465b230f041172109bd947cc8efe0a/globalmount and target /var/lib/kubelet/pods/1abca335-484b-4ae2-af8b-b05b56ff2963/volumes/kubernetes.io~csi/pvc-b7764996-1743-43a8-920f-676532db1aab/mount
time="2023-11-19T03:20:00Z" level=info msg="mount command" args="-o bind /var/lib/kubelet/plugins/kubernetes.io/csi/my-csi/d08d403ad6988c61773651e8849be90a8a465b230f041172109bd947cc8efe0a/globalmount /var/lib/kubelet/pods/1abca335-484b-4ae2-af8b-b05b56ff2963/volumes/kubernetes.io~csi/pvc-b7764996-1743-43a8-920f-676532db1aab/mount" cmd=mount
time="2023-11-19T03:20:00Z" level=info msg="mount command" args="-o remount /var/lib/kubelet/plugins/kubernetes.io/csi/my-csi/d08d403ad6988c61773651e8849be90a8a465b230f041172109bd947cc8efe0a/globalmount /var/lib/kubelet/pods/1abca335-484b-4ae2-af8b-b05b56ff2963/volumes/kubernetes.io~csi/pvc-b7764996-1743-43a8-920f-676532db1aab/mount" cmd=mount
NodePublishVolume for f902fee8-f463-4260-a67f-f21fc320b0cb successful to path /var/lib/kubelet/pods/1abca335-484b-4ae2-af8b-b05b56ff2963/volumes/kubernetes.io~csi/pvc-b7764996-1743-43a8-920f-676532db1aab/mountNodeGetCapabilities of node serivce was called
NodeGetCapabilities of node serivce was called
```

### 6. Delete pod
```
kubo@XMlgbMxLVEskU:~/my-csi$ kubectl delete pod task-pv-pod
pod "task-pv-pod" deleted
```
node plugin continues reporting:
```
NodeUnpublishVolume of node serivce was called to unmount target /var/lib/kubelet/pods/1abca335-484b-4ae2-af8b-b05b56ff2963/volumes/kubernetes.io~csi/pvc-b7764996-1743-43a8-920f-676532db1aab/mount for volume f902fee8-f463-4260-a67f-f21fc320b0cb
time="2023-11-19T03:23:30Z" level=info msg="unmount command" cmd=umount path="/var/lib/kubelet/pods/1abca335-484b-4ae2-af8b-b05b56ff2963/volumes/kubernetes.io~csi/pvc-b7764996-1743-43a8-920f-676532db1aab/mount"
NodeUnpublishVolume successful for volume f902fee8-f463-4260-a67f-f21fc320b0cb
NodeGetCapabilities of node serivce was called
time="2023-11-19T03:23:30Z" level=info msg="unmount command" cmd=umount path=/var/lib/kubelet/plugins/kubernetes.io/csi/my-csi/d08d403ad6988c61773651e8849be90a8a465b230f041172109bd947cc8efe0a/globalmount
NodeUnstageVolume of node serivce was called to unmount target /var/lib/kubelet/plugins/kubernetes.io/csi/my-csi/d08d403ad6988c61773651e8849be90a8a465b230f041172109bd947cc8efe0a/globalmount for volume f902fee8-f463-4260-a67f-f21fc320b0cb
NodeUnstageVolume successful for target /var/lib/kubelet/plugins/kubernetes.io/csi/my-csi/d08d403ad6988c61773651e8849be90a8a465b230f041172109bd947cc8efe0a/globalmount for volume f902fee8-f463-4260-a67f-f21fc320b0cb
```
controller plugin continues reporting:
```
ControllerUnpublishVolume of the controller service was called
Detaching volume using the spec: {DynamicData:{} VolumeId:{DynamicData:{} Id:f902fee8-f463-4260-a67f-f21fc320b0cb} Vm:VirtualMachine:vm-69}
Volume detached sucessfully
```

### 7. Delete pvc
```
kubo@XMlgbMxLVEskU:~$ kubectl delete pvc myclaim
persistentvolumeclaim "myclaim" deleted
kubo@XMlgbMxLVEskU:~$ kubectl get pvc
No resources found in default namespace.
kubo@XMlgbMxLVEskU:~$ kubectl get pv
No resources found
```

controller plugin continues reporting:
```
DeleteVolume of the controller service was called
Volume: "f902fee8-f463-4260-a67f-f21fc320b0cb" deleted sucessfully
```


#### Sample look of vsphere volume

![volume](images/volume.jpg)

One more vmdk of 1G will be attached to the VM