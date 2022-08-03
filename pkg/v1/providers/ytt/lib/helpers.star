load("@ytt:data", "data")
load("@ytt:assert", "assert")
load("@ytt:regexp", "regexp")

TKGSProductName = "VMware Tanzu Kubernetes Grid Service for vSphere"
TKGProductName = "VMware Tanzu Kubernetes Grid"
# kapp-controller requires the data values for addons to have this header to indicate its a data value
ValuesFormatStr = "#@data/values\n#@overlay/match-child-defaults missing_ok=True\n---\n{}"

valid_md_rollout_strategy_types = ["OnDelete", "RollingUpdate"]

# verify_and_configure_machine_deployment_rollout_strategy, verify strategy type input against allowed type and return type if correct.
def verify_and_configure_machine_deployment_rollout_strategy(strategy_type):
   if strategy_type not in valid_md_rollout_strategy_types:
      strs = ", ".join(valid_md_rollout_strategy_types)
      assert.fail("Invalid Strategy type, Allowed values: \""+strs+"\"")
   end
   return strategy_type
end

def get_tkr_name_from_k8s_version(k8s_version):
   strs = k8s_version.split("+")
   return strs[0] + "---" + strs[1]
end

def get_tkr_version_from_tkr_name(tkr_name):
   strs = tkr_name.split("---")
   return strs[0] + "+" + strs[1]
end

def get_default_tkg_bom_data():
  for bom_entry in data.values.boms:
    if bom_entry.bom_name == data.values.TKG_DEFAULT_BOM:
      return bom_entry.bom_data
    end
  end
  assert.fail("unable to find the default BOM file: " + data.values.TKG_DEFAULT_BOM)
end


def get_default_tkr_bom_data():
   default_tkg_bom = get_default_tkg_bom_data()
   k8s_version = default_tkg_bom.default.k8sVersion

   for bom_entry in data.values.boms:
        if bom_entry.bom_data.release.version == k8s_version:
            return bom_entry.bom_data
        end
   end
   assert.fail("unable to get TanzuKubernetesRelease BoM file for TKG version: "+k8s_version +" " + default_tkg_bom.release.version)
end

def get_bom_data_for_tkr_name():
    for bom_entry in data.values.boms:
        tkr_version = get_tkr_version_from_tkr_name(data.values.KUBERNETES_RELEASE)
        if bom_entry.bom_data.release.version == tkr_version:
            return bom_entry.bom_data
        end
    end
    assert.fail("unable to get BoM file for the TanzuKubernetesRelease version: " + bom_entry.bom_data.release.version + " " +  data.values.KUBERNETES_RELEASE )
end

tkgBomData = get_default_tkg_bom_data()
if data.values.PROVIDER_TYPE != "tkg-service-vsphere":
    tkrBomData = get_bom_data_for_tkr_name()
end

def kubeadm_image_repo(default_repo):
  if data.values.TKG_CUSTOM_IMAGE_REPOSITORY:
    return data.values.TKG_CUSTOM_IMAGE_REPOSITORY
  end

  # if downstream images are not available for Azure, change the kubeadm image repository to the staging repository name(global 'imageRepository' in BOM)
  if data.values.PROVIDER_TYPE == "azure":
    if hasattr(tkrBomData, 'azure') and hasattr(tkrBomData.azure[0], 'thirdPartyImage') and tkrBomData.azure[0].thirdPartyImage == True:
      return default_repo
    else:
      # when downstream image not available
      return tkrBomData.kubeadmConfigSpec.imageRepository
    end
  end

  return default_repo
end

def tkg_image_repo_customized():
    return data.values.TKG_CUSTOM_IMAGE_REPOSITORY != ""
end

#! XXX: until all overlays are refactored to account for cluster class, any
#! scenario involving cc will be processed via supplying a plan name that
#! adheres to the convention of ending in "cc". Any overlay not meant to be used
#! scenario should be guarded by calling this helper.
def tkg_use_clusterclass():
    return data.values.CLUSTER_PLAN != "" and data.values.CLUSTER_PLAN.endswith("cc")
end

def tkg_image_repo():
  return data.values.TKG_CUSTOM_IMAGE_REPOSITORY if data.values.TKG_CUSTOM_IMAGE_REPOSITORY else tkgBomData.imageConfig.imageRepository
end

def get_image_repo_for_component(image):
  # if custom image repo is specified use that repo
  if data.values.TKG_CUSTOM_IMAGE_REPOSITORY:
    return data.values.TKG_CUSTOM_IMAGE_REPOSITORY
  end

  # if imageRepo is specified for the component with image.imageRepository
  # than use image.imageRepository else use default imageRepo from BoM file
  if hasattr(image, 'imageRepository'):
    return image.imageRepository
  else:
    return tkgBomData.imageConfig.imageRepository
  end
