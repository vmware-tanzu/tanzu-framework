load("@ytt:data", "data")
load("@ytt:overlay", "overlay")
load("@ytt:yaml", "yaml")
load("/lib/helpers.star", "get_default_tkg_bom_data")

#! This file contains function 'config_variable_association' which specifies all configuration variables
#! mentioned in 'config_default.yaml' and describes association of each configuration variable with
#! the corresponding infrastructure providers.
#! Please make sure to add config variable to this file if you add new config variable to 'config_default.yaml'

def config_variable_association():

return {
"CLUSTER_NAME": ["vsphere", "aws", "azure", "tkg-service-vsphere", "docker"],
"CLUSTER_PLAN": ["vsphere", "aws", "azure", "tkg-service-vsphere", "docker"],
"NAMESPACE": ["vsphere", "aws", "azure", "tkg-service-vsphere", "docker"],
"INFRASTRUCTURE_PROVIDER": ["vsphere", "aws", "azure", "docker"],
"IS_WINDOWS_WORKLOAD_CLUSTER": ["vsphere"],

"CLUSTER_API_SERVER_PORT": ["vsphere", "aws", "azure"],

"SIZE": ["vsphere", "aws", "azure", "docker"],
"CONTROLPLANE_SIZE": ["vsphere", "aws", "azure", "docker"],
"WORKER_SIZE": ["vsphere", "aws", "azure", "docker"],

"ENABLE_CEIP_PARTICIPATION": ["vsphere", "aws", "azure", "tkg-service-vsphere", "docker"],
"DEPLOY_TKG_ON_VSPHERE7": ["vsphere"],
"ENABLE_TKGS_ON_VSPHERE7": ["vsphere"],

"VSPHERE_NUM_CPUS": ["vsphere"],
"VSPHERE_DISK_GIB": ["vsphere"],
"VSPHERE_MEM_MIB": ["vsphere"],
"VSPHERE_CONTROL_PLANE_NUM_CPUS": ["vsphere"],
"VSPHERE_CONTROL_PLANE_DISK_GIB": ["vsphere"],
"VSPHERE_CONTROL_PLANE_MEM_MIB": ["vsphere"],
"VSPHERE_WORKER_NUM_CPUS": ["vsphere"],
"VSPHERE_WORKER_DISK_GIB": ["vsphere"],
"VSPHERE_WORKER_MEM_MIB": ["vsphere"],
"VSPHERE_CLONE_MODE": ["vsphere"],
"VSPHERE_NETWORK": ["vsphere"],
"VSPHERE_TEMPLATE": ["vsphere"],
"VSPHERE_WINDOWS_TEMPLATE": ["windows-vsphere"],
"VIP_NETWORK_INTERFACE": ["vsphere"],
"VSPHERE_SSH_AUTHORIZED_KEY": ["vsphere"],
"VSPHERE_USERNAME": ["vsphere"],
"VSPHERE_PASSWORD": ["vsphere"],
"VSPHERE_REGION": ["vsphere"],
"VSPHERE_ZONE": ["vsphere"],
"VSPHERE_SERVER": ["vsphere"],
"VSPHERE_DATACENTER": ["vsphere"],
"VSPHERE_RESOURCE_POOL": ["vsphere"],
"VSPHERE_DATASTORE": ["vsphere"],
"VSPHERE_FOLDER": ["vsphere"],
"VSPHERE_STORAGE_POLICY_ID": ["vsphere"],
"VSPHERE_TLS_THUMBPRINT": ["vsphere"],
"VSPHERE_INSECURE": ["vsphere"],
"VSPHERE_CONTROL_PLANE_ENDPOINT": ["vsphere"],
"VSPHERE_CONTROL_PLANE_ENDPOINT_PORT": ["vsphere"],

"NSXT_POD_ROUTING_ENABLED": ["vsphere"],
"NSXT_ROUTER_PATH": ["vsphere"],
"NSXT_USERNAME": ["vsphere"],
"NSXT_PASSWORD": ["vsphere"],
"NSXT_MANAGER_HOST": ["vsphere"],
"NSXT_ALLOW_UNVERIFIED_SSL": ["vsphere"],
"NSXT_REMOTE_AUTH": ["vsphere"],
"NSXT_VMC_ACCESS_TOKEN": ["vsphere"],
"NSXT_VMC_AUTH_HOST": ["vsphere"],
"NSXT_CLIENT_CERT_KEY_DATA": ["vsphere"],
"NSXT_CLIENT_CERT_DATA": ["vsphere"],
"NSXT_ROOT_CA_DATA_B64": ["vsphere"],
"NSXT_SECRET_NAME": ["vsphere"],
"NSXT_SECRET_NAMESPACE": ["vsphere"],

"AWS_REGION": ["aws"],
"AWS_NODE_AZ": ["aws"],
"AWS_NODE_AZ_1": ["aws"],
"AWS_NODE_AZ_2": ["aws"],
"AWS_VPC_ID": ["aws"],
"AWS_PRIVATE_SUBNET_ID": ["aws"],
"AWS_PUBLIC_SUBNET_ID": ["aws"],
"AWS_PUBLIC_SUBNET_ID_1": ["aws"],
"AWS_PRIVATE_SUBNET_ID_1": ["aws"],
"AWS_PUBLIC_SUBNET_ID_2": ["aws"],
"AWS_PRIVATE_SUBNET_ID_2": ["aws"],
"AWS_VPC_CIDR": ["aws"],
"AWS_PRIVATE_NODE_CIDR": ["aws"],
"AWS_PUBLIC_NODE_CIDR": ["aws"],
"AWS_PRIVATE_NODE_CIDR_1": ["aws"],
"AWS_PUBLIC_NODE_CIDR_1": ["aws"],
"AWS_PRIVATE_NODE_CIDR_2": ["aws"],
"AWS_PUBLIC_NODE_CIDR_2": ["aws"],
"AWS_SECURITY_GROUP_APISERVER_LB": ["aws"],
"AWS_SECURITY_GROUP_BASTION": ["aws"],
"AWS_SECURITY_GROUP_CONTROLPLANE": ["aws"],
"AWS_SECURITY_GROUP_LB": ["aws"],
"AWS_SECURITY_GROUP_NODE": ["aws"],
"AWS_IDENTITY_REF_KIND": ["aws"],
"AWS_IDENTITY_REF_NAME": ["aws"],
"AWS_CONTROL_PLANE_OS_DISK_SIZE_GIB": ["aws"],
"AWS_NODE_OS_DISK_SIZE_GIB": ["aws"],
"CONTROL_PLANE_MACHINE_TYPE": ["aws"],
"NODE_MACHINE_TYPE": ["aws"],
"NODE_MACHINE_TYPE_1": ["aws"],
"NODE_MACHINE_TYPE_2": ["aws"],
"AWS_SSH_KEY_NAME": ["aws"],
"BASTION_HOST_ENABLED": ["aws"],
"AWS_LOAD_BALANCER_SCHEME_INTERNAL": ["aws"],

"CONTROL_PLANE_STORAGE_CLASS": ["tkg-service-vsphere"],
"CONTROL_PLANE_VM_CLASS": ["tkg-service-vsphere"],
"DEFAULT_STORAGE_CLASS": ["tkg-service-vsphere"],
"SERVICE_DOMAIN": ["tkg-service-vsphere"],
"STORAGE_CLASSES": ["tkg-service-vsphere"],
"WORKER_STORAGE_CLASS": ["tkg-service-vsphere"],
"WORKER_VM_CLASS": ["tkg-service-vsphere"],
"KUBERNETES_RELEASE": ["tkg-service-vsphere"],
"NODE_POOL_0_NAME": ["tkg-service-vsphere"],
"NODE_POOL_0_LABELS": ["tkg-service-vsphere"],
"NODE_POOL_0_TAINTS": ["tkg-service-vsphere"],

"AZURE_ENVIRONMENT": ["azure"],
"AZURE_TENANT_ID": ["azure"],
"AZURE_SUBSCRIPTION_ID": ["azure"],
"AZURE_CLIENT_ID": ["azure"],
"AZURE_CLIENT_SECRET": ["azure"],
"AZURE_LOCATION": ["azure"],
"AZURE_SSH_PUBLIC_KEY_B64": ["azure"],
"AZURE_CONTROL_PLANE_MACHINE_TYPE": ["azure"],
"AZURE_NODE_MACHINE_TYPE": ["azure"],
"AZURE_ENABLE_ACCELERATED_NETWORKING": ["azure"],
"AZURE_RESOURCE_GROUP": ["azure"],
"AZURE_VNET_RESOURCE_GROUP": ["azure"],
"AZURE_VNET_NAME": ["azure"],
"AZURE_VNET_CIDR": ["azure"],
"AZURE_CONTROL_PLANE_SUBNET_NAME": ["azure"],
"AZURE_CONTROL_PLANE_SUBNET_CIDR": ["azure"],
"AZURE_CONTROL_PLANE_SUBNET_SECURITY_GROUP": ["azure"],
"AZURE_NODE_SUBNET_NAME": ["azure"],
"AZURE_NODE_SUBNET_CIDR": ["azure"],
"AZURE_NODE_SUBNET_SECURITY_GROUP": ["azure"],
"AZURE_NODE_AZ": ["azure"],
"AZURE_NODE_AZ_1": ["azure"],
"AZURE_NODE_AZ_2": ["azure"],
"AZURE_CUSTOM_TAGS": ["azure"],
"AZURE_CONTROL_PLANE_OS_DISK_SIZE_GIB": ["azure"],
"AZURE_CONTROL_PLANE_OS_DISK_STORAGE_ACCOUNT_TYPE": ["azure"],
"AZURE_NODE_OS_DISK_SIZE_GIB": ["azure"],
"AZURE_NODE_OS_DISK_STORAGE_ACCOUNT_TYPE": ["azure"],
"AZURE_CONTROL_PLANE_DATA_DISK_SIZE_GIB": ["azure"],
"AZURE_ENABLE_NODE_DATA_DISK": ["azure"],
"AZURE_NODE_DATA_DISK_SIZE_GIB": ["azure"],
"AZURE_ENABLE_PRIVATE_CLUSTER": ["azure"],
"AZURE_FRONTEND_PRIVATE_IP": ["azure"],
"AZURE_IMAGE_ID": ["azure"],
"AZURE_IMAGE_RESOURCE_GROUP": ["azure"],
"AZURE_IMAGE_NAME": ["azure"],
"AZURE_IMAGE_SUBSCRIPTION_ID": ["azure"],
"AZURE_IMAGE_GALLERY": ["azure"],
"AZURE_IMAGE_PUBLISHER": ["azure"],
"AZURE_IMAGE_OFFER": ["azure"],
"AZURE_IMAGE_SKU": ["azure"],
"AZURE_IMAGE_THIRD_PARTY": ["azure"],
"AZURE_IMAGE_VERSION": ["azure"],
"AZURE_IDENTITY_NAME": ["azure"],
"AZURE_IDENTITY_NAMESPACE": ["azure"],
"AZURE_CONTROL_PLANE_OUTBOUND_LB_FRONTEND_IP_COUNT": ["azure"],
"AZURE_ENABLE_CONTROL_PLANE_OUTBOUND_LB": ["azure"],
"AZURE_NODE_OUTBOUND_LB_FRONTEND_IP_COUNT": ["azure"],
"AZURE_ENABLE_NODE_OUTBOUND_LB": ["azure"],
"AZURE_NODE_OUTBOUND_LB_IDLE_TIMEOUT_IN_MINUTES": ["azure"],
"ENABLE_OIDC": ["vsphere", "aws", "azure", "docker"],
"OIDC_ISSUER_URL": ["vsphere", "aws", "azure", "docker"],
"OIDC_USERNAME_CLAIM": ["vsphere", "aws", "azure", "docker"],
"OIDC_GROUPS_CLAIM": ["vsphere", "aws", "azure", "docker"],
"OIDC_DEX_CA": ["vsphere", "aws", "azure", "docker"],

"ENABLE_MHC": ["vsphere", "aws", "azure", "docker"],
"ENABLE_MHC_WORKER_NODE": ["vsphere", "aws", "azure", "docker"],
"ENABLE_MHC_CONTROL_PLANE": ["vsphere", "aws", "azure", "docker"],
"MHC_UNKNOWN_STATUS_TIMEOUT": ["vsphere", "aws", "azure", "docker"],
"MHC_FALSE_STATUS_TIMEOUT": ["vsphere", "aws", "azure", "docker"],

"TKG_CUSTOM_IMAGE_REPOSITORY": ["vsphere", "aws", "azure", "docker"],
"TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY": ["vsphere", "aws", "azure", "docker"],
"TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE": ["vsphere", "aws", "azure", "docker"],

"TKG_HTTP_PROXY": ["vsphere", "aws", "azure", "docker"],
"TKG_HTTPS_PROXY": ["vsphere", "aws", "azure", "docker"],
"TKG_NO_PROXY": ["vsphere", "aws", "azure", "docker"],
"TKG_PROXY_CA_CERT": ["vsphere", "aws", "azure", "docker"],

"TKG_IP_FAMILY": ["vsphere", "aws", "azure", "docker"],

"ENABLE_AUDIT_LOGGING": ["vsphere", "aws", "azure", "docker"],

"ENABLE_DEFAULT_STORAGE_CLASS": ["vsphere", "aws", "azure", "docker"],

"CLUSTER_CIDR": ["vsphere", "aws", "azure", "tkg-service-vsphere", "docker"],
"SERVICE_CIDR": ["vsphere", "aws", "azure", "tkg-service-vsphere", "docker"],
"NODE_STARTUP_TIMEOUT": ["vsphere", "aws", "azure", "docker"],

"CONTROL_PLANE_MACHINE_COUNT": ["vsphere", "aws", "azure", "tkg-service-vsphere", "docker"],
"WORKER_MACHINE_COUNT": ["vsphere", "aws", "azure", "tkg-service-vsphere", "docker"],
"WORKER_MACHINE_COUNT_0": ["vsphere", "aws", "azure", "docker"],
"WORKER_MACHINE_COUNT_1": ["vsphere", "aws", "azure", "docker"],
"WORKER_MACHINE_COUNT_2": ["vsphere", "aws", "azure", "docker"],

"OS_NAME": ["vsphere", "aws", "azure", "docker"],
"OS_VERSION": ["vsphere", "aws", "azure", "docker"],
"OS_ARCH": ["vsphere", "aws", "azure", "docker"],

"ENABLE_AUTOSCALER": ["vsphere", "aws", "azure", "docker"],
"AUTOSCALER_MAX_NODES_TOTAL": ["vsphere", "aws", "azure", "docker"],
"AUTOSCALER_SCALE_DOWN_DELAY_AFTER_ADD": ["vsphere", "aws", "azure", "docker"],
"AUTOSCALER_SCALE_DOWN_DELAY_AFTER_DELETE": ["vsphere", "aws", "azure", "docker"],
"AUTOSCALER_SCALE_DOWN_DELAY_AFTER_FAILURE": ["vsphere", "aws", "azure", "docker"],
"AUTOSCALER_SCALE_DOWN_UNNEEDED_TIME": ["vsphere", "aws", "azure", "docker"],
"AUTOSCALER_MAX_NODE_PROVISION_TIME": ["vsphere", "aws", "azure", "docker"],
"AUTOSCALER_MIN_SIZE_0": ["vsphere", "aws", "azure", "docker"],
"AUTOSCALER_MAX_SIZE_0": ["vsphere", "aws", "azure", "docker"],
"AUTOSCALER_MIN_SIZE_1": ["vsphere", "aws", "azure", "docker"],
"AUTOSCALER_MAX_SIZE_1": ["vsphere", "aws", "azure", "docker"],
"AUTOSCALER_MIN_SIZE_2": ["vsphere", "aws", "azure", "docker"],
"AUTOSCALER_MAX_SIZE_2": ["vsphere", "aws", "azure", "docker"],

"DOCKER_MACHINE_TEMPLATE_IMAGE": ["docker"],

"IDENTITY_MANAGEMENT_TYPE": ["vsphere", "aws", "azure", "docker"],
"CERT_DURATION": ["vsphere", "aws", "azure", "docker"],
"CERT_RENEW_BEFORE": ["vsphere", "aws", "azure", "docker"],
"OIDC_IDENTITY_PROVIDER_NAME": ["vsphere", "aws", "azure", "docker"],
"OIDC_IDENTITY_PROVIDER_ISSUER_URL": ["vsphere", "aws", "azure", "docker"],
"OIDC_IDENTITY_PROVIDER_CLIENT_ID": ["vsphere", "aws", "azure", "docker"],
"OIDC_IDENTITY_PROVIDER_CLIENT_SECRET": ["vsphere", "aws", "azure", "docker"],
"OIDC_IDENTITY_PROVIDER_SCOPES": ["vsphere", "aws", "azure", "docker"],
"OIDC_IDENTITY_PROVIDER_USERNAME_CLAIM": ["vsphere", "aws", "azure", "docker"],
"OIDC_IDENTITY_PROVIDER_GROUPS_CLAIM": ["vsphere", "aws", "azure", "docker"],
"SUPERVISOR_ISSUER_URL": ["vsphere", "aws", "azure", "docker"],
"SUPERVISOR_ISSUER_CA_BUNDLE_DATA_B64": ["vsphere", "aws", "azure", "docker"],
"LDAP_BIND_DN": ["vsphere", "aws", "azure", "docker"],
"LDAP_BIND_PASSWORD": ["vsphere", "aws", "azure", "docker"],
"LDAP_HOST": ["vsphere", "aws", "azure", "docker"],
"LDAP_USER_SEARCH_BASE_DN": ["vsphere", "aws", "azure", "docker"],
"LDAP_USER_SEARCH_FILTER": ["vsphere", "aws", "azure", "docker"],
"LDAP_USER_SEARCH_USERNAME": ["vsphere", "aws", "azure", "docker"],
"LDAP_USER_SEARCH_ID_ATTRIBUTE": ["vsphere", "aws", "azure", "docker"],
"LDAP_USER_SEARCH_EMAIL_ATTRIBUTE": ["vsphere", "aws", "azure", "docker"],
"LDAP_USER_SEARCH_NAME_ATTRIBUTE": ["vsphere", "aws", "azure", "docker"],
"LDAP_GROUP_SEARCH_BASE_DN": ["vsphere", "aws", "azure", "docker"],
"LDAP_GROUP_SEARCH_FILTER": ["vsphere", "aws", "azure", "docker"],
"LDAP_GROUP_SEARCH_USER_ATTRIBUTE": ["vsphere", "aws", "azure", "docker"],
"LDAP_GROUP_SEARCH_GROUP_ATTRIBUTE": ["vsphere", "aws", "azure", "docker"],
"LDAP_GROUP_SEARCH_NAME_ATTRIBUTE": ["vsphere", "aws", "azure", "docker"],
"LDAP_ROOT_CA_DATA_B64": ["vsphere", "aws", "azure", "docker"],

"AVI_ENABLE": ["vsphere"],
"AVI_NAMESPACE": ["vsphere"],
"AVI_DISABLE_INGRESS_CLASS": ["vsphere"],
"AVI_AKO_IMAGE_PULL_POLICY": ["vsphere"],
"AVI_ADMIN_CREDENTIAL_NAME": ["vsphere"],
"AVI_CA_NAME": ["vsphere"],
"AVI_CONTROLLER": ["vsphere"],
"AVI_USERNAME": ["vsphere"],
"AVI_PASSWORD": ["vsphere"],
"AVI_CLOUD_NAME": ["vsphere"],
"AVI_SERVICE_ENGINE_GROUP": ["vsphere"],
"AVI_DATA_NETWORK": ["vsphere"],
"AVI_DATA_NETWORK_CIDR": ["vsphere"],
"AVI_MANAGEMENT_CLUSTER_VIP_NETWORK_NAME": ["vsphere"],
"AVI_MANAGEMENT_CLUSTER_VIP_NETWORK_CIDR": ["vsphere"],
"AVI_CA_DATA_B64": ["vsphere"],
"AVI_LABELS": ["vsphere"],
"AVI_DISABLE_STATIC_ROUTE_SYNC": ["vsphere"],
"AVI_INGRESS_DEFAULT_INGRESS_CONTROLLER": ["vsphere"],
"AVI_INGRESS_SHARD_VS_SIZE": ["vsphere"],
"AVI_INGRESS_SERVICE_TYPE": ["vsphere"],
"AVI_INGRESS_NODE_NETWORK_LIST": ["vsphere"],
"AVI_CONTROL_PLANE_HA_PROVIDER": ["vsphere"],

"ANTREA_NO_SNAT": ["vsphere", "aws", "azure", "docker"],
"ANTREA_TRAFFIC_ENCAP_MODE": ["vsphere", "aws", "azure", "docker"],
"ANTREA_PROXY": ["vsphere", "aws", "azure", "docker"],
"ANTREA_ENDPOINTSLICE": ["vsphere", "aws", "azure", "docker"],
"ANTREA_POLICY": ["vsphere", "aws", "azure", "docker"],
"ANTREA_NODEPORTLOCAL": ["vsphere", "aws", "azure", "docker"],
"ANTREA_TRACEFLOW": ["vsphere", "aws", "azure", "docker"],
"ANTREA_DISABLE_UDP_TUNNEL_OFFLOAD": ["vsphere", "aws", "azure", "docker"],

"PROVIDER_TYPE": ["vsphere", "aws", "azure", "tkg-service-vsphere", "docker"],
"TKG_CLUSTER_ROLE": ["vsphere", "aws", "azure", "tkg-service-vsphere", "docker"],
"TKG_VERSION": ["vsphere", "aws", "azure", "tkg-service-vsphere", "docker"],
"CNI": ["vsphere", "aws", "azure", "docker"],
"VSPHERE_VERSION": ["vsphere"],
}

