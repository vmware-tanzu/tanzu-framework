import { browser, by, element } from 'protractor';
import { Step } from '../step.po';

export class OsImage extends Step {
    hasMovedToStep() {
        return this.getOsImages().isPresent();
    }

    getTitleText() {
        return element(by.css('clr-stepper-panel[formgroupname="osImageForm"] clr-step-description')).getText() as Promise<string>;
    }

    getOsImages() {
        return element(by.css('select[formcontrolname="osImage"]'));
    }

    getRefreshButton() {
        return element(by.css('app-os-image-step clr-icon[shape="refresh"]'));
    }

    getOsImageCount() {
        return element.all(by.css('select[name="osImage"] option')).count();
    }
}
