import { Component, Input, OnChanges, OnDestroy, SimpleChanges } from '@angular/core';
import { FormArray, FormGroup } from '@angular/forms';
import { Subject } from 'rxjs';
import { distinctUntilChanged, takeUntil } from 'rxjs/operators';
import { ValidatorEnum } from '../../../constants/validation.constants';
import { FieldMapping } from '../../../field-mapping/FieldMapping';
import { FormUtils } from '../../../utils/form-utils';
import { TKGLabelsConfig } from './interfaces/tkg-labels.interface';

@Component({
    selector: 'app-tkg-labels',
    templateUrl: './tkg-labels.component.html',
    styleUrls: ['./tkg-labels.component.scss']
})
export class TkgLabelsComponent implements OnChanges, OnDestroy {
    @Input() config: TKGLabelsConfig;

    validatorEnum = ValidatorEnum;

    stopSubscriptions$ = new Subject<void>();

    get labelsFormArray(): FormArray {
        return this.config.forms.control as FormArray;
    }

    ngOnChanges(changes: SimpleChanges): void {
        if (changes.config.previousValue !== changes.config.currentValue && changes.config.currentValue) {
            this.labelsFormArray.valueChanges
                .pipe(
                    takeUntil(this.stopSubscriptions$),
                    distinctUntilChanged((k, v) => JSON.stringify(k) === JSON.stringify(v))
                )
                .subscribe(() => {
                    this.labelsFormArray.controls.forEach((label: FormGroup) => {
                        (this.config.fields.fieldMapping as FieldMapping).children.forEach((field) =>
                            label.get(field.name).updateValueAndValidity()
                        );
                    });
                });
        }
    }

    createLabel(): FormGroup {
        const labelFormGroup = new FormGroup({});
        (this.config.fields.fieldMapping as FieldMapping).children.forEach((fieldMapping) =>
            FormUtils.addDynamicControl(labelFormGroup, '', fieldMapping)
        );

        return labelFormGroup;
    }

    deleteLabel(index: number): void {
        this.labelsFormArray.removeAt(index);
    }

    addNewLabel(): void {
        this.labelsFormArray.markAllAsTouched();

        if (this.labelsFormArray.invalid) {
            return;
        }
        this.labelsFormArray.push(this.createLabel());
    }

    ngOnDestroy(): void {
        this.stopSubscriptions$.next();
        this.stopSubscriptions$.complete();
    }
}
