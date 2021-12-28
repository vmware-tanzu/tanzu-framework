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
import ServiceBroker from '../../../../../../../shared/service/service-broker';

// We define an OsImage as one that has a name field, so that our HTML can dereference this field in the listbox display
export interface OsImage {
    name?: string
}
export interface OsImageProviderInputs<IMAGE> {
    event: TkgEventType,
    fetcher: (data: any) => Observable<IMAGE[]>,
    osImageTooltipContent: string,
    nonTemplateAlertMessage?: string,
    noImageAlertMessage?: string,
}
export abstract class SharedOsImageStepComponent<IMAGE extends OsImage> extends StepFormDirective {
    // used by HTML as well as locally
    public providerInputs: OsImageProviderInputs<IMAGE>;

    osImages: Array<IMAGE>;
    loadingOsTemplate: boolean = false;
    displayNonTemplateAlert: boolean = false;
    // TODO: It's questionable whether tkrVersion should be in this class, since it's only used for vSphere
    tkrVersion: Observable<string>;

    protected constructor(protected fieldMapUtilities: FieldMapUtilities, protected serviceBroker: ServiceBroker) {
        super();
        this.tkrVersion = Broker.appDataService.getTkrVersion();
    }

    // This method allows child classes to supply the inputs (rather than having them passed as part of an HTML component tag).
    // This allows this step to follow the same pattern as all the other steps, which only take formGroup and formName as inputs.
    protected abstract supplyProviderInputs(): OsImageProviderInputs<IMAGE>;

    private registerProviderService() {
        // our provider event should be associated with the provider fetcher (which will get the data from the backend)
        this.serviceBroker.register<IMAGE>(this.providerInputs.event, this.providerInputs.fetcher);
    }

    private subscribeToProviderEvent() {
        // we register a handler for when our event receives data, namely that we'll populate our array of osImages
        this.serviceBroker.stepSubscribe<IMAGE>(this, this.providerInputs.event, this.onFetchedOsImages.bind(this));
    }

    private onFetchedOsImages(images: Array<IMAGE>) {
        this.osImages = images;
        this.loadingOsTemplate = false;
        if (this.osImages.length === 1) {
            this.setControlValueSafely('osImage', images[0]);
        } else {
            this.setControlWithSavedValue('osImage', '');
        }
    }

    // onInit() should be called from subclass' ngOnInit()
    protected onInit() {
        super.ngOnInit();
        this.fieldMapUtilities.buildForm(this.formGroup, this.formName, OsImageStepMapping);
        this.providerInputs = this.supplyProviderInputs();
        this.registerProviderService();
        this.subscribeToProviderEvent();
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
