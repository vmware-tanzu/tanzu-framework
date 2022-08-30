import { by, element } from 'protractor';
import { Step } from '../step.po';

export class Metadata extends Step {
    hasMovedToStep() {
        return this.getMCDescription().isPresent();
    }

    getMCDescription() {
        return element(by.css('textarea[formcontrolname="clusterDescription"]'));
    }

    getMCLocation() {
        return element(by.css('input[formcontrolname="clusterLocation"]'));
    }

    getMCLabelsKey() {
        return element(by.css('input[formcontrolname="newLabelKey"]'));
    }

    getMCLabelsValue() {
        return element(by.css('input[formcontrolname="newLabelValue"]'));
    }

    getMCLabelsAddButton() {
        return element(by.buttonText('ADD'));
    }

    getMCLabelsDeleteButton(key: string) {
        const id = "label-delete-" + key;
        return element(by.id(id));
    }

}
