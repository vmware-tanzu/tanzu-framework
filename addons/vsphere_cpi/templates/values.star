load("@ytt:data", "data")
load("@ytt:assert", "assert")

def validate_vsphereCPI():
   data.values.vsphereCPI.server or assert.fail("vsphereCPI server should be provided")
   data.values.vsphereCPI.datacenter or assert.fail("vsphereCPI datacenter should be provided")
   data.values.vsphereCPI.publicNetwork or assert.fail("vsphereCPI publicNetwork should be provided")
   data.values.vsphereCPI.username or assert.fail("vsphereCPI username should be provided")
   data.values.vsphereCPI.password or assert.fail("vsphereCPI password should be provided")
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
   elif data.values.vsphereCPI.nsxt.clientAuthKeyFile != "":
     if data.values.vsphereCPI.nsxt.clientAuthCertFile == "":
       assert.fail("client cert file is required if client key file is provided")
     end
   elif data.values.vsphereCPI.nsxt.clientAuthCertFile != "":
     if data.values.vsphereCPI.nsxt.clientAuthKeyFile == "":
       assert.fail("client key file is required if client cert file is provided")
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
     assert.fail("user/password or vmc access token or client cert file must be set")  
   end
   data.values.vsphereCPI.nsxt.host or assert.fail("vsphereCPI nsxtHost should be provided")
end

# export
values = data.values

# validate
validate_vsphereCPI()
validate_nsxt_config()