end

def get_cluster_variables():
    vars = {}
    kvs = config_variable_association()
    network = {}
    for configVariable in kvs:
        if data.values.PROVIDER_TYPE in kvs[configVariable]:
            if data.values[configVariable] != None:
                vars[configVariable] = data.values[configVariable]
            else:
                continue
            end
            if configVariable == "TKG_HTTP_PROXY":
                if vars["TKG_HTTP_PROXY"] != "":
                    network["proxy"] = {
                        "httpProxy": vars["TKG_HTTP_PROXY"],
                        "httpsProxy": data.values["TKG_HTTPS_PROXY"],
                        "noProxy": data.values["TKG_NO_PROXY"].split(","),
                    }
                end
            end
        end
    end

    if data.values["TKG_IP_FAMILY"] == "ipv6,ipv4":
        network["ipv6Primary"] = True
    end

    if network != {}:
        vars["network"] = network
    end

    vars["cni"] = data.values["CNI"]

    customImageRepository = {}
    if data.values["TKG_CUSTOM_IMAGE_REPOSITORY"] != "":
        customImageRepository["host"] = data.values["TKG_CUSTOM_IMAGE_REPOSITORY"]
    end
    if data.values["TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY"] != False:
        customImageRepository["tlsCertificateValidation"] = not data.values["TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY"]
    end

    if customImageRepository != {}:
        vars["imageRepository"] = customImageRepository
    end

    if data.values["TKG_CLUSTER_ROLE"] != "":
        vars["clusterRole"] = data.values["TKG_CLUSTER_ROLE"]
    end

    if data.values["ENABLE_AUDIT_LOGGING"] != "":
        vars["auditLogging"] = {
            "enabled": data.values["ENABLE_AUDIT_LOGGING"]
        }
    end

    trust = []
    if data.values["TKG_PROXY_CA_CERT"] != "":
        trust.append({
            "name": "proxy",
            "data": data.values["TKG_PROXY_CA_CERT"]
        })
    end
    if data.values["TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE"] != "":
        trust.append({
            "name": "imageRepository",
            "data": data.values["TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE"]
        })
    end
    if len(trust) > 0:
        vars["trust"] = trust
    end

    if data.values["CLUSTER_API_SERVER_PORT"] != None:
        vars["apiServerPort"] = data.values["CLUSTER_API_SERVER_PORT"]
    end

    if data.values["TKR_DATA"] != "":
        vars["TKR_DATA"] = data.values["TKR_DATA"]
    end

    return vars
