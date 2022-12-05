load("@ytt:data", "data")
load("@ytt:assert", "assert")
load("/lib/helpers.star", "validate_proxy_bypass_vsphere_host")

required_variable_list_vsphere = [
  "VSPHERE_USERNAME",
  "VSPHERE_PASSWORD",
  "VSPHERE_SERVER",
  "VSPHERE_DATACENTER",
  "VSPHERE_RESOURCE_POOL",
  "VSPHERE_FOLDER",
  "VSPHERE_SSH_AUTHORIZED_KEY"]

required_variable_list_aws = [
  "AWS_REGION",
  "AWS_SSH_KEY_NAME"]

required_variable_list_azure = [
  "AZURE_TENANT_ID",
  "AZURE_SUBSCRIPTION_ID",
  "AZURE_CLIENT_ID",
  "AZURE_CLIENT_SECRET",
  "AZURE_LOCATION",
  "AZURE_SSH_PUBLIC_KEY_B64"]

required_variable_list_tkgs = [
  "CONTROL_PLANE_STORAGE_CLASS",
  "CONTROL_PLANE_VM_CLASS",
  "SERVICE_DOMAIN",
  "WORKER_STORAGE_CLASS",
  "WORKER_VM_CLASS",
  "NODE_POOL_0_NAME"]

def validate_configuration(provider):
  #! skip validation when only employing the template to generate
  #! addon resources (which is doable without a complete set of 
  #! config data values).
  if data.values.FILTER_BY_ADDON_TYPE:
    return
  end

  if provider == "vsphere":
    flag_missing_variable_error(required_variable_list_vsphere)
    if data.values.NSXT_POD_ROUTING_ENABLED == True:
      validate_nsxt_config()
    end
    #! known issue for govc: https://github.com/vmware/govmomi/issues/2494
    #! TODO: remove the validation once the issue is resolved
    if data.values.TKG_HTTP_PROXY != "":
      validate_proxy_bypass_vsphere_host()
    end
  elif provider == "aws":
    flag_missing_variable_error(required_variable_list_aws)
  elif provider == "azure":
    flag_missing_variable_error(required_variable_list_azure)
  elif provider == "tkgs":
    flag_missing_variable_error(required_variable_list_tkgs)
  end
end

def flag_missing_variable_error(variables_list):
  missing_variable_str = ""
  for variable in variables_list:
    value = getattr(data.values, variable, None)
    if value == None or value == "":
      missing_variable_str = missing_variable_str + variable + ", "
    end
  end
  if missing_variable_str != "":
    assert.fail("missing configuration variables: " + missing_variable_str[:-2])
  end
end

def validate_nsxt_config():
   if data.values.NSXT_VMC_ACCESS_TOKEN != "":
     if data.values.NSXT_VMC_AUTH_HOST == "":
       assert.fail("vmc auth host must be provided if access token is provided")
     end
   elif data.values.NSXT_USERNAME != "":
     if data.values.NSXT_PASSWORD == "" or data.values.NSXT_PASSWORD == "None":
       assert.fail("password is reqruied if username is provided")
     end
     if data.values.NSXT_SECRET_NAME == "" or data.values.NSXT_SECRET_NAMESPACE == "":
       assert.fail("secretName and secretNamespace should not be empty if username and password are provided")
     end
   elif data.values.NSXT_CLIENT_CERT_KEY_DATA != "":
     if data.values.NSXT_CLIENT_CERT_DATA == "":
       assert.fail("client cert data is required if client cert key data is provided")
     end
   elif data.values.NSXT_CLIENT_CERT_DATA != "":
     if data.values.NSXT_CLIENT_CERT_KEY_DATA == "":
       assert.fail("client cert key data is required if client cert data is provided")
     end
   elif data.values.NSXT_SECRET_NAME != "":
     if data.values.NSXT_SECRET_NAMESPACE == "":
       assert.fail("secret namespace is required if secret name is provided")  
     end
   elif data.values.NSXT_SECRET_NAMESPACE != "":
     if data.values.NSXT_SECRET_NAME == "":
       assert.fail("secret name is required if secret namespace is provided")  
     end
   else:
     assert.fail("user/password or vmc access token or client certificates must be set")  
   end
   data.values.NSXT_MANAGER_HOST != "" or assert.fail("missing configuration variables: NSXT_MANAGER_HOST")
end
