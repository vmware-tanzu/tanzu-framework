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
"ANTREA_EGRESS": ["vsphere", "aws", "azure", "docker"],
"ANTREA_FLOWEXPORTER": ["vsphere", "aws", "azure", "docker"],
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
    for configVariable in kvs:
        if data.values.PROVIDER_TYPE in kvs[configVariable]:
            if data.values[configVariable] != None:
                vars[configVariable] = data.values[configVariable]
            else:
                continue
            end
            if configVariable == "TKG_HTTP_PROXY":
                if vars["TKG_HTTP_PROXY"] != "":
                    vars["proxy"] = {
                        "httpProxy": vars["TKG_HTTP_PROXY"],
                        "httpsProxy": data.values["TKG_HTTPS_PROXY"],
                        "noProxy": data.values["TKG_NO_PROXY"].split(",")
                    }
                else:
                    vars["proxy"] = None
                end
            end
            if configVariable == "TKG_CUSTOM_IMAGE_REPOSITORY":
                vars["TKG_CUSTOM_IMAGE_REPOSITORY_HOSTNAME"] = vars[configVariable].split("/")[0]
            end
        end
    end
    return vars
end
