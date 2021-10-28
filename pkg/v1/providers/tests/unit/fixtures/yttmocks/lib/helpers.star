load("@ytt:struct", "struct")

def get_bom_data_for_tkr_name():
    return struct.encode({
        "components": {
            "antrea": [{
                "images":  {
                    "antreaImage": {
                        "imagePath": "fake-antrea-image"
                    }
                }
            }],
            "cloud_provider_vsphere": [{
                "images": {
                    "ccmControllerImage": {
                        "imagePath": "fake-ccm-image",
                        "tag": "fake-ccm-tag"
                    }
                }
            }]
        },
        "kubeadmConfigSpec": {
            "imageRepository": "fake-kubeadm-image-repo",
            "etcd": {
                "local": {
                    "imageRepository": "fake-etcd-image-repo",
                    "imageTag": "fake-etcd-image-tag"
                }
            },
            "dns": {
                "imageRepository": "fake-dns-image-repo",
                "imageTag": "fake-dns-image-tag"
            },
        }
    })
end

def tkg_image_repo():
    return "fake-tkg-image-repo"
end

def get_image_repo_for_component(image):
    return "fake-image-repo"
end

def get_default_tkg_bom_data():
    return struct.encode({
        "components": {
            "kube-vip": [{
                "images": {
                    "kubeVipImage": {
                        "imagePath": "fake-kube-vip-image-path",
                        "tag": "fake-kube-vip-tag"
                    }
                }
            }]
        }
    })
end

def kubeadm_image_repo(image_repo):
    return "fake-kubeadm-image-repo"
end

def get_vsphere_thumbprint():
    return "fake-thumbprint"
end

def get_no_proxy():
    return "fake-no-proxy"
end
