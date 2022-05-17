load("@ytt:data", "data")
load("@ytt:assert", "assert")

def getWebhookServerPort():
    webhookServerPort = str(data.values.deployment.webhookServerPort)
    if hasattr(data.values, 'deployment') and webhookServerPort.isdigit():
        return data.values.deployment.webhookServerPort
    else:
        assert.fail("Invalid webhook server port!!")
    end
end