end

def get_aws_vars():
    simpleMapping = {}
    simpleMapping["AWS_REGION"] = "region"
    simpleMapping["AWS_SSH_KEY_NAME"] = "sshKeyName"
    simpleMapping["AWS_LOAD_BALANCER_SCHEME_INTERNAL"] = "loadBalancerSchemeInternal"
    vars = get_cluster_variables()

    for key in simpleMapping:
        if data.values[key] != None:
            vars[simpleMapping[key]] = data.values[key]
        end
    end


    vars["bastion"] = {
        "enabled": data.values["BASTION_HOST_ENABLED"] 
    }

    if vars.get("network") == None:
        vars["network"] = {}
    end

    identityRef = {}
    if data.values["AWS_IDENTITY_REF_NAME"] != "":
        identityRef["name"] = data.values["AWS_IDENTITY_REF_NAME"]
    end
    if data.values["AWS_IDENTITY_REF_KIND"] != "":
        identityRef["kind"] = data.values["AWS_IDENTITY_REF_KIND"]
    end
    vars["identityRef"] = identityRef

    vpc = {}
    if data.values["AWS_VPC_ID"]:
        vpc["existingID"] = data.values["AWS_VPC_ID"]
    end
    if data.values["AWS_VPC_CIDR"]:
        vpc["cidr"] = data.values["AWS_VPC_CIDR"]
    end
    vars["network"]["vpc"] = vpc

    securityGroup = {}
    if data.values["AWS_SECURITY_GROUP_BASTION"] != "":
        securityGroup["bastion"] = data.values["AWS_SECURITY_GROUP_BASTION"]
    end
    if data.values["AWS_SECURITY_GROUP_APISERVER_LB"] != "":
        securityGroup["apiServerLB"] = data.values["AWS_SECURITY_GROUP_APISERVER_LB"]
    end
    if data.values["AWS_SECURITY_GROUP_LB"] != "":
        securityGroup["lb"] = data.values["AWS_SECURITY_GROUP_LB"]
    end
    if data.values["AWS_SECURITY_GROUP_CONTROLPLANE"] != "":
        securityGroup["controlPlane"] = data.values["AWS_SECURITY_GROUP_CONTROLPLANE"]
    end
    if data.values["AWS_SECURITY_GROUP_NODE"] != "":
        securityGroup["node"] = data.values["AWS_SECURITY_GROUP_NODE"]
    end

    if securityGroup != {}:
        vars["network"]["securityGroupOverrides"] = securityGroup
    end

    subnets = []

    private0 = {}
    if data.values["AWS_PRIVATE_NODE_CIDR"] != "":
        private0["cidr"] = data.values["AWS_PRIVATE_NODE_CIDR"]
    end
    if data.values["AWS_PRIVATE_SUBNET_ID"] != "":
        private0["id"] = data.values["AWS_PRIVATE_SUBNET_ID"]
    end
    public0 = {}
    if data.values["AWS_PUBLIC_NODE_CIDR"] != "":
        public0["cidr"] = data.values["AWS_PUBLIC_NODE_CIDR"]
    end
    if data.values["AWS_PUBLIC_SUBNET_ID"] != "":
        public0["id"] = data.values["AWS_PUBLIC_SUBNET_ID"]
    end

    subnet0 = {}
    if private0 != {}:
        subnet0["private"] = private0
    end
    if public0 != {}:
        subnet0["public"] = public0
    end
    if data.values["AWS_NODE_AZ"] != None:
        subnet0["az"] = data.values["AWS_NODE_AZ"]
    end
    if subnet0 != {}:
        subnets.append(subnet0)
    end

    if data.values["CLUSTER_PLAN"] == "prodcc":
        private1 = {}
        if data.values["AWS_PRIVATE_NODE_CIDR_1"] != "":
            private1["cidr"] = data.values["AWS_PRIVATE_NODE_CIDR_1"]
        end
        if data.values["AWS_PRIVATE_SUBNET_ID_1"] != "":
            private1["id"] = data.values["AWS_PRIVATE_SUBNET_ID_1"]
        end
        public1 = {}
        if data.values["AWS_PUBLIC_NODE_CIDR_1"] != "":
            public1["cidr"] = data.values["AWS_PUBLIC_NODE_CIDR_1"]
        end
        if data.values["AWS_PUBLIC_SUBNET_ID_1"] != "":
            public1["id"] = data.values["AWS_PUBLIC_SUBNET_ID_1"]
        end

        subnet1 = {}
        if private1 != {}:
            subnet1["private"] = private1
        end
        if public1 != {}:
            subnet1["public"] = public1
        end
        if data.values["AWS_NODE_AZ_1"] != "":
            subnet1["az"] = data.values["AWS_NODE_AZ_1"]
        end

        if subnet1 != {}:
            subnets.append(subnet1)
        end

        private2 = {}
        if data.values["AWS_PRIVATE_NODE_CIDR_2"] != "":
            private2["cidr"] = data.values["AWS_PRIVATE_NODE_CIDR_2"]
        end
        if data.values["AWS_PRIVATE_SUBNET_ID_2"] != "":
            private2["id"] = data.values["AWS_PRIVATE_SUBNET_ID_2"]
        end
        public2 = {}
        if data.values["AWS_PUBLIC_NODE_CIDR_2"] != "":
            public2["cidr"] = data.values["AWS_PUBLIC_NODE_CIDR_2"]
        end
        if data.values["AWS_PUBLIC_SUBNET_ID_2"] != "":
            public2["id"] = data.values["AWS_PUBLIC_SUBNET_ID_2"]
        end

        subnet2 = {}
        if private2 != {}:
            subnet2["private"] = private2
        end
        if public2 != {}:
            subnet2["public"] = public2
        end
        if data.values["AWS_NODE_AZ_2"] != "":
            subnet2["az"] = data.values["AWS_NODE_AZ_2"]
        end

        if subnet2 != {}:
            subnets.append(subnet2)
        end
    end

    vars["network"]["subnets"] = subnets

    worker = {}
    if data.values["NODE_MACHINE_TYPE"] != None:
        worker["instanceType"] = data.values["NODE_MACHINE_TYPE"]
    end
    if data.values["AWS_NODE_OS_DISK_SIZE_GIB"] != None:
        rootVolume = {}
        rootVolume["sizeGiB"] = data.values["AWS_NODE_OS_DISK_SIZE_GIB"]
        worker["rootVolume"] = rootVolume
    end

    vars["worker"] = worker

    controlPlane = {}
    if data.values["CONTROL_PLANE_MACHINE_TYPE"] != None:
        controlPlane["instanceType"] = data.values["CONTROL_PLANE_MACHINE_TYPE"]
    end
    if data.values["AWS_CONTROL_PLANE_OS_DISK_SIZE_GIB"] != None:
        rootVolume = {}
        rootVolume["sizeGiB"] = data.values["AWS_CONTROL_PLANE_OS_DISK_SIZE_GIB"]
        controlPlane["rootVolume"] = rootVolume
    end

    vars["controlPlane"] = controlPlane

    return vars