end

def tkg_image_repo_skip_tls_verify():
  return data.values.TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY == True and tkg_image_repo_customized()
end

def tkg_image_repo_ca_cert():
  return data.values.TKG_PROXY_CA_CERT if data.values.TKG_PROXY_CA_CERT else data.values.TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE
end

def tkg_image_repo_hostname():
  return tkg_image_repo().split("/")[0]
end

def get_provider():
  if data.values.PROVIDER_TYPE == "tkg-service-vsphere":
    return "vsphere"
  end
  return data.values.PROVIDER_TYPE
end

def get_kubernetes_provider():
  return TKGSProductName if data.values.PROVIDER_TYPE == "tkg-service-vsphere" else TKGProductName
end

def get_az_from_region(region, az, suffix):
  regionWithSuffix = ""
  if region != "" and region != None:
    regionWithSuffix = region + suffix
  end
  return az if az != "" else regionWithSuffix
end

def validate():
  validate_funcs = [validate_oidc]
  for fn in validate_funcs:
    fn()
  end
  return True
end

def validate_oidc():
  if data.values.ENABLE_OIDC :
    data.values.OIDC_ISSUER_URL or assert.fail("oidc enabled, oidc issuer url should be provided")
    data.values.OIDC_USERNAME_CLAIM or assert.fail("oidc enabled, oidc username claim should be provided")
    data.values.OIDC_GROUPS_CLAIM or assert.fail("oidc enabled, oidc groups claim should be provided")
    data.values.OIDC_DEX_CA or assert.fail("oidc enabled, oidc dex ca should be provided.")
  end
end

def get_azure_image(tkrBomData):
  image = get_azure_image_from_config()
  if image != None:
    return image
  end

  return get_azure_image_from_bom(tkrBomData)
end

def get_azure_image_from_config():
  if data.values.AZURE_IMAGE_ID:
    return {
      "id": data.values.AZURE_IMAGE_ID
    }
  end

  if data.values.AZURE_IMAGE_RESOURCE_GROUP:
    sharedGallery = {
      "resourceGroup": data.values.AZURE_IMAGE_RESOURCE_GROUP,
      "name": data.values.AZURE_IMAGE_NAME,
      "subscriptionID": data.values.AZURE_IMAGE_SUBSCRIPTION_ID,
      "gallery": data.values.AZURE_IMAGE_GALLERY,
      "version": data.values.AZURE_IMAGE_VERSION
    }
    return {
      "sharedGallery": sharedGallery
    }
  end

  if data.values.AZURE_IMAGE_PUBLISHER:
    marketplace = {
      "publisher": data.values.AZURE_IMAGE_PUBLISHER,
      "offer": data.values.AZURE_IMAGE_OFFER,
      "sku": data.values.AZURE_IMAGE_SKU,
      "version": data.values.AZURE_IMAGE_VERSION,
      "thirdPartyImage": data.values.AZURE_IMAGE_THIRD_PARTY
    }

    return {
      "marketplace": marketplace
    }
  end

  return None
end

def get_azure_image_from_bom(tkrBomData):

  if not hasattr(tkrBomData, 'azure'):
    fail("no image information in BOM")
  end

  sharedGallery = get_shared_gallery_image(tkrBomData)
  if sharedGallery != None:
    return sharedGallery
  end

  marketplace = get_marketplace_image(tkrBomData)
  if marketplace != None:
    return marketplace
  end

  fail("invalid image information in BOM")
end


def get_shared_gallery_image(tkrBomData):
  keysRequired = ['resourceGroup', 'name', 'subscriptionID', 'gallery', 'version', 'osinfo', 'metadata']
  keysFromBom = dir(tkrBomData.azure[0])
  keysFromBom.append('osinfo')
  keysFromBom.append('metadata')

  if set(keysRequired) == set(keysFromBom):
    sharedGallery = {
      "resourceGroup": tkrBomData.azure[0].resourceGroup,
      "name": tkrBomData.azure[0].name,
      "subscriptionID": tkrBomData.azure[0].subscriptionID,
      "gallery": tkrBomData.azure[0].gallery,
      "version": tkrBomData.azure[0].version
    }
    return {
      "sharedGallery": sharedGallery
    }
  else:
    return None
  end
end

def get_marketplace_image(tkrBomData):
  keysRequired = ['publisher', 'offer', 'sku', 'version', 'thirdPartyImage', 'osinfo', 'metadata']
  keysFromBom = dir(tkrBomData.azure[0])
  keysFromBom.append('thirdPartyImage')
  keysFromBom.append('osinfo')
  keysFromBom.append('metadata')

  if set(keysRequired) == set(keysFromBom):
    marketplace = {
      "publisher": tkrBomData.azure[0].publisher,
      "offer": tkrBomData.azure[0].offer,
      "sku": tkrBomData.azure[0].sku,
      "version": tkrBomData.azure[0].version,
      "thirdPartyImage": False
    }

    if hasattr(tkrBomData.azure[0], 'thirdPartyImage'):
      marketplace["thirdPartyImage"] = tkrBomData.azure[0].thirdPartyImage
    end

    return {
      "marketplace": marketplace
    }
  else:
    return None
  end
end

def get_vsphere_thumbprint():
  if data.values.VSPHERE_INSECURE:
    return ""
  end
  return data.values.VSPHERE_TLS_THUMBPRINT
end

#! Change done in this function needs to be done in `kapp-controller-values/helper.star` as well.
def get_no_proxy():
  if data.values.TKG_HTTP_PROXY != "":
    full_no_proxy_list = []
    if data.values.TKG_NO_PROXY != "":
      # trim space in the no_proxy list
      full_no_proxy_list = regexp.replace(" ", data.values.TKG_NO_PROXY, "").split(",")
    end
    if data.values.PROVIDER_TYPE == "aws":
      if data.values.AWS_VPC_CIDR != "":
        full_no_proxy_list.append(data.values.AWS_VPC_CIDR)
      end
      full_no_proxy_list.append("169.254.0.0/16")
    end
    if data.values.PROVIDER_TYPE == "azure":
      if data.values.AZURE_VNET_CIDR != "":
        full_no_proxy_list.append(data.values.AZURE_VNET_CIDR)
      end
      full_no_proxy_list.append("169.254.0.0/16")
      full_no_proxy_list.append("168.63.129.16")
    end
    full_no_proxy_list.append(data.values.SERVICE_CIDR)
    full_no_proxy_list.append(data.values.CLUSTER_CIDR)
    full_no_proxy_list.append("localhost")
    full_no_proxy_list.append("127.0.0.1")
    if data.values.TKG_IP_FAMILY in ["ipv6", "ipv4,ipv6", "ipv6,ipv4"]:
      full_no_proxy_list.append("::1")
    end
    full_no_proxy_list.append(".svc")
    full_no_proxy_list.append(".svc.cluster.local")
    populated_no_proxy = ",".join(list(set(full_no_proxy_list)))
    return populated_no_proxy
  end
  return ""
end

def validate_proxy_bypass_vsphere_host():
  if data.values.PROVIDER_TYPE == "vsphere" and not data.values.VSPHERE_INSECURE:
    no_proxy_list = []
    if data.values.TKG_NO_PROXY != "":
      no_proxy_list = data.values.TKG_NO_PROXY.split(",")
      if data.values.VSPHERE_SERVER not in no_proxy_list:
        assert.fail("unable to proxy traffic to vSphere host in security connection, either set VSPHERE_INSECURE to true or add VSPHERE_SERVER to TKG_NO_PROXY")
      end
    end
  end
end

# get_labels_map_from_string constructs a map from given string of the format "key1=label1,key2=label2"
def get_labels_map_from_string(labelString):
   labelMap = {}
   for val in regexp.replace(" ", labelString, "").split(','):
    kv = val.split('=')
    if len(kv) != 2:
      assert.fail("given labels string \""+labelString+"\" must be in the  \"key1=label1,key2=label2\" format ")
    end
    labelMap.update({kv[0]: kv[1]})
   end
   return labelMap
end

def compare_semver_versions(a, b):
  a_array = regexp.replace("v?(\d+\.\d+\.\d+).*", a, "$1").split(".")
  b_array = regexp.replace("v?(\d+\.\d+\.\d+).*", b, "$1").split(".")
  for i in range(len(a_array)):
    if int(a_array[i]) > int(b_array[i]):
      return 1
    elif int(a_array[i]) < int(b_array[i]):
      return -1
    end
  end
  return 0
end

def enable_csi_driver():
  tkrVersion = get_tkr_version_from_tkr_name(data.values.KUBERNETES_RELEASE)
  if compare_semver_versions(tkrVersion, "v1.23.0") >= 0:
     return True
  end
  return False
end

def map(f, list):
    return [f(x) for x in list]
end