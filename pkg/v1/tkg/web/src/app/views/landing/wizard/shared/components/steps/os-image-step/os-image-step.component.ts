// App imports
import AppServices from '../../../../../../../shared/service/appServices';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { Observable } from 'rxjs/internal/Observable';
import { OsImageStepMapping } from './os-image-step.fieldmapping';
import { StepFormDirective } from '../../../step-form/step-form';
import { TkgEventType } from 'src/app/shared/service/Messenger';
import { Directive, OnInit } from '@angular/core';

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
    event: TkgEventType,
    osImageTooltipContent: string,
    nonTemplateAlertMessage?: string,
    noImageAlertMessage?: string,
}
@Directive()
export abstract class SharedOsImageStepDirective<IMAGE extends OsImage> extends StepFormDirective implements OnInit {
    static description = 'Specify the OS Image';

    // used by HTML as well as locally
    public providerInputs: OsImageProviderInputs;

    osImages: Array<IMAGE>;
    loadingOsTemplate: boolean = false;
    displayNonTemplateAlert: boolean = false;
    // TODO: It's questionable whether tkrVersion should be in this class, since it's only used for vSphere
    tkrVersion: Observable<string>;

    protected constructor(protected fieldMapUtilities: FieldMapUtilities) {
        super();
        this.tkrVersion = AppServices.appDataService.getTkrVersion();
    }

    // This method allows child classes to supply the inputs (rather than having them passed as part of an HTML component tag).
    // This allows this step to follow the same pattern as all the other steps, which only take formGroup and formName as inputs.
    protected abstract supplyProviderInputs(): OsImageProviderInputs;

    private subscribeToProviderEvent() {
        // we register a handler for when our event receives data, namely that we'll populate our array of osImages
        AppServices.dataServiceRegistrar.stepSubscribe<IMAGE>(this, this.providerInputs.event, this.onFetchedOsImages.bind(this));
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

    ngOnInit() {
        super.ngOnInit();
        this.fieldMapUtilities.buildForm(this.formGroup, this.formName, OsImageStepMapping);
        this.providerInputs = this.supplyProviderInputs();
        this.registerFieldsAffectingStepDescription(['osImage']);
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
        AppServices.messenger.publish({
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
        return SharedOsImageStepDirective.description;
    }
}
