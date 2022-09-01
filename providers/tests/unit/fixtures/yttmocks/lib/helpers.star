load("@ytt:struct", "struct")
load("@ytt:assert", "assert")

valid_md_rollout_strategy_types = ["OnDelete", "RollingUpdate"]

# verify_and_configure_machine_deployment_rollout_strategy, verify strategy type input against allowed type and return type if correct.
def verify_and_configure_machine_deployment_rollout_strategy(strategy_type):
   if strategy_type not in valid_md_rollout_strategy_types:
      strs = ", ".join(valid_md_rollout_strategy_types)
      assert.fail("Invalid Strategy type, Allowed values: \""+strs+"\"")
   end
   return strategy_type
end

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

def get_tkr_version_from_tkr_name(tkr_name):
   strs = tkr_name.split("---")
   return strs[0] + "+" + strs[1]
end

def map(f, list):
    return [f(x) for x in list]
end