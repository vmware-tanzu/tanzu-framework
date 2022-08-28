// Angular imports
import { Directive, OnInit } from '@angular/core';
// Third party imports
import { Observable } from 'rxjs/internal/Observable';
// App imports
import AppServices from '../../../../../../../shared/service/appServices';
import { OsImageField, OsImageStepMapping } from './os-image-step.fieldmapping';
import { StepFormDirective } from '../../../step-form/step-form';
import { StepMapping } from '../../../field-mapping/FieldMapping';
import { TanzuEventType } from 'src/app/shared/service/Messenger';

// The intention of this class is to provide the common plumbing for the osImage step that many providers need.
// The basic functionality is to subscribe to an event and load the resulting images into a local field.
// Note that we assume that the event has been registered already with a backend service that will return an array of images
// of type IMAGE

// We define an OsImage as one that has a name field, so that our HTML can dereference this field in the listbox display.
// Even though ALL the osImage types do have a name field, it is generated as OPTIONAL by Swagger, so we leave it optional in our interface.
export interface OsImage {
    name?: string
}
export interface OsImageProviderInputs {
    createOsImageEventPayload?: () => any,  // some providers need to send a payload when refreshing the list of OS images; some do not
    event: TanzuEventType,
    eventImportFileSuccess: TanzuEventType,
    eventImportFileFailure: TanzuEventType,
    osImageTooltipContent: string,
    nonTemplateAlertMessage?: string,
    noImageAlertMessage?: string,
}

@Directive()
export abstract class SharedOsImageStepDirective<IMAGE extends OsImage> extends StepFormDirective implements OnInit {
    static description = 'Specify the OS Image';

    // used by HTML as well as locally
    public providerInputs: OsImageProviderInputs;

    osImages: Array<IMAGE> = [];
    loadingOsTemplate: boolean = false;
    displayNonTemplateAlert: boolean = false;
    // TODO: It's questionable whether tkrVersion should be in this class, since it's only used for vSphere
    tkrVersion: Observable<string>;
    private stepMapping: StepMapping;

    constructor() {
        super();
        this.tkrVersion = AppServices.appDataService.getTkrVersion();
    }

    // This method allows child classes to supply the inputs (rather than having them passed as part of an HTML component tag).
    // This allows this step to follow the same pattern as all the other steps, which only take formGroup and formName as inputs.
    protected abstract supplyProviderInputs(): OsImageProviderInputs;

    private subscribeToProviderEvent() {
        // we register a handler for when our event receives data, namely that we'll populate our array of osImages
        AppServices.dataServiceRegistrar.stepSubscribe<IMAGE>(this, this.providerInputs.event, this.onOsImageEvent.bind(this));
    }

    private onOsImageEvent(images: Array<IMAGE>) {
        this.osImages = images ? images : [];
        this.loadingOsTemplate = false;
        this.restoreField(OsImageField.IMAGE, this.stepMapping, this.osImages);
    }

    protected getImageFromStoredValue(osImageValue: string): IMAGE {
        return this.osImages ? this.osImages.find(image => image.name === osImageValue) : null;
    }

    ngOnInit() {
        super.ngOnInit();
        this.stepMapping = this.createStepMapping();

        AppServices.userDataFormService.buildForm(this.formGroup, this.wizardName, this.formName, this.stepMapping);
        this.htmlFieldLabels = AppServices.fieldMapUtilities.getFieldLabelMap(this.stepMapping);
        this.storeDefaultLabels(this.stepMapping);

        this.providerInputs = this.supplyProviderInputs();
        this.registerStepDescriptionTriggers({fields: [OsImageField.IMAGE]});
        this.subscribeToProviderEvent();

        this.registerDefaultFileImportedHandler(this.providerInputs.eventImportFileSuccess, this.stepMapping);
        this.registerDefaultFileImportErrorHandler(this.providerInputs.eventImportFileFailure);
    }

    /**
     * @method retrieveOsImages
     * helper method to retrieve and preload list of available OS images from connected VC environment.
     * emits list of OS images to wizard-data.service
     */
    retrieveOsImages() {
        this.loadingOsTemplate = true;
        this.displayNonTemplateAlert = false;
        // some providers need to send a payload when refreshing the list of OS images; some do not
        const payload = this.providerInputs.createOsImageEventPayload ? this.providerInputs.createOsImageEventPayload() : undefined;
        AppServices.messenger.publish({
            type: this.providerInputs.event,
            payload
        });
    }

    /**
     * @method onOptionsSelected
     * helper method to determine if osImage.isTemplate is true or false; if false show warning
     */
    onOptionsSelected() {
        this.displayNonTemplateAlert = !this.formGroup.value[OsImageField.IMAGE].isTemplate;
    }

    dynamicDescription(): string {
        if (this.getFieldValue(OsImageField.IMAGE, true) && this.getFieldValue(OsImageField.IMAGE).name) {
            return 'OS Image: ' + this.getFieldValue(OsImageField.IMAGE).name;
        }
        return SharedOsImageStepDirective.description;
    }

    protected storeUserData() {
        this.storeUserDataFromMapping(this.stepMapping);
        this.storeDefaultDisplayOrder(this.stepMapping);
    }

    protected supplyStepMapping(): StepMapping {
        return OsImageStepMapping;
    }

    private createStepMapping(): StepMapping {
        const result = this.supplyStepMapping();
        const osImageFieldMapping = AppServices.fieldMapUtilities.getFieldMapping(OsImageField.IMAGE, result);
        // The retriever is a closure that can return an object based on the key
        // By setting the retriever in the field mapping, we allow buildForm to "call us back" to get the osImage using the saved key
        osImageFieldMapping.retriever = this.getImageFromStoredValue.bind(this);
        return result;
    }
}
