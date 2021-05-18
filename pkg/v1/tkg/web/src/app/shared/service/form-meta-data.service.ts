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
    saveFormMetadata(formName: string, container: any) {
        if (container) {
            const fields = Array.from(container.querySelectorAll('[data-step-metadata]'));
            fields.forEach((f: any) => {
                this.getFormMetadata(formName, f);
            })
        }
    }

    /**
     * Extract the meta data for one particular form field by analyzing the
     * native DOM tree.
     *
     * TODO: supports addtional form control types for a complete framework.
     */
    getFormMetadata(formName: string, container: any) {

        // First handle the case where label or value are hard-coded in the container node
        let label = container.getAttribute("data-label");
        let displayValue = container.getAttribute("data-value");
        let controlName = container.getAttribute("data-name");

        if (!label) {
            let labelEl = container.querySelector("label");
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

        label = label && label.toLocaleUpperCase();

        // First handle the case of value in container node
        if (!displayValue) {
            const controlEl = container.querySelector("[formcontrolname]");
            if (controlEl) {
                controlName = controlEl.getAttribute("formcontrolname");

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

        FormMetaDataStore.saveMetaDataEntry(formName, controlName, {
            label,
            displayValue
        });
    }
}