end

def get_azure_vars():
    simpleMapping = {}
    simpleMapping["AZURE_LOCATION"] = "location"
    simpleMapping["AZURE_RESOURCE_GROUP"] = "resourceGroup"
    simpleMapping["AZURE_SUBSCRIPTION_ID"] = "subscriptionID"
    simpleMapping["AZURE_ENVIRONMENT"] = "environment"
    simpleMapping["AZURE_SSH_PUBLIC_KEY_B64"] = "sshPublicKey"
    simpleMapping["AZURE_FRONTEND_PRIVATE_IP"] = "frontendPrivateIP"
    simpleMapping["AZURE_CUSTOM_TAGS"] = "customTags"
    vars = get_cluster_variables()

    for key in simpleMapping:
        if data.values[key] != None:
            vars[simpleMapping[key]] = data.values[key]
        end
    end

    if data.values["AZURE_ENABLE_ACCELERATED_NETWORKING"] != "":
        vars["acceleratedNetworking"] =  {
            "enabled": data.values["AZURE_ENABLE_ACCELERATED_NETWORKING"]
        }
    end
    if data.values["AZURE_ENABLE_PRIVATE_CLUSTER"] != "":
        vars["privateCluster"] = {
            "enabled": data.values["AZURE_ENABLE_PRIVATE_CLUSTER"]
        }
    end

    if vars.get("network") == None:
        vars["network"] = {}
    end

    vnet = {}
    if data.values["AZURE_VNET_NAME"] != "":
        vnet["name"] = data.values["AZURE_VNET_NAME"]
    end
    if data.values["AZURE_VNET_CIDR"] != "":
        vnet["cidrBlocks"] = [
            data.values["AZURE_VNET_CIDR"]
        ]
    end
    if data.values["AZURE_VNET_RESOURCE_GROUP"] != "":
        vnet["resourceGroup"] = data.values["AZURE_VNET_RESOURCE_GROUP"]
    end
    if vnet != {}:
        vars["network"]["vnet"] = vnet
    end

    identity = {}
    if data.values["AZURE_IDENTITY_NAME"] != "":
        identity["name"] = data.values["AZURE_IDENTITY_NAME"]
    end
    if data.values["AZURE_IDENTITY_NAMESPACE"] != "":
        identity["namespace"] = data.values["AZURE_IDENTITY_NAMESPACE"]
    end
    if identity != {}:
        vars["identityRef"] = identity
    end

    controlPlane = {}
    if data.values["AZURE_CONTROL_PLANE_MACHINE_TYPE"] != "":
        controlPlane["vmSize"] = data.values["AZURE_CONTROL_PLANE_MACHINE_TYPE"]
    end

    osDisk = {}
    if data.values["AZURE_CONTROL_PLANE_OS_DISK_STORAGE_ACCOUNT_TYPE"] != "":
        osDisk["storageAccountType"] = data.values["AZURE_CONTROL_PLANE_OS_DISK_STORAGE_ACCOUNT_TYPE"]
    end
    if data.values["AZURE_CONTROL_PLANE_OS_DISK_SIZE_GIB"] != "":
        osDisk["sizeGiB"] = data.values["AZURE_CONTROL_PLANE_OS_DISK_SIZE_GIB"]
    end
    if osDisk != {}:
        controlPlane["osDisk"] = osDisk
    end

    dataDisk = {}
    if data.values["AZURE_CONTROL_PLANE_DATA_DISK_SIZE_GIB"] != "":
        dataDisk["sizeGiB"] = data.values["AZURE_CONTROL_PLANE_DATA_DISK_SIZE_GIB"]
        controlPlane["dataDisks"] = [dataDisk]
    end

    subnet = {}
    if data.values["AZURE_CONTROL_PLANE_SUBNET_NAME"] != "":
        subnet["name"] = data.values["AZURE_CONTROL_PLANE_SUBNET_NAME"]
    end
    if data.values["AZURE_CONTROL_PLANE_SUBNET_CIDR"] != "":
        subnet["cidr"] = data.values["AZURE_CONTROL_PLANE_SUBNET_CIDR"]
    end
    if data.values["AZURE_CONTROL_PLANE_SUBNET_SECURITY_GROUP"] != "":
        subnet["securityGroup"] = data.values["AZURE_CONTROL_PLANE_SUBNET_SECURITY_GROUP"]
    end
    if subnet != {}:
        controlPlane["subnet"] = subnet
    end

    outboundLB = {}
    if data.values["AZURE_ENABLE_CONTROL_PLANE_OUTBOUND_LB"] != False:
        outboundLB["enabled"] = data.values["AZURE_ENABLE_CONTROL_PLANE_OUTBOUND_LB"]
    end
    if data.values["AZURE_CONTROL_PLANE_OUTBOUND_LB_FRONTEND_IP_COUNT"] != "":
        outboundLB["frontendIPCount"] = data.values["AZURE_CONTROL_PLANE_OUTBOUND_LB_FRONTEND_IP_COUNT"]
    end
    if outboundLB != {}:
        controlPlane["outboundLB"] = outboundLB
    end

    if controlPlane != {}:
        vars["controlPlane"] = controlPlane
    end

    worker = {}
    if data.values["AZURE_NODE_MACHINE_TYPE"] != "":
        worker["vmSize"] = data.values["AZURE_NODE_MACHINE_TYPE"]
    end

    osDisk = {}
    if data.values["AZURE_NODE_OS_DISK_STORAGE_ACCOUNT_TYPE"] != "":
        osDisk["storageAccountType"] = data.values["AZURE_NODE_OS_DISK_STORAGE_ACCOUNT_TYPE"]
    end
    if data.values["AZURE_NODE_OS_DISK_SIZE_GIB"] != "":
        osDisk["sizeGiB"] = data.values["AZURE_NODE_OS_DISK_SIZE_GIB"]
    end
    if osDisk != {}:
        worker["osDisk"] = osDisk
    end

    dataDisk = {}
    if data.values["AZURE_ENABLE_NODE_DATA_DISK"] != False:
        worker["dataDisks"] = [dataDisk]
    end
    if data.values["AZURE_NODE_DATA_DISK_SIZE_GIB"] != "":
        dataDisk["sizeGiB"] = data.values["AZURE_NODE_DATA_DISK_SIZE_GIB"]
    end

    subnet = {}
    if data.values["AZURE_NODE_SUBNET_NAME"] != "":
        subnet["name"] = data.values["AZURE_NODE_SUBNET_NAME"]
    end
    if data.values["AZURE_NODE_SUBNET_CIDR"] != "":
        subnet["cidr"] = data.values["AZURE_NODE_SUBNET_CIDR"]
    end
    if data.values["AZURE_NODE_SUBNET_SECURITY_GROUP"] != "":
        subnet["securityGroup"] = data.values["AZURE_NODE_SUBNET_SECURITY_GROUP"]
    end
    if subnet != {}:
        worker["subnet"] = subnet
    end

    outboundLB = {}
    if data.values["AZURE_ENABLE_NODE_OUTBOUND_LB"] != False:
        outboundLB["enabled"] = data.values["AZURE_ENABLE_NODE_OUTBOUND_LB"]
    end
    if data.values["AZURE_NODE_OUTBOUND_LB_FRONTEND_IP_COUNT"] != "":
        outboundLB["frontendIPCount"] = data.values["AZURE_NODE_OUTBOUND_LB_FRONTEND_IP_COUNT"]
    end
    if data.values["AZURE_NODE_OUTBOUND_LB_IDLE_TIMEOUT_IN_MINUTES"] != "":
        outboundLB["idleTimeoutInMinutes"] = data.values["AZURE_NODE_OUTBOUND_LB_IDLE_TIMEOUT_IN_MINUTES"]
    end
    if outboundLB != {}:
        worker["outboundLB"] = outboundLB
    end

    if worker != {}:
        vars["worker"] = worker
    end

    return vars
