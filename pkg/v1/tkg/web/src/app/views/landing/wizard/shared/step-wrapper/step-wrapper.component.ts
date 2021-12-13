import {
    AfterViewInit,
    Component,
    ComponentFactoryResolver,
    Injector,
    Input,
    OnInit,
    Type,
    ViewChild,
    ViewContainerRef
} from '@angular/core';
import { FormDataForHTML } from '../components/steps/form-utility';
import { FormGroup } from '@angular/forms';
import { WizardBaseDirective, WizardStepRegistrar } from '../wizard-base/wizard-base';
import { StepFormDirective } from '../step-form/step-form';

// The mission of this class is to (1) instantiate a step based on the passed "clazz" which is a component class, and
// (2) register the step with a "registrar" (which is basically a wizard implementing the registrar interface).
// This allows us to drive the creation of the step component via data (ie the passed component class),
// rather than having to specify each component in a separate HTML block in the wizard
@Component({
    selector: 'app-step-wrapper',
    templateUrl: './step-wrapper.component.html',
})
export class StepWrapperComponent implements AfterViewInit {
    // ViewChild annotation allows us to have a reference to our inner container where we want
    // to create the wrapped step component. It is defined in our HTML.
    @ViewChild('stepContainer', {read: ViewContainerRef}) stepContainer: ViewContainerRef;

    @Input() registrar: WizardStepRegistrar;    // the wizard that holds the steps we create
    @Input() formName: string;                  // name of the form for this step
    @Input() clazz: Type<StepFormDirective>;    // class of the backing component for the wrapped step

    constructor(private injector: Injector, private componentFactoryResolver: ComponentFactoryResolver) {
    }

    // NOTE: we initialize on ngAfterViewInit rather than ngOnInit because the stepContainer element is not
    // guaranteed to be available until ngAfterViewInit
    ngAfterViewInit() {
        this.initialize();
    }

    private initialize() {
        if (!this.clazz) {
            console.error('No clazz was given to StepWrapperComponent');
            return;
        }
        if (!this.registrar) {
            console.error('No registrar was given to StepWrapperComponent');
            return;
        }
        if (!this.formName || this.formName.length === 0) {
            console.error('No form name was given to StepWrapperComponent');
            return;
        }
        if (this.stepContainer === undefined) {
            console.warn(this.formName + ': there was no stepContainer element defined in the step-wrapper template HTML');
            return;
        }

        const componentFactory = this.componentFactoryResolver.resolveComponentFactory<StepFormDirective>(this.clazz);
        const wrappedComponent = this.stepContainer.createComponent<StepFormDirective>(componentFactory);

        this.registrar.registerStep(this.formName, wrappedComponent.instance);
        console.log('Registered ' + this.formName);
    }
}
