load("@ytt:data", "data")
load("@ytt:assert", "assert")

def validate_vsphereCPI():
   data.values.vsphereCPI.server or assert.fail("vsphereCPI server should be provided")
   data.values.vsphereCPI.datacenter or assert.fail("vsphereCPI datacenter should be provided")
   data.values.vsphereCPI.publicNetwork or assert.fail("vsphereCPI publicNetwork should be provided")
   data.values.vsphereCPI.username or assert.fail("vsphereCPI username should be provided")
   data.values.vsphereCPI.password or assert.fail("vsphereCPI password should be provided")
end

#export
values = data.values

# validate
validate_vsphereCPI()