end

def get_vsphere_vars():
    simpleMapping = {}
    simpleMapping["VSPHERE_CONTROL_PLANE_ENDPOINT"] = "apiServerEndpoint"
    simpleMapping["VIP_NETWORK_INTERFACE"] = "vipNetworkInterface"
    simpleMapping["AVI_CONTROL_PLANE_HA_PROVIDER"] = "aviAPIServerHAProvider"

    vars = get_cluster_variables()

    for key in simpleMapping:
        if data.values[key] != None:
            vars[simpleMapping[key]] = data.values[key]
        end
    end

    vcenter = {}
    vcenterMapping = {}
    vcenterMapping["VSPHERE_CLONE_MODE"] = "cloneMode"
    vcenterMapping["VSPHERE_NETWORK"] = "network"
    vcenterMapping["VSPHERE_DATACENTER"] = "datacenter"
    vcenterMapping["VSPHERE_DATASTORE"] = "datastore"
    vcenterMapping["VSPHERE_FOLDER"] = "folder"
    vcenterMapping["VSPHERE_RESOURCE_POOL"] = "resourcePool"
    vcenterMapping["VSPHERE_STORAGE_POLICY_ID"] = "storagePolicyID"
    vcenterMapping["VSPHERE_SERVER"] = "server"
    vcenterMapping["VSPHERE_TLS_THUMBPRINT"] = "tlsThumbprint"
    vcenterMapping["VSPHERE_TEMPLATE"] = "template"

    for key in vcenterMapping:
        if data.values[key] != None:
            vcenter[vcenterMapping[key]] = data.values[key]
        end
    end

    if vcenter != {}:
        vars["vcenter"] = vcenter
    end

    if data.values["VSPHERE_SSH_AUTHORIZED_KEY"] != None:
        vars["user"] = {
            "sshAuthorizedKeys": [data.values["VSPHERE_SSH_AUTHORIZED_KEY"]]
        }
    end

    controlPlane = {}
    machine = {}
    if data.values["VSPHERE_CONTROL_PLANE_NUM_CPUS"] != "":
        machine["numCPUs"] = data.values["VSPHERE_CONTROL_PLANE_NUM_CPUS"]
    end
    if data.values["VSPHERE_CONTROL_PLANE_DISK_GIB"] != "":
        machine["diskGiB"] = data.values["VSPHERE_CONTROL_PLANE_DISK_GIB"]
    end
    if data.values["VSPHERE_CONTROL_PLANE_MEM_MIB"] != "":
        machine["memoryMiB"] = data.values["VSPHERE_CONTROL_PLANE_MEM_MIB"]
    end
    if machine != {}:
        controlPlane["machine"] = machine
    end

    network = {}
    if data.values["CONTROL_PLANE_NODE_NAMESERVERS"] != None:
        network["nameservers"] = data.values["CONTROL_PLANE_NODE_NAMESERVERS"]
    end
    if network != {}:
        controlPlane["network"] = network
    end
    if controlPlane != {}:
        vars["controlPlane"] = controlPlane
    end

    worker = {}
    if data.values["WORKER_MACHINE_COUNT"] != "":
        worker["count"] = data.values["WORKER_MACHINE_COUNT"]
    end
    machine = {}
    if data.values["VSPHERE_WORKER_NUM_CPUS"] != "":
        machine["numCPUs"] = data.values["VSPHERE_WORKER_NUM_CPUS"]
    end
    if data.values["VSPHERE_WORKER_DISK_GIB"] != "":
        machine["diskGiB"] = data.values["VSPHERE_WORKER_DISK_GIB"]
    end
    if data.values["VSPHERE_WORKER_MEM_MIB"] != "":
        machine["memoryMiB"] = data.values["VSPHERE_WORKER_MEM_MIB"]
    end
    if machine != {}:
        worker["machine"] = machine
    end

    network = {}
    if data.values["WORKER_NODE_NAMESERVERS"] != None:
        network["nameservers"] = data.values["WORKER_NODE_NAMESERVERS"]
    end
    if network != {}:
        worker["network"] = network
    end
    if worker != {}:
        vars["worker"] = worker
    end

    return vars
