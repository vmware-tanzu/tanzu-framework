#! This file implements helper function needed to create kapp-controller-values file.
#! Note: This is file implements/copies function implemented in ytt/lib/helpers.star file
#! Any changes affecting this helper functions need to be done at both the places

load("@ytt:data", "data")
load("@ytt:regexp", "regexp")

def tkg_image_repo_customized():
    return data.values.TKG_CUSTOM_IMAGE_REPOSITORY != ""
end

def tkg_image_repo_skip_tls_verify():
  return data.values.TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY == True and tkg_image_repo_customized()
end

def tkg_image_repo_ca_cert():
  return data.values.TKG_PROXY_CA_CERT if data.values.TKG_PROXY_CA_CERT else data.values.TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE
end

def tkg_custom_image_repo_hostname():
  return data.values.TKG_CUSTOM_IMAGE_REPOSITORY.split("/")[0]
end

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
