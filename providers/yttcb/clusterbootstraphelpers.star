load("@ytt:data", "data")
load("@ytt:regexp", "regexp")

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

#! Takes in something that looks like "prompt=consent,something-else=other-value"
#! and returns something that looks like [{name: prompt, value: consent},{name: something-else, value: other-value}]
#! This is purposely intended to be forgiving of additional whitespace.
def get_additional_authorize_params():
  result = []

  if data.values.OIDC_IDENTITY_PROVIDER_ADDITIONAL_AUTHORIZE_PARAMS:
    for pair in data.values.OIDC_IDENTITY_PROVIDER_ADDITIONAL_AUTHORIZE_PARAMS.split(","):
      pair = pair.lstrip().rstrip()
      elements = pair.split("=")

      if len(elements) == 2:
        result.append(dict(name=elements[0].lstrip().rstrip(), value=elements[1].lstrip().rstrip()))
      end
    end
  end

  return result
end