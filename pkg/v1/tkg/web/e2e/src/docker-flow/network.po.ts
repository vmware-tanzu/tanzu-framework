import { browser, by, element } from 'protractor';
import { Step } from '../step.po';

export class Network extends Step {
    hasMovedToStep() {
        return this.getClusterPodCidr().isPresent();
    }

    getClusterPodCidr() {
        return element(by.css('input[formcontrolname="clusterPodCidr"]'));
    }

    getTitleText() {
        return element(by.css('clr-stepper-panel[formgroupname="networkForm"] clr-step-description')).getText() as Promise<string>;
    }

}
