import { browser, by, element } from 'protractor';
import { Step } from '../step.po';

export class Resources extends Step {
    hasMovedToStep() {
        return this.getDataStore().isPresent();
    }

    getTitleText() {
        return element(by.css('clr-stepper-panel[formgroupname="resourceForm"] clr-step-description')).getText() as Promise<string>;
    }

    getResources() {
        return element.all(by.css('clr-tree label.clr-control-label'));
    }

    getResource(poolName) {
        return element(by.cssContainingText('label.clr-control-label', poolName));
    }

    getVMfolder() {
        return element(by.css('ng-select[name="vmFolder"]'));
    }

    getDataStore() {
        return element(by.css('datalist[id="datastore"]'));
    }

    getRefreshButton() {
        return element.all(by.css('app-resource-step clr-icon[shape="refresh"]'));
    }

    getResourcePoolSize() {
        return element.all(by.css('app-tree-select clr-checkbox-wrapper')).count();
    }
}