end

def get_azure_vars():
    simpleMapping = {}
    simpleMapping["AZURE_LOCATION"] = "location"
    simpleMapping["AZURE_RESOURCE_GROUP"] = "resourceGroup"
    simpleMapping["AZURE_SUBSCRIPTION_ID"] = "subscriptionID"
    simpleMapping["AZURE_ENVIRONMENT"] = "environment"
    simpleMapping["AZURE_SSH_PUBLIC_KEY_B64"] = "sshPublicKey"
    simpleMapping["AZURE_FRONTEND_PRIVATE_IP"] = "frontendPrivateIP"
    simpleMapping["AZURE_CUSTOM_TAGS"] = "customTags"
    vars = get_cluster_variables()

    for key in simpleMapping:
        if data.values[key] != None:
            vars[simpleMapping[key]] = data.values[key]
        end
    end

    if data.values["AZURE_ENABLE_ACCELERATED_NETWORKING"] != "":
        vars["acceleratedNetworking"] =  {
            "enabled": data.values["AZURE_ENABLE_ACCELERATED_NETWORKING"]
        }
    end
    if data.values["AZURE_ENABLE_PRIVATE_CLUSTER"] != "":
        vars["privateCluster"] = {
            "enabled": data.values["AZURE_ENABLE_PRIVATE_CLUSTER"]
        }
    end

    if vars.get("network") == None:
        vars["network"] = {}
    end

    vnet = {}
    if data.values["AZURE_VNET_NAME"] != "":
        vnet["name"] = data.values["AZURE_VNET_NAME"]
    end
    if data.values["AZURE_VNET_CIDR"] != "":
        vnet["cidrBlocks"] = [
            data.values["AZURE_VNET_CIDR"]
        ]
    end
    if data.values["AZURE_VNET_RESOURCE_GROUP"] != "":
        vnet["resourceGroup"] = data.values["AZURE_VNET_RESOURCE_GROUP"]
    end
    if vnet != {}:
        vars["network"]["vnet"] = vnet
    end

    identity = {}
    if data.values["AZURE_IDENTITY_NAME"] != "":
        identity["name"] = data.values["AZURE_IDENTITY_NAME"]
    end
    if data.values["AZURE_IDENTITY_NAMESPACE"] != "":
        identity["namespace"] = data.values["AZURE_IDENTITY_NAMESPACE"]
    end
    if identity != {}:
        vars["identityRef"] = identity
    end

    controlPlane = {}
    if data.values["AZURE_CONTROL_PLANE_MACHINE_TYPE"] != "":
        controlPlane["vmSize"] = data.values["AZURE_CONTROL_PLANE_MACHINE_TYPE"]
    end

    osDisk = {}
    if data.values["AZURE_CONTROL_PLANE_OS_DISK_STORAGE_ACCOUNT_TYPE"] != "":
        osDisk["storageAccountType"] = data.values["AZURE_CONTROL_PLANE_OS_DISK_STORAGE_ACCOUNT_TYPE"]
    end
    if data.values["AZURE_CONTROL_PLANE_OS_DISK_SIZE_GIB"] != "":
        osDisk["sizeGiB"] = data.values["AZURE_CONTROL_PLANE_OS_DISK_SIZE_GIB"]
    end
    if osDisk != {}:
        controlPlane["osDisk"] = osDisk
    end

    dataDisk = {}
    if data.values["AZURE_CONTROL_PLANE_DATA_DISK_SIZE_GIB"] != "":
        dataDisk["sizeGiB"] = data.values["AZURE_CONTROL_PLANE_DATA_DISK_SIZE_GIB"]
        controlPlane["dataDisks"] = [dataDisk]
    end

    subnet = {}
    if data.values["AZURE_CONTROL_PLANE_SUBNET_NAME"] != "":
        subnet["name"] = data.values["AZURE_CONTROL_PLANE_SUBNET_NAME"]
    end
    if data.values["AZURE_CONTROL_PLANE_SUBNET_CIDR"] != "":
        subnet["cidr"] = data.values["AZURE_CONTROL_PLANE_SUBNET_CIDR"]
    end
    if data.values["AZURE_CONTROL_PLANE_SUBNET_SECURITY_GROUP"] != "":
        subnet["securityGroup"] = data.values["AZURE_CONTROL_PLANE_SUBNET_SECURITY_GROUP"]
    end
    if subnet != {}:
        controlPlane["subnet"] = subnet
    end

    outboundLB = {}
    if data.values["AZURE_ENABLE_CONTROL_PLANE_OUTBOUND_LB"] != False:
        outboundLB["enabled"] = data.values["AZURE_ENABLE_CONTROL_PLANE_OUTBOUND_LB"]
    end
    if data.values["AZURE_CONTROL_PLANE_OUTBOUND_LB_FRONTEND_IP_COUNT"] != "":
        outboundLB["frontendIPCount"] = data.values["AZURE_CONTROL_PLANE_OUTBOUND_LB_FRONTEND_IP_COUNT"]
    end
    if outboundLB != {}:
        controlPlane["outboundLB"] = outboundLB
    end

    if controlPlane != {}:
        vars["controlPlane"] = controlPlane
    end

    worker = {}
    if data.values["AZURE_NODE_MACHINE_TYPE"] != "":
        worker["vmSize"] = data.values["AZURE_NODE_MACHINE_TYPE"]
    end

    osDisk = {}
    if data.values["AZURE_NODE_OS_DISK_STORAGE_ACCOUNT_TYPE"] != "":
        osDisk["storageAccountType"] = data.values["AZURE_NODE_OS_DISK_STORAGE_ACCOUNT_TYPE"]
    end
    if data.values["AZURE_NODE_OS_DISK_SIZE_GIB"] != "":
        osDisk["sizeGiB"] = data.values["AZURE_NODE_OS_DISK_SIZE_GIB"]
    end
    if osDisk != {}:
        worker["osDisk"] = osDisk
    end

    dataDisk = {}
    if data.values["AZURE_ENABLE_NODE_DATA_DISK"] != False:
        worker["dataDisks"] = [dataDisk]
    end
    if data.values["AZURE_NODE_DATA_DISK_SIZE_GIB"] != "":
        dataDisk["sizeGiB"] = data.values["AZURE_NODE_DATA_DISK_SIZE_GIB"]
    end

    subnet = {}
    if data.values["AZURE_NODE_SUBNET_NAME"] != "":
        subnet["name"] = data.values["AZURE_NODE_SUBNET_NAME"]
    end
    if data.values["AZURE_NODE_SUBNET_CIDR"] != "":
        subnet["cidr"] = data.values["AZURE_NODE_SUBNET_CIDR"]
    end
    if data.values["AZURE_NODE_SUBNET_SECURITY_GROUP"] != "":
        subnet["securityGroup"] = data.values["AZURE_NODE_SUBNET_SECURITY_GROUP"]
    end
    if subnet != {}:
        worker["subnet"] = subnet
    end

    outboundLB = {}
    if data.values["AZURE_ENABLE_NODE_OUTBOUND_LB"] != False:
        outboundLB["enabled"] = data.values["AZURE_ENABLE_NODE_OUTBOUND_LB"]
    end
    if data.values["AZURE_NODE_OUTBOUND_LB_FRONTEND_IP_COUNT"] != "":
        outboundLB["frontendIPCount"] = data.values["AZURE_NODE_OUTBOUND_LB_FRONTEND_IP_COUNT"]
    end
    if data.values["AZURE_NODE_OUTBOUND_LB_IDLE_TIMEOUT_IN_MINUTES"] != "":
        outboundLB["idleTimeoutInMinutes"] = data.values["AZURE_NODE_OUTBOUND_LB_IDLE_TIMEOUT_IN_MINUTES"]
    end
    if outboundLB != {}:
        worker["outboundLB"] = outboundLB
    end

    if worker != {}:
        vars["worker"] = worker
    end

    return vars
