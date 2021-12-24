import { Validators, FormControl } from '@angular/forms';
import { takeUntil } from 'rxjs/operators';

import { StepFormDirective } from '../../../step-form/step-form';
import { TkgEventType } from 'src/app/shared/service/Messenger';
import Broker from 'src/app/shared/service/broker';
import { Observable } from 'rxjs/internal/Observable';
import { FormUtils } from '../../../utils/form-utils';
import { AWSVirtualMachine, AzureVirtualMachine } from 'src/app/swagger/models';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { OsImageStepMapping } from './os-image-step.fieldmapping';
import { StepMapping } from '../../../field-mapping/FieldMapping';

// We define an OsImage as one that has a name field, so that our HTML can dereference this field in the listbox display
export interface OsImage {
    name?: string
}
export interface OsImageProviderInputs<IMAGE> {
    event: TkgEventType,
    osImageService: OsImageService<IMAGE>,
    osImageTooltipContent: string,
    nonTemplateAlertMessage?: string,
    noImageAlertMessage?: string,
}
export interface OsImageService<IMAGE> {
    getDataStream(TkgEventType): Observable<IMAGE[]>,
    getErrorStream(TkgEventType): Observable<string>,
}
export abstract class SharedOsImageStepComponent<IMAGE extends OsImage> extends StepFormDirective {
    // used by HTML as well as locally
    public providerInputs: OsImageProviderInputs<IMAGE>;

    osImages: Array<IMAGE>;
    loadingOsTemplate: boolean = false;
    displayNonTemplateAlert: boolean = false;
    // TODO: It's questionable whether tkrVersion should be in this class, since it's only used for vSphere
    tkrVersion: Observable<string>;

    protected constructor(protected fieldMapUtilities: FieldMapUtilities) {
        super();
        this.tkrVersion = Broker.appDataService.getTkrVersion();
    }

    // This method allows child classes to supply the inputs (rather than having them passed as part of an HTML component tag).
    // This allows the step to follow the same pattern as all the other steps, which only take formGroup and formName as inputs.
    protected abstract supplyProviderInputs(): OsImageProviderInputs<IMAGE>;

    private customizeForm() {
        this.providerInputs.osImageService.getErrorStream(this.providerInputs.event)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(error => {
                this.errorNotification = error;
            });

        this.providerInputs.osImageService.getDataStream(this.providerInputs.event)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((images: Array<IMAGE>) => {
                this.osImages = images;
                this.loadingOsTemplate = false;
                if (this.osImages.length === 1) {
                    this.setControlValueSafely('osImage', images[0]);
                } else {
                    this.setControlWithSavedValue('osImage', '');
                }
            });
    }

    // onInit() should be called from subclass' ngOnInit()
    protected onInit() {
        super.ngOnInit();
        this.fieldMapUtilities.buildForm(this.formGroup, this.formName, OsImageStepMapping);
        super.ngOnInit();
        this.providerInputs = this.supplyProviderInputs();
        this.customizeForm();
        this.initFormWithSavedData();
    }

    /**
     * @method retrieveOsImages
     * helper method to retrieve and preload list of available OS images from connected VC environment.
     * emits list of OS images to wizard-data.service
     */
    retrieveOsImages() {
        this.loadingOsTemplate = true;
        this.displayNonTemplateAlert = false;
        Broker.messenger.publish({
            type: this.providerInputs.event
        });
    }

    /**
     * @method onOptionsSelected
     * helper method to determine if osImage.isTemplate is true or false; if false show warning
     */
    onOptionsSelected() {
        this.displayNonTemplateAlert = !this.formGroup.value['osImage'].isTemplate;
    }

    dynamicDescription(): string {
        if (this.getFieldValue('osImage', true) && this.getFieldValue('osImage').name) {
            return 'OS Image: ' + this.getFieldValue('osImage').name;
        }
        return 'Specify the OS Image';
    }
}
