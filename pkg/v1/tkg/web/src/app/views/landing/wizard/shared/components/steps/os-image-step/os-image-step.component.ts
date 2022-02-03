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
    event: TanzuEventType,
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

    constructor() {
        super();
        this.tkrVersion = AppServices.appDataService.getTkrVersion();
    }

    // This method allows child classes to supply the inputs (rather than having them passed as part of an HTML component tag).
    // This allows this step to follow the same pattern as all the other steps, which only take formGroup and formName as inputs.
    protected abstract supplyProviderInputs(): OsImageProviderInputs;
    protected abstract supplyImportFileSuccessEvent(): TkgEventType;
    protected abstract supplyImportFileFailureEvent(): TkgEventType;

    private subscribeToProviderEvent() {
        // we register a handler for when our event receives data, namely that we'll populate our array of osImages
        AppServices.dataServiceRegistrar.stepSubscribe<IMAGE>(this, this.providerInputs.event, this.onOsImageEvent.bind(this));
    }

    private onOsImageEvent(images: Array<IMAGE>) {
        this.osImages = images ? images : [];
        this.loadingOsTemplate = false;
        if (this.osImages.length === 1) {
            this.setControlValueSafely(OsImageField.IMAGE, images[0]);
        } else {
            const fieldMapping = AppServices.fieldMapUtilities.getFieldMapping(WizardField.OSIMAGE, this.supplyStepMapping());
            AppServices.userDataFormService.restoreField(this.createUserDataIdentifier(fieldMapping.name), fieldMapping,
                this.formGroup, {}, this.getImageFromStoredValue.bind(this));
        }
    }

    protected getImageFromStoredValue(osImageValue: string): IMAGE {
        return this.osImages ? this.osImages.find(image => image.name === osImageValue) : null;
    }

    private getObjectRetrievalMap(): Map<string, (string) => any> {
        const objectRetrievalMap = new Map<string, (string) => any>();
        objectRetrievalMap.set('osImage', this.getImageFromStoredValue.bind(this));
        return objectRetrievalMap;
    }

    ngOnInit() {
        super.ngOnInit();
        // The objectRetrievalMap associates a field with a closure that can return an object based on a key
        // By sending the objectRetrievalMap to buildForm(), we allow buildForm to "call us back" to get the osImage using the saved key
        AppServices.userDataFormService.buildForm(this.formGroup, this.wizardName, this.formName, this.supplyStepMapping(),
            this.getObjectRetrievalMap());
        this.htmlFieldLabels = AppServices.fieldMapUtilities.getFieldLabelMap(this.supplyStepMapping());
        this.storeDefaultLabels(this.supplyStepMapping());
        this.registerDefaultFileImportedHandler(this.supplyImportFileSuccessEvent(), this.supplyStepMapping(),
            this.getObjectRetrievalMap());
        this.registerDefaultFileImportErrorHandler(this.supplyImportFileFailureEvent());

        this.providerInputs = this.supplyProviderInputs();
        this.registerStepDescriptionTriggers({fields: [OsImageField.IMAGE]});
        this.subscribeToProviderEvent();
    }

    /**
     * @method retrieveOsImages
     * helper method to retrieve and preload list of available OS images from connected VC environment.
     * emits list of OS images to wizard-data.service
     */
    retrieveOsImages() {
        this.loadingOsTemplate = true;
        this.displayNonTemplateAlert = false;
        AppServices.messenger.publish({
            type: this.providerInputs.event
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
        this.storeUserDataFromMapping(this.supplyStepMapping());
        this.storeDefaultDisplayOrder(this.supplyStepMapping());
    }

    protected supplyStepMapping(): StepMapping {
        return OsImageStepMapping;
    }
}
