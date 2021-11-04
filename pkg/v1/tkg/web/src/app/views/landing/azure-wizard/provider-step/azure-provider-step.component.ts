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

export const AzureAccountParamsKeys = ["tenantId", "clientId", "clientSecret", "subscriptionId", "azureCloud"];
const requiredFields = ["region", "sshPublicKey", "resourceGroupOption", "resourceGroupExisting"];
const optionalFields = ["resourceGroupCustom"];

@Component({
    selector: 'app-azure-provider-step',
    templateUrl: './azure-provider-step.component.html',
    styleUrls: ['./azure-provider-step.component.scss']
})
export class AzureProviderStepComponent extends StepFormDirective implements OnInit {

    loadingRegions = false;
    loadingState: ClrLoadingState = ClrLoadingState.DEFAULT;
    resourceGroupOption = 'existing';

    regions = [];
    // NOTE: order is important here; we default to the first cloud in the azureClouds array
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

    resourceGroupCreationState = 'create';
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
        AzureAccountParamsKeys.concat(requiredFields).forEach(controlName => this.formGroup.addControl(
            controlName,
            new FormControl('', [
                Validators.required
            ])
        ));

        this.formGroup.get('resourceGroupOption').setValue(this.resourceGroupOption);

        optionalFields.forEach(controlName => this.formGroup.addControl(
            controlName,
            new FormControl('', [])
        ));

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
            .subscribe((azureResourceGroups: AzureResourceGroup[]) => {
                this.resourceGroupSelection = null; // SHIMON SEZ: we should be able to take this out
                this.resourceGroups = azureResourceGroups;
                this.initResourceGroupFromSavedData();
                if (azureResourceGroups.length === 1) {
                    this.formGroup.get('resourceGroupExisting').setValue(azureResourceGroups[0].name);
                }
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

        this.initFormWithSavedData();
    }

    private initResourceGroupFromSavedData() {
        // if the user did an import, then we expect the value to be stored in 'resourceGroupCustom'
        // we'll check and see if that value is now existing
        let savedGroupExisting = this.getSavedValue('resourceGroupExisting', '');
        let savedGroupCustom = this.getSavedValue('resourceGroupCustom', '');

        if (this.handleIfSavedCustomResourceGroupIsNowExisting(savedGroupCustom)) {
            savedGroupExisting = savedGroupCustom;
            savedGroupCustom = '';
        }

        if (savedGroupCustom !== '') {
            this.formGroup.get('resourceGroupCustom').setValue(savedGroupCustom);
            this.showResourceGroup('custom');
        } else if (savedGroupExisting !== '') {
            this.formGroup.get('resourceGroupExisting').setValue(savedGroupExisting);
            this.showResourceGroup('existing');
        } else {
            this.showResourceGroup(this.resourceGroupOption);
        }
    }

    initFormWithSavedData() {
        console.log('azure-provider-step.initFormWithSavedData()');
        // Initializations not needed the first time the form is loaded, but
        // required to re-initialize after form has been used
        this.validCredentials = false;
        this.regions = [];
        this.resourceGroups = [];
        this.resourceGroupCreationState = "create";
        this.resourceGroupSelection = 'disabled';
        this.resourceGroupOption = 'existing';

        // rather than call our parent class' initFormWithSavedData to initalize ALL the fields on this form,
        // we only want to initialize the credential fields (and ssh key) and then have the user connect to the server,
        // which will populate our data arrays. We need to connect to the server before being able to set
        // other fields (e.g. to pick a region, the listbox must be populated from the data array).
        if (this.hasSavedData()) {
            AzureAccountParamsKeys.forEach( accountField => {
                this.initFieldWithSavedData(accountField);
            });
            this.initFieldWithSavedData('sshPublicKey');
        }
        this.scrubPasswordField('clientSecret');
        if (this.getFieldValue('azureCloud') === '') {
            this.setFieldValue('azureCloud', this.azureClouds[0].name);
        }
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

    private setAzureCredentialsValuesFromAPI(credentials) {
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
                    this.loadingRegions = false;
                }),
                takeUntil(this.unsubscribe)
            )
            .subscribe(
                regions => {
                    this.regions = regions.sort((regionA, regionB) => regionA.name.localeCompare(regionB.name));
                    const selectedRegion = this.regions.length === 1 ? this.regions[0].name : this.getSavedValue('region', '');
                    // setting the region value will trigger other data calls to the back end for resource groups, osimages, etc
                    this.formGroup.get('region').setValue(selectedRegion);
                },
                () => {
                    this.errorNotification = 'Unable to retrieve Azure regions';
                },
                () => { }
            );
    }

    /**
     * @method verifyCredentials
     * helper method to verify Azure connection credentials
     */
    verifyCredentials() {
        this.loadingState = ClrLoadingState.LOADING
        this.errorNotification = '';
        const accountParams = {};
        for (const key of AzureAccountParamsKeys) {
            accountParams[key] = this.formGroup.get(key).value;
        }
        this.apiClient.setAzureEndpoint({
            accountParams: accountParams
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

    showResourceGroup(option) {
        this.resourceGroupOption = option;
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
        } else {
            console.log('WARNING: showResourceGroup() received unrecognized value of ' + option);
        }
        this.formGroup.controls['resourceGroupCustom'].updateValueAndValidity();
        this.formGroup.controls['resourceGroupExisting'].updateValueAndValidity();
    }

    /**
     * Event handler when 'region' selection has changed
     */
    onRegionChange(val) {
        console.log('azure-provider-step.onRegionChange() detects region change to ' + val + '; publishing AZURE_REGION_CHANGED');
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

    private handleIfSavedCustomResourceGroupIsNowExisting(savedGroupCustom: string): boolean {
        // handle case where user originally created a new (custom) resource group (and value was either saved
        // to local storage or to config file as a custom resource group), but now when the data is restored,
        // the resource group exists (so we should move the custom value over to the existing data slot).
        const customIsNowExisting = this.resourceGroupContains(savedGroupCustom);
        if (customIsNowExisting) {
            this.clearFieldSavedData('resourceGroupCustom');
            this.saveFieldData('resourceGroupExisting', savedGroupCustom);
            return true;
        }
        return false;
    }

    private resourceGroupContains(resourceGroupName: string) {
        return this.resourceGroups.find( resourceGroup => { return resourceGroup.name === resourceGroupName; });
    }
}
