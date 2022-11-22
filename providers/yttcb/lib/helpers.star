load("@ytt:data", "data")
load("@ytt:assert", "assert")
load("@ytt:regexp", "regexp")

TKGSProductName = "VMware Tanzu Kubernetes Grid Service for vSphere"
TKGProductName = "VMware Tanzu Kubernetes Grid"
# kapp-controller requires the data values for addons to have this header to indicate its a data value
ValuesFormatStr = "#@data/values\n#@overlay/match-child-defaults missing_ok=True\n---\n{}"

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
