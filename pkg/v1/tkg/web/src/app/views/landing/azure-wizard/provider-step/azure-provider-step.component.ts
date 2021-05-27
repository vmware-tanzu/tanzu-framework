/**
 * Angular Modules
 */
import { Component, OnInit } from '@angular/core';
import { FormControl, Validators } from '@angular/forms';
import { ClrLoadingState } from "@clr/angular";
import { debounceTime, distinctUntilChanged, finalize, takeUntil } from 'rxjs/operators';

import { APIClient } from '../../../../swagger/api-client.service';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { TkgEventType } from '../../../../shared/service/Messenger';
import { AzureResourceGroup } from './../../../../swagger/models/azure-resource-group.model';
import { AzureWizardFormService } from 'src/app/shared/service/azure-wizard-form.service';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import Broker from 'src/app/shared/service/broker';

export const AzureAccountParamsKeys = ["tenantId", "clientId", "clientSecret", "subscriptionId"];
const extraFields = ["region", "sshPublicKey", "resourceGroupOption", "resourceGroupExisting"];
const optionalFields = ["resourceGroupCustom"];

@Component({
    selector: 'app-azure-provider-step',
    templateUrl: './azure-provider-step.component.html',
    styleUrls: ['./azure-provider-step.component.scss']
})
export class AzureProviderStepComponent extends StepFormDirective implements OnInit {

    loadingRegions = false;
    loadingState: ClrLoadingState = ClrLoadingState.DEFAULT;
    showOption = "existing";

    regions = [];
    azureClouds = [
        {
            name: 'AzurePublicCloud',
            displayName: 'Public Cloud'
        },
        {
            name: 'AzureUSGovernmentCloud',
            displayName: 'US Government Cloud'
        }
    ];
    resourceGroups = [];
    validCredentials = false;

    resourceGroupCreationState = "create";
    resourceGroupSelection: string = 'disabled';

    constructor(
        private apiClient: APIClient,
        private wizardFormService: AzureWizardFormService,
        private validationService: ValidationService) {
        super();
    }

    /**
     * Create the initial form
     */
    private buildForm() {
        const fields = AzureAccountParamsKeys.concat(extraFields);
        fields.forEach(key => this.formGroup.addControl(
            key,
            new FormControl('', [
                Validators.required
            ])
        ));

        this.formGroup.get('resourceGroupOption').setValue(this.showOption);

        optionalFields.forEach(key => this.formGroup.addControl(
            key,
            new FormControl('', [])
        ));

        this.formGroup.addControl(
            "azureCloud",
            new FormControl('AzurePublicCloud', [Validators.required]));

        this.formGroup['canMoveToNext'] = () => {
            return this.formGroup.valid && this.validCredentials;
        }
    }

    /**
     * Set the hidden form field to proper value based on form validity
     * @param valid whether we want the form to be valid
     */
    setValidCredentials(valid) {
        this.validCredentials = valid;
    }

    /**
     * Initialize the form with data from the backend
     * @param credentials Azure credentials
     * @param regions Azure regions
     */
    private initForm() {
        this.initAzureCredentials();
    }

    ngOnInit() {
        super.ngOnInit();

        this.buildForm();
        this.initForm();

        this.wizardFormService.getErrorStream(TkgEventType.AZURE_GET_RESOURCE_GROUPS)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(error => {
                this.errorNotification = error;
            });

        this.wizardFormService.getDataStream(TkgEventType.AZURE_GET_RESOURCE_GROUPS)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((rgs: AzureResourceGroup[]) => {
                this.resourceGroupSelection = null;
                this.resourceGroups = rgs;
            });

        this.formGroup.valueChanges
            .pipe(
                debounceTime(200),
                distinctUntilChanged((prev, curr) => {
                    for (const key of AzureAccountParamsKeys) {
                        if (prev[key] !== curr[key]) {
                            return false;
                        }
                    }
                    return true;
                }),
                takeUntil(this.unsubscribe)
            )
            .subscribe(
                () => {
                    this.validCredentials = false
                }
            );

        this.formGroup.get('region').valueChanges
            .pipe(
                distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
                takeUntil(this.unsubscribe)
            ).subscribe((val) => {
                this.onRegionChange(val)
            });

    }

    setSavedDataAfterLoad() {
        super.setSavedDataAfterLoad();
        // don't fill password field with ****
        this.formGroup.get('clientSecret').setValue('');
    }

