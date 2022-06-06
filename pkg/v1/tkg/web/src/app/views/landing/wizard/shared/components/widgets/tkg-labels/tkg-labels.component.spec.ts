import { TkgLabelsComponent } from './tkg-labels.component';
import { FormArray, FormBuilder, FormControl, FormGroup } from "@angular/forms";
import { ControlType } from "../../../field-mapping/FieldMapping";
import { SimpleValidator } from "../../../constants/validation.constants";
import { TKGLabelsConfig } from "./interfaces/tkg-labels.interface";
import { FormUtils } from "../../../utils/form-utils";

describe('TkgLabelsComponent', () => {
    let component: TkgLabelsComponent;
    let parentFormGroup: FormGroup;

    const formBuilder = new FormBuilder();
    const tkgLabelsConfig: TKGLabelsConfig = {
        label: {
            title: 'LABELS (OPTIONAL)',
            tooltipText: `Optionally specify labels for the Management cluster.`
        },
        forms: {
            parent: null,
            control: null
        },
        fields: {
            clusterType: 'Management',
            fieldMapping: {
                name: 'clusterLabels',
                label: 'LABELS (OPTIONAL)',
                controlType: ControlType.FormArray,
                children: [
                    {
                        name: 'key',
                        defaultValue: '',
                        controlType: ControlType.FormControl,
                        validators: [
                            SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION,
                            SimpleValidator.RX_UNIQUE,
                            SimpleValidator.RX_REQUIRED_IF_VALUE
                        ]
                    },
                    {
                        name: 'value',
                        defaultValue: '',
                        controlType: ControlType.FormControl,
                        validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION, SimpleValidator.RX_REQUIRED_IF_KEY]
                    }
                ]
            }
        }
    };

    beforeEach(() => {
        parentFormGroup = formBuilder.group({
            clusterLabels: formBuilder.array([
                formBuilder.group({
                    key: [''],
                    value: ['']
                })
            ])
        });
        tkgLabelsConfig.forms.parent = parentFormGroup;
        tkgLabelsConfig.forms.control = parentFormGroup.get('clusterLabels') as FormArray;

        component = new TkgLabelsComponent();
        component.config = tkgLabelsConfig;
    });

    it('should create label', () => {
        const group = component.createLabel();
        expect(group.controls.key).toBeInstanceOf(FormControl);
        expect(group.controls.value).toBeInstanceOf(FormControl);
    })

    it('should not add new Label on form invalid', () => {
        spyOn(component.labelsFormArray, 'markAllAsTouched');
        spyOn(component.labelsFormArray, 'push');

        spyOnProperty(component.labelsFormArray, 'invalid').and.returnValue(true);

        component.addNewLabel();

        expect(component.labelsFormArray.markAllAsTouched).toHaveBeenCalledTimes(1);
        expect(component.labelsFormArray.push).not.toHaveBeenCalledTimes(1);

    });

    it('should add new Label on form valid', () => {
        spyOn(component.labelsFormArray, 'markAllAsTouched');
        spyOn(FormUtils, 'addDynamicControl');

        spyOnProperty(component.labelsFormArray, 'invalid').and.returnValue(false);

        expect(component.labelsFormArray.controls.length).toEqual(1);

        component.addNewLabel();

        expect(component.labelsFormArray.markAllAsTouched).toHaveBeenCalledTimes(1);
        expect(FormUtils.addDynamicControl).toHaveBeenCalledTimes(2);
        expect(component.labelsFormArray.length).toEqual(2);

    });

    it('should delete label', () => {
        component.deleteLabel(0);
        expect(component.labelsFormArray.length).toEqual(0)
    })

});
