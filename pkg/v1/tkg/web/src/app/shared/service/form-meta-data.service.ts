import { Injectable } from '@angular/core';
import { FormMetaDataStore } from 'src/app/views/landing/wizard/shared/FormMetaDataStore';

@Injectable({
    providedIn: 'root'
})
export class FormMetaDataService {

    constructor() {}

    /**
     * Save the form meta data so as to produce the confirmation page with
     */
    saveFormMetadata(formName: string, domForm: any) {
        if (domForm) {
            const fields = Array.from(domForm.querySelectorAll('[data-step-metadata]'));
            fields.forEach((domField: any) => {
                this.saveDataFromDomField(formName, domField);
            })
        }
    }

    /**
     * Extract and save the meta data for one particular form field by analyzing the
     * native DOM tree.
     */
    private saveDataFromDomField(formName: string, domField: any) {
        const label = this.extractLabelFromDomField(domField);
        const controlName = this.extractControlNameFromDomField(domField);
        const saveRequiresValue = this.extractSaveRequiresValueFromDomField(domField);
        const displayValue = this.extractDisplayValueFromDomField(domField);
        const key = this.extractKeyFromDomField(domField);

        if (this.shouldSaveValue(displayValue, saveRequiresValue)) {
            FormMetaDataStore.saveMetaDataEntry(formName, controlName, {
                label,
                displayValue,
                key
            });
        }
    }

    private shouldSaveValue(displayValue: string, saveRequiresValue: boolean) {
        // either the domField should have a value, or it's ok to save without a value
        return displayValue !== '' || !saveRequiresValue;
    }

    // extractKeyFromDomField() returns the selected listbox key (if the control is a listbox); blank string otherwise.
    private extractKeyFromDomField(domField: any): string {
        const controlEl = domField.querySelector("[formcontrolname]");
        if (controlEl && controlEl.tagName === "SELECT" && controlEl.value !== undefined && controlEl.value !== 'undefined') {
            return controlEl.value;
        }
        return '';
    }

    private extractDisplayValueFromDomField(domField: any) {
        let displayValue = domField.getAttribute("data-value");
        // First handle the case of value in container node
        if (!displayValue) {
            const controlEl = domField.querySelector("[formcontrolname]");
            if (controlEl) {
                /**
                 * Get control display text: this may be different from
                 * control value in cases of "select", "checkbox" and "rodio"
                 * etc. for confirmation purpose, hence requires special handling.
                 */
                if (controlEl.tagName === "SELECT") {
                    if (controlEl.options) {
                        const options: any[] = Array.from(controlEl.options);
                        displayValue = options.filter(option => option.selected).map(option => option.label).join(",");
                    }
                } else {
                    switch (controlEl.type) {
                        case "password": {
                            displayValue = "*".repeat(controlEl.value ? controlEl.value.length : 0);
                            break;
                        }
                        case "checkbox": {
                            displayValue = controlEl.checked ? "yes" : "no";
                            break;
                        }
                        case "radio": {
                            throw new Error("input type of 'radio' is yet to be supported");
                        }
                        case "file": {
                            throw new Error("input type of 'file' is yet to be supported");
                        }
                        case "hidden":
                        case "button":
                        case "reset":
                        case "submit":
                        case "image": {
                            // ignore
                            break;
                        }
                        case "": {
                            break;
                        }
                        default: {  // Assuming "text", "number", "date" etc.
                            displayValue = controlEl.value;
                        }
                    }
                }
            }
        }
        return displayValue;
    }

    private extractLabelFromDomField(domField: any) {
        // First handle the case where label or value are hard-coded in the container node
        let label = domField.getAttribute("data-label");
        if (!label) {
            let labelEl = domField.querySelector("label");
            if (labelEl) {
                labelEl = labelEl.cloneNode(true);
                // strip tooltip from label content if found
                const labelRemoveContent = labelEl.querySelector("clr-tooltip");
                if (labelRemoveContent) {
                    labelEl.removeChild(labelRemoveContent);
                }
                label = labelEl.getAttribute("data-full") || labelEl.innerHTML || "";
            }
        }
        return label && label.toLocaleUpperCase();
    }

    private extractControlNameFromDomField(domField: any) {
        let controlName = domField.getAttribute("data-name");
        // First handle the case of value in container node
        if (!domField.getAttribute("data-value")) {
            const controlEl = domField.querySelector("[formcontrolname]");
            if (controlEl) {
                controlName = controlEl.getAttribute("formcontrolname");
            }
        }
        return controlName;
    }

    private extractSaveRequiresValueFromDomField(domField: any) {
        let saveRequiresValue = domField.getAttribute("save-requires-value") === 'true';
        // First handle the case of value in container node
        if (!domField.getAttribute("data-value")) {
            const controlEl = domField.querySelector("[formcontrolname]");
            if (controlEl) {
                saveRequiresValue = controlEl.getAttribute("save-requires-value") === 'true';
            }
        }
        return saveRequiresValue;
    }

    /**
     * saveFormFieldData saves the data value for one particular form field
     * Note that at this time it destroys the LABEL value entirely, trusting that
     * when the confirmation page is displayed, the label will be made available
     */
    saveFormFieldData(formName, controlName, value) {
        FormMetaDataStore.saveMetaDataEntry(formName, controlName, {
            label: '',
            displayValue: value
        })
    }

    saveFormListboxData(formName, listboxName, key) {
        console.log('***** saveFormListboxData() saving ' + formName + '.' + listboxName + ' KEY: ' + key);
        FormMetaDataStore.saveMetaDataEntry(formName, listboxName, {
            label: '',
            key
        })
    }
}
