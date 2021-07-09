import { Component, OnDestroy, Output, EventEmitter, ElementRef, ViewChild, AfterViewInit, Input } from '@angular/core';
import { ClrStepButton } from '@clr/angular';
import { distinctUntilChanged } from 'rxjs/operators';
import { FormMetaDataService } from 'src/app/shared/service/form-meta-data.service';

@Component({
    selector: 'app-step-controller',
    templateUrl: './step-controller.component.html',
    styleUrls: ['./step-controller.component.scss']
})
export class StepControllerComponent implements AfterViewInit, OnDestroy {

    @Output()
    nextStep = new EventEmitter();
    nativeElement;
    @ViewChild(ClrStepButton) nextButton: ClrStepButton;

    @Input()
    formName: string;

    isDisabled: boolean;

    constructor(
        el: ElementRef,
        private formMetaDataService: FormMetaDataService
    ) {
        this.nativeElement = el.nativeElement;
    }

    ngAfterViewInit(): void {
        // We overwrite the stepper's default "move to next" logic with our own.
        // You may provide your own "ready to move to next step" logic
        // via injecting a method ('canMoveToNext') to "formGroup" object.
        if (this.nextButton) {
            const w = this.nextButton;
            this.formName = w['clrStep'].id;
            w['navigateToNextPanel'] = () => {
                const move = w['clrStep']['formGroup']['canMoveToNext'] ?
                    w['clrStep']['formGroup']['canMoveToNext']() : w['clrStep']['formGroup'].valid;
                w['stepperService']['navigateToNextPanel'](w['clrStep'].id, move);
            }
            w['clrStep']['formGroup'].statusChanges.pipe(
                distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr))
            ).subscribe((val) => {
                this.isDisabled = val === 'INVALID';
            });
        }
    }

    /**
     * Find the step container of this element
     */
    findContainer() {
        let container = this.nativeElement;
        while (container) {
            const temp = container.querySelector("div.clr-accordion-inner-content");
            if (temp) {
                container = temp;
                break;
            }
            container = container.parentNode;
        }
        return container;
    }

    ngOnDestroy(): void {
        this.nextStep.emit(true);
        this.formMetaDataService.saveFormMetadata(this.formName, this.findContainer());
    }
}
