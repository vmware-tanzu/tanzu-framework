import { browser, by, element } from 'protractor';
import { Step } from '../step.po';

export class Provider extends Step {
    hasMovedToStep() {
        throw new Error("Method not implemented.");
    }

    getTitleText() {
        return element(by.css('clr-stepper-panel[formgroupname="azureProviderForm"] clr-step-description')).getText() as Promise<string>;
    }

    getTenantId() {
        return element(by.css('input[formcontrolname="tenantId"]'));
    }

    getClientId() {
        return element(by.css('input[formcontrolname="clientId"]'));
    }

    getClientSecret() {
        return element(by.css('input[formcontrolname="clientSecret"]'));
    }

    getSubscriptionId() {
        return element(by.css('input[formcontrolname="subscriptionId"]'));
    }

    getSshPublicKey() {
        return element(by.css('textarea[formControlName="sshPublicKey"]'));
    }

    getRegion() {
        return element(by.css('select[name="region"]'));
    }

    getAzureCloud() {
        return element(by.css('select[name="azureCloud"]'));
    }

    getCreateNewResourceGroup() {
        return element(by.cssContainingText("label", "Create a new resource group"));
    }

    getCustomResourceGroup() {
        return element(by.css('input[formControlName="resourceGroupCustom"]'));
    }

    getExistingResourceGroup() {
        return element(by.cssContainingText("label", "Select an existing resource group"));
    }

    selectExistingResourceGroup() {
        return element(by.css('select[name=\"resourceGroupExisting\"]'));
    }

    getConectButton() {
        return element(by.id('btn-azure-provider-connect'));
    }
}
