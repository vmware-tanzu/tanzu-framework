// Angular imports
import { Component, OnInit } from '@angular/core';
import { FormArray } from '@angular/forms';
// App imports
import AppServices from '../../../../../../../shared/service/appServices';
import { StepMapping } from '../../../field-mapping/FieldMapping';
import { StepFormDirective } from '../../../step-form/step-form';
import { TKGLabelsConfig } from '../../widgets/tkg-labels/interfaces/tkg-labels.interface';
import { MetadataField, MetadataStepMapping } from './metadata-step.fieldmapping';

@Component({
    selector: 'app-metadata-step',
    templateUrl: './metadata-step.component.html',
    styleUrls: ['./metadata-step.component.scss']
})
export class MetadataStepComponent extends StepFormDirective implements OnInit {
    tkgLabelsConfig: TKGLabelsConfig;

    constructor() {
        super();
    }

    ngOnInit() {
        super.ngOnInit();
        AppServices.userDataFormService.buildForm(this.formGroup, this.wizardName, this.formName, MetadataStepMapping);
        this.htmlFieldLabels = AppServices.fieldMapUtilities.getFieldLabelMap(MetadataStepMapping);
        this.storeDefaultLabels(MetadataStepMapping);
        this.registerStepDescriptionTriggers({
            fields: [MetadataField.CLUSTER_LOCATION],
            clusterTypeDescriptor: true
        });
        this.registerDefaultFileImportedHandler(this.eventFileImported, MetadataStepMapping);
        this.registerDefaultFileImportErrorHandler(this.eventFileImportError);

        this.tkgLabelsConfig = {
            label: {
                title: this.htmlFieldLabels['clusterLabels'],
                tooltipText: `Optionally specify labels for the ${this.clusterTypeDescriptor} cluster.`
            },
            forms: {
                parent: this.formGroup,
                control: this.formGroup.get('clusterLabels') as FormArray
            },
            fields: {
                clusterTypeDescriptor: this.clusterTypeDescriptorTitleCase,
                fieldMapping: MetadataStepMapping.fieldMappings.find((m) => m.name === MetadataField.CLUSTER_LABELS)
            }
        };
    }

    dynamicDescription(): string {
        const clusterLocation = this.getFieldValue(MetadataField.CLUSTER_LOCATION, true);
        return clusterLocation
            ? `Location: ${clusterLocation}`
            : `Specify metadata for the ${this.clusterTypeDescriptor} cluster`;
    }

    protected storeUserData() {
        this.storeUserDataFromMapping(MetadataStepMapping);
        this.storeDefaultDisplayOrder(MetadataStepMapping);
    }

}
