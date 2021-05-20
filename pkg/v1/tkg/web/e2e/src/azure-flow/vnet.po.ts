import { browser, by, element } from 'protractor';
import { Step } from '../step.po';

export class Vnet extends Step {
    hasMovedToStep() {
        return this.getSelectAnExistingVnet().isPresent();
    }

    getSelectAnExistingVnet() {
        return element(by.cssContainingText("label", "Select an existing VNET"));
    }

    getCreateNewVnet() {
        return element(by.cssContainingText("label", "Create a new VNET on Azure"));
    }

    getResourceGroup() {
        return element(by.css('select[formControlName="resourceGroup"]'))
    }

    getVnetName() {
        return element(by.css('input[formControlName="vnetNameCustom"]'))
    }

    getVnetNameExisting() {
        return element(by.css('select[formControlName="vnetNameExisting"]'))
    }

    getVnetCidrText() {
        return element(by.css('input[formControlName="vnetCidrBlock"]')).getAttribute('value') as Promise<string>;
    }

    getControlPlaneSubnet() {
        return element(by.css('select[formControlName="controlPlaneSubnet"]'))
    }

    getWorkerNodeSubnet() {
        return element(by.css('select[formControlName="workerNodeSubnet"]'))
    }

    getControlPlaneSubnetNew() {
        return element(by.css('input[formControlName="controlPlaneSubnetNew"]'))
    }

    getControlPlaneSubnetCidrNew() {
        return element(by.css('input[formControlName="controlPlaneSubnetCidrNew"]'))
    }

    getControlPlaneSubnetCidrNewText() {
        return element(by.css('input[formControlName="controlPlaneSubnetCidrNew"]')).getAttribute('value') as Promise<string>;
    }

    getWorkerNodeSubnetNew() {
        return element(by.css('input[formControlName="workerNodeSubnetNew"]'))
    }

    getWorkerNodeSubnetCidrNew() {
        return element(by.css('input[formControlName="workerNodeSubnetCidrNew"]'))
    }

    getWorkerNodeSubnetCidrNewText() {
        return element(by.css('input[formControlName="workerNodeSubnetCidrNew"]')).getAttribute('value') as Promise<string>;
    }

    getPrivateCluster() {
        return element(by.cssContainingText("label", "PRIVATE AZURE CLUSTER"));
    }

    getPrivateIP() {
        return element(by.css('input[formControlName="privateIP"]'))
    }
}
