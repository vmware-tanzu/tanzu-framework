import { WizardBaseDirective } from '../../wizard-base/wizard-base';
import { IdentityManagementType, WizardForm } from '../../constants/wizard.constants';
import { Type } from '@angular/core';
import { StepFormDirective } from '../../step-form/step-form';

interface I18nDataForHtml {
    title: string,
    description: string,
}
export interface FormDataForHTML {
    name: string,           // name of this step (not displayable)
    description: string,    // description of this step (displayed)
    title: string,          // title of this step (displayed)
    i18n: I18nDataForHtml,  // data used for navigating UI (displayed)
    clazz: Type<StepFormDirective>,    // the class of the step component
}

export class FormUtility {
    static titleCase(target): string {
        if (target === undefined || target === null || target.length === 0) {
            return '';
        }
        return target.replace(/(^|\s)\S/g, function(t) { return t.toUpperCase() });
    }

    static formWithOverrides(formData: FormDataForHTML, overrideData: { description?: string, clazz?: Type<StepFormDirective> }):
        FormDataForHTML {
        if (overrideData.description) {
            formData.description = overrideData.description;
        }
        if (overrideData.clazz) {
            formData.clazz = overrideData.clazz;
        }
        return formData;
    }

    static formOverrideClazz(formData: FormDataForHTML, clazz: Type<StepFormDirective>) {
        formData.clazz = clazz;
        return formData;
    }
}