    initAzureCredentials() {
        this.apiClient.getAzureEndpoint()
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(
                credentials => {
                    this.setAzureCredentialsValuesFromAPI(credentials);
                },
                () => {
                    this.errorNotification = 'Unable to retrieve Azure credentials';
                },
                () => { }
            );
    }

    setAzureCredentialsValuesFromAPI(credentials) {
        if (!this.hasSavedData()) {
            // init form values for Azure credentials
            for (const key of AzureAccountParamsKeys) {
                this.formGroup.get(key).setValue(credentials[key]);
            }
        }
    }

    getRegions() {
        this.loadingRegions = true;
        this.apiClient.getAzureRegions()
            .pipe(
                finalize(() => {
                    this.loadingRegions = false
                    const matchingRegions = this.regions.filter(r => r.displayName === this.getSavedValue('region', ''))
                    if (matchingRegions.length > 0) {
                        this.formGroup.get('region').setValue(matchingRegions[0].name);
                    }
                    // handle case where a new resource group was created, saved to local storage, and the browser was refreshed
                    if (this.getSavedValue('resourceGroupExisting', '') === ''
                        && this.regions.indexOf(this.getSavedValue('resourceGroupCustom', '') >= 0)) {
                        // select the newly created resource group in the existing resource group dropdown
                        this.formGroup.get('resourceGroupExisting').setValue(this.getSavedValue('resourceGroupCustom', ''))
                    }
                }),
                takeUntil(this.unsubscribe)
            )
            .subscribe(
                regions => {
                    this.regions = regions;
                },
                () => {
                    this.errorNotification = 'Unable to retrieve Azure regions';
                },
                () => { }
            );
    }

    /**
     * @method verifyCredentails
     * helper method to verify Azure connection credentials
     */
    verifyCredentails() {
        this.loadingState = ClrLoadingState.LOADING
        this.errorNotification = '';
        const params = {};
        for (const key of AzureAccountParamsKeys) {
            params[key] = this.formGroup.get(key).value;
        }
        this.apiClient.setAzureEndpoint({
            accountParams: params
        })
            .pipe(
                finalize(() => this.loadingState = ClrLoadingState.DEFAULT),
                takeUntil(this.unsubscribe)
            )
            .subscribe(
                (() => {
                    this.errorNotification = '';
                    this.setValidCredentials(true);
                    this.getRegions();
                }),
                (err => {
                    const error = err.error.message || err.message || JSON.stringify(err);
                    this.errorNotification = `${error}`;
                    this.setValidCredentials(false);
                    this.regions = [];
                    this.formGroup.get('region').setValue("");
                }),
                (() => {
                })
            );
    }

    /**
     * Whether to disable "Connect" button
     */
    isConnectDisabled() {
        return !AzureAccountParamsKeys.reduce((accu, key) => this.formGroup.get(key).valid && accu, true);
    }

    show(option) {
        this.showOption = option;
        if (option === "existing") {
            this.formGroup.controls['resourceGroupCustom'].clearValidators();
            this.formGroup.controls['resourceGroupCustom'].setValue('');
            this.formGroup.controls['resourceGroupExisting'].setValidators([
                Validators.required
            ]);
            this.clearFieldSavedData('resourceGroupCustom')
        } else if (option === "custom") {
            this.formGroup.controls['resourceGroupExisting'].clearValidators();
            this.formGroup.controls['resourceGroupExisting'].setValue('');
            this.formGroup.controls['resourceGroupCustom'].setValidators([
                Validators.required,
                this.validationService.isValidResourceGroupName(),
                this.validationService.isUniqueResourceGroupName(this.resourceGroups),
            ]);
            this.clearFieldSavedData('resourceGroupExisting')
        }
        this.formGroup.controls['resourceGroupCustom'].updateValueAndValidity();
        this.formGroup.controls['resourceGroupExisting'].updateValueAndValidity();
    }

    /**
     * Event handler when 'region' selection has changed
     */
    onRegionChange(val) {
        Broker.messenger.publish({
            type: TkgEventType.AZURE_REGION_CHANGED,
            payload: val
        });
    }

    /**
     * Update the "create" button if name has been changed.
     */
    onResourceGroupNameChange() {
        Broker.messenger.publish({
            type: TkgEventType.AZURE_RESOURCEGROUP_CHANGED,
            payload: this.formGroup.get('resourceGroupCustom').value
        });
    }

}
