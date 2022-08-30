import { FormArray, FormGroup } from '@angular/forms';

export interface TKGLabelsConfig {
    label: {
        title: string;
        tooltipText: string;
        helperText?: string;
    };
    forms: {
        parent: FormGroup; // the parent form group
        control: FormArray; // the control of the labels form array
    };
    fields: {
        [key: string]: any;
    };
}
