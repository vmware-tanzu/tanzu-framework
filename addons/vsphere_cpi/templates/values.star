load("@ytt:data", "data")
load("@ytt:assert", "assert")

def validate_vsphereCPI():
   data.values.vsphereCPI.server or assert.fail("vsphereCPI server should be provided")
   data.values.vsphereCPI.datacenter or assert.fail("vsphereCPI datacenter should be provided")
   data.values.vsphereCPI.username or assert.fail("vsphereCPI username should be provided")
   data.values.vsphereCPI.password or assert.fail("vsphereCPI password should be provided")
   if not data.values.vsphereCPI.insecureFlag:
     data.values.vsphereCPI.tlsThumbprint or assert.fail("vsphereCPI tlsThumbprint should be provided when insecureFlag is False")
   end
   if data.values.vsphereCPI.ipFamily and (data.values.vsphereCPI.ipFamily not in ["ipv4", "ipv6"]):
     assert.fail("vsphereCPI ipFamily should be either ipv4 or ipv6 if provided")
   end
end

def validate_nsxt_config():
   if data.values.vsphereCPI.nsxt.vmcAccessToken != "":
     if data.values.vsphereCPI.nsxt.vmcAuthHost == "":
       assert.fail("vmc auth host must be provided if access token is provided")
     end
   elif data.values.vsphereCPI.nsxt.username != "":
     if data.values.vsphereCPI.nsxt.password == "":
       assert.fail("password is reqruied if username is provided")
     end
     if data.values.vsphereCPI.nsxt.secretName == "" or data.values.vsphereCPI.nsxt.secretNamespace == "":
       assert.fail("secretName and secretNamespace should not be empty if username and password are provided")
     end
   elif data.values.vsphereCPI.nsxt.clientCertKeyData != "":
     if data.values.vsphereCPI.nsxt.clientCertData == "":
       assert.fail("client cert data is required if client cert key data is provided")
     end
   elif data.values.vsphereCPI.nsxt.clientCertData != "":
     if data.values.vsphereCPI.nsxt.clientCertKeyData == "":
       assert.fail("client cert key data is required if client cert data is provided")
     end
   elif data.values.vsphereCPI.nsxt.secretName != "":
     if data.values.vsphereCPI.nsxt.secretNamespace == "":
       assert.fail("secret namespace is required if secret name is provided")  
     end
   elif data.values.vsphereCPI.nsxt.secretNamespace != "":
     if data.values.vsphereCPI.nsxt.secretName == "":
       assert.fail("secret name is required if secret namespace is provided")  
     end
   else:
     assert.fail("user/password or vmc access token or client certificates must be set")  
   end
   data.values.vsphereCPI.nsxt.host or assert.fail("vsphereCPI nsxtHost should be provided")
end

# export
values = data.values

# validate
validate_vsphereCPI()
if data.values.vsphereCPI.nsxt.podRoutingEnabled:
validate_nsxt_config()
end