end

def get_vsphere_vars():
    simpleMapping = {}
    simpleMapping["VSPHERE_CONTROL_PLANE_ENDPOINT"] = "apiServerEndpoint"
    simpleMapping["VIP_NETWORK_INTERFACE"] = "vipNetworkInterface"
    simpleMapping["AVI_CONTROL_PLANE_HA_PROVIDER"] = "aviAPIServerHAProvider"

    vars = get_cluster_variables()

    for key in simpleMapping:
        if data.values[key] != None:
            vars[simpleMapping[key]] = data.values[key]
        end
    end

    vcenter = {}
    vcenterMapping = {}
    vcenterMapping["VSPHERE_CLONE_MODE"] = "cloneMode"
    vcenterMapping["VSPHERE_NETWORK"] = "network"
    vcenterMapping["VSPHERE_DATACENTER"] = "datacenter"
    vcenterMapping["VSPHERE_DATASTORE"] = "datastore"
    vcenterMapping["VSPHERE_FOLDER"] = "folder"
    vcenterMapping["VSPHERE_RESOURCE_POOL"] = "resourcePool"
    vcenterMapping["VSPHERE_STORAGE_POLICY_ID"] = "storagePolicyID"
    vcenterMapping["VSPHERE_SERVER"] = "server"
    vcenterMapping["VSPHERE_TLS_THUMBPRINT"] = "tlsThumbprint"
    vcenterMapping["VSPHERE_TEMPLATE"] = "template"

    for key in vcenterMapping:
        if data.values[key] != None:
            vcenter[vcenterMapping[key]] = data.values[key]
        end
    end

    if vcenter != {}:
        vars["vcenter"] = vcenter
    end

    if data.values["VSPHERE_SSH_AUTHORIZED_KEY"] != None:
        vars["user"] = {
            "sshAuthorizedKeys": [data.values["VSPHERE_SSH_AUTHORIZED_KEY"]]
        }
    end

    controlPlane = {}
    machine = {}
    if data.values["VSPHERE_CONTROL_PLANE_NUM_CPUS"] != "":
        machine["numCPUs"] = data.values["VSPHERE_CONTROL_PLANE_NUM_CPUS"]
    end
    if data.values["VSPHERE_CONTROL_PLANE_DISK_GIB"] != "":
        machine["diskGiB"] = data.values["VSPHERE_CONTROL_PLANE_DISK_GIB"]
    end
    if data.values["VSPHERE_CONTROL_PLANE_MEM_MIB"] != "":
        machine["memoryMiB"] = data.values["VSPHERE_CONTROL_PLANE_MEM_MIB"]
    end
    if machine != {}:
        controlPlane["machine"] = machine
    end

    network = {}
    if data.values["CONTROL_PLANE_NODE_NAMESERVERS"] != None:
        network["nameservers"] = data.values["CONTROL_PLANE_NODE_NAMESERVERS"]
    end
    if network != {}:
        controlPlane["network"] = network
    end
    if controlPlane != {}:
        vars["controlPlane"] = controlPlane
    end

    worker = {}
    if data.values["WORKER_MACHINE_COUNT"] != "":
        worker["count"] = data.values["WORKER_MACHINE_COUNT"]
    end
    machine = {}
    if data.values["VSPHERE_WORKER_NUM_CPUS"] != "":
        machine["numCPUs"] = data.values["VSPHERE_WORKER_NUM_CPUS"]
    end
    if data.values["VSPHERE_WORKER_DISK_GIB"] != "":
        machine["diskGiB"] = data.values["VSPHERE_WORKER_DISK_GIB"]
    end
    if data.values["VSPHERE_WORKER_MEM_MIB"] != "":
        machine["memoryMiB"] = data.values["VSPHERE_WORKER_MEM_MIB"]
    end
    if machine != {}:
        worker["machine"] = machine
    end

    network = {}
    if data.values["WORKER_NODE_NAMESERVERS"] != None:
        network["nameservers"] = data.values["WORKER_NODE_NAMESERVERS"]
    end
    if network != {}:
        worker["network"] = network
    end
    if worker != {}:
        vars["worker"] = worker
    end

    return vars
end