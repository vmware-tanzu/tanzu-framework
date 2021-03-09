### vSphere CSI manifests

vSphere CSI 6.7u3 source: https://github.com/kubernetes-sigs/vsphere-csi-driver/tree/release-2.1/manifests/v2.1.0/vsphere-67u3/vanilla

vSphere CSI 7.0 source: https://github.com/kubernetes-sigs/vsphere-csi-driver/blob/master/manifests/v2.1.1/vsphere-7.0u1/vanilla

To customize deployment for example to add csi-resizer to vSphere 7.0, update the addon secret with overlays as shown below.
```yaml
apiVersion: v1
kind: Secret
stringData:
  values.yaml: |
    #@data/values
    #@overlay/match-child-defaults missing_ok=True
    ---
    vsphereCSI:
      CSIAttacherImage:
        repository: projects-stg.registry.vmware.com/tkg
        path: csi/csi-attacher
        tag: v3.0.0_vmware.1
        pullPolicy: IfNotPresent
      vsphereCSIControllerImage:
        repository: projects-stg.registry.vmware.com/tkg
        path: csi/vsphere-block-csi-driver
        tag: v2.1.0_vmware.1
        pullPolicy: IfNotPresent
      livenessProbeImage:
        repository: projects-stg.registry.vmware.com/tkg
        path: csi/csi-livenessprobe
        tag: v2.1.0_vmware.1
        pullPolicy: IfNotPresent
      vsphereSyncerImage:
        repository: projects-stg.registry.vmware.com/tkg
        path: csi/volume-metadata-syncer
        tag: v2.1.0_vmware.1
        pullPolicy: IfNotPresent
      CSIProvisionerImage:
        repository: projects-stg.registry.vmware.com/tkg
        path: csi/csi-provisioner
        tag: v2.0.0_vmware.1
        pullPolicy: IfNotPresent
      CSINodeDriverRegistrarImage:
        repository: projects-stg.registry.vmware.com/tkg
        path: csi/csi-node-driver-registrar
        tag: v2.0.1_vmware.1
        pullPolicy: IfNotPresent
      namespace: kube-system
      clusterName: wc-1
      server: 10.92.127.215
      datacenter: /dc0
      publicNetwork: VM Network
      username: administrator@vsphere.local
      password: Admin!23
  overlays.yaml: |
    #@ load("@ytt:overlay", "overlay")

    #@overlay/match by=overlay.subset({"kind": "Deployment", "metadata": {"name": "vsphere-csi-controller"}})
    ---
    spec:
      template:
        spec:
          containers:
          #@overlay/append
            - name: csi-resizer
              image: quay.io/k8scsi/csi-resizer:v1.0.0
              args:
                - "--v=4"
                - "--timeout=300s"
                - "--csi-address=$(ADDRESS)"
                - "--leader-election"
              env:
                - name: ADDRESS
                  value: /csi/csi.sock
              volumeMounts:
                - mountPath: /csi
                  name: socket-dir
  ```