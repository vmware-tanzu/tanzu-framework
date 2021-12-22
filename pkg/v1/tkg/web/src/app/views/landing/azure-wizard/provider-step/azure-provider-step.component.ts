/**
 * Angular Modules
 */
import { Component, OnInit } from '@angular/core';
import { FormControl, Validators } from '@angular/forms';
import { ClrLoadingState } from "@clr/angular";
import { debounceTime, distinctUntilChanged, finalize, takeUntil } from 'rxjs/operators';

import { APIClient } from '../../../../swagger/api-client.service';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { TkgEvent, TkgEventType } from '../../../../shared/service/Messenger';
import { AzureResourceGroup } from './../../../../swagger/models/azure-resource-group.model';
import { AzureWizardFormService } from 'src/app/shared/service/azure-wizard-form.service';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import Broker from 'src/app/shared/service/broker';
import { FormMetaDataStore } from "../../wizard/shared/FormMetaDataStore";
import {NotificationTypes} from "../../../../shared/components/alert-notification/alert-notification.component";
import { AzureField, ResourceGroupOption } from '../azure-wizard.constants';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { AzureProviderStepMapping } from './azure-provider-step.fieldmapping';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';

enum ProviderField {
    AZURECLOUD = 'azureCloud',
    CLIENT = 'clientId',
    CLIENTSECRET = 'clientSecret',
    SUBSCRIPTION = 'subscriptionId',
    TENANT = 'tenantId',
    SSHPUBLICKEY = 'sshPublicKey',
    REGION = 'region',
    RESOURCEGROUPOPTION = 'resourceGroupOption',
    RESOURCEGROUPEXISTING = 'resourceGroupExisting',
    RESOURCEGROUPCUSTOM = 'resourceGroupCustom',
}

// NOTE: the keys of AzureAccountParamsKeys values are used by backend endpoints, so don't change them
export const AzureAccountParamsKeys = [AzureField.PROVIDER_TENANT, AzureField.PROVIDER_CLIENT,
    AzureField.PROVIDER_CLIENTSECRET, AzureField.PROVIDER_SUBSCRIPTION, AzureField.PROVIDER_AZURECLOUD];
const requiredFields = [AzureField.PROVIDER_REGION, AzureField.PROVIDER_SSHPUBLICKEY, AzureField.PROVIDER_RESOURCEGROUPOPTION,
    AzureField.PROVIDER_RESOURCEGROUPEXISTING];
const optionalFields = [AzureField.PROVIDER_RESOURCEGROUPCUSTOM];

@Component({
    selector: 'app-azure-provider-step',
    templateUrl: './azure-provider-step.component.html',
    styleUrls: ['./azure-provider-step.component.scss']
})
export class AzureProviderStepComponent extends StepFormDirective implements OnInit {
    successImportFile: string;

    loadingRegions = false;
    loadingState: ClrLoadingState = ClrLoadingState.DEFAULT;
    resourceGroupOption = ResourceGroupOption.EXISTING;

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

    constructor(private apiClient: APIClient,
                private fieldMapUtilities: FieldMapUtilities,
                private wizardFormService: AzureWizardFormService,
                private validationService: ValidationService) {
        super();
    }

    private customizeForm() {
        this.formGroup['canMoveToNext'] = () => {
            return this.formGroup.valid && this.validCredentials;
        }

        this.wizardFormService.getErrorStream(TkgEventType.AZURE_GET_RESOURCE_GROUPS)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(error => {
                this.errorNotification = error;
            });

        this.wizardFormService.getDataStream(TkgEventType.AZURE_GET_RESOURCE_GROUPS)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((azureResourceGroups: AzureResourceGroup[]) => {
                this.resourceGroups = azureResourceGroups;
                if (azureResourceGroups.length === 1) {
                    this.formGroup.get(AzureField.PROVIDER_RESOURCEGROUPEXISTING).setValue(azureResourceGroups[0].name);
                } else {
                    this.initResourceGroupFromSavedData();
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

        this.formGroup.get(AzureField.PROVIDER_REGION).valueChanges
            .pipe(
                distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
                takeUntil(this.unsubscribe)
            ).subscribe((val) => {
            this.onRegionChange(val)
        });

        Broker.messenger.getSubject(TkgEventType.CONFIG_FILE_IMPORTED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data: TkgEvent) => {
                this.configFileNotification = {
                    notificationType: NotificationTypes.SUCCESS,
                    message: data.payload
                };
                // The file import saves the data to local storage, so we reinitialize this step's form from there
                this.savedMetadata = FormMetaDataStore.getMetaData(this.formName);
                this.initFormWithImportedData();

                // Clear event so that listeners in other provider workflows do not receive false notifications
                Broker.messenger.clearEvent(TkgEventType.CONFIG_FILE_IMPORTED);
            });
    }

    /**
     * Set the hidden form field to proper value based on form validity
     * @param valid whether we want the form to be valid
     */
    setValidCredentials(valid) {
        this.validCredentials = valid;
    }

    ngOnInit() {
        super.ngOnInit();

        this.fieldMapUtilities.buildForm(this.formGroup, this.formName, AzureProviderStepMapping);
        this.customizeForm();

        this.initAzureCredentials();
        this.initFormWithSavedData();
    }

    private initResourceGroupFromSavedData() {
        // if the user did an import, then we expect the value to be stored in AzureField.PROVIDER_RESOURCEGROUPCUSTOM
        // we'll check and see if that value is now existing
        let savedGroupExisting = this.getSavedValue(AzureField.PROVIDER_RESOURCEGROUPEXISTING, '');
        let savedGroupCustom = this.getSavedValue(AzureField.PROVIDER_RESOURCEGROUPCUSTOM, '');

        if (this.handleIfSavedCustomResourceGroupIsNowExisting(savedGroupCustom)) {
            savedGroupExisting = savedGroupCustom;
            savedGroupCustom = '';
        }

        if (savedGroupCustom !== '') {
            this.formGroup.get(AzureField.PROVIDER_RESOURCEGROUPCUSTOM).setValue(savedGroupCustom);
            this.showResourceGroup(ResourceGroupOption.CUSTOM);
        } else if (savedGroupExisting !== '') {
            this.formGroup.get(AzureField.PROVIDER_RESOURCEGROUPEXISTING).setValue(savedGroupExisting);
            this.showResourceGroup(ResourceGroupOption.EXISTING);
        } else {
            this.showResourceGroup(this.resourceGroupOption);
        }
    }

    initFormWithSavedData() {
        // rather than call our parent class' initFormWithSavedData() which would initialize ALL the fields on this form,
        // we only want to initialize the credential fields (and ssh key) and then have the user connect to the server,
        // which will populate our data arrays. We need to connect to the server before being able to set
        // other fields (e.g. to pick a region, the listbox must be populated from the data array).
        if (this.hasSavedData()) {
            AzureAccountParamsKeys.forEach( accountField => {
                this.initFieldWithSavedData(accountField);
            });
            this.initFieldWithSavedData(AzureField.PROVIDER_SSHPUBLICKEY);
        }
        this.scrubPasswordField(AzureField.PROVIDER_CLIENTSECRET);
        if (this.getFieldValue(AzureField.PROVIDER_AZURECLOUD) === '') {
            this.setFieldValue(AzureField.PROVIDER_AZURECLOUD, this.azureClouds[0].name);
        }
    }

    initFormWithImportedData() {
        console.log('azure-provider-step.initFormWithSavedDataCOPY()');
        // Initializations not needed the first time the form is loaded, but
        // required to re-initialize after form has been used
        this.validCredentials = false;
        this.regions = [];
        this.resourceGroups = [];
        this.resourceGroupCreationState = 'create';
        this.resourceGroupOption = ResourceGroupOption.EXISTING;

        [   AzureField.PROVIDER_TENANT,
            AzureField.PROVIDER_CLIENT,
            AzureField.PROVIDER_SUBSCRIPTION,
            AzureField.PROVIDER_AZURECLOUD
        ].forEach( accountField => {
            this.initFieldWithSavedData(accountField);
        });
        this.initFieldWithSavedData(AzureField.PROVIDER_SSHPUBLICKEY);

        // AzureField.PROVIDER_CLIENTSECRET causes us to break
        // this.initFieldWithSavedData(AzureField.PROVIDER_CLIENTSECRET);
        this.scrubPasswordField(AzureField.PROVIDER_CLIENTSECRET);

        if (this.getFieldValue(AzureField.PROVIDER_AZURECLOUD) === '') {
            this.setFieldValue(AzureField.PROVIDER_AZURECLOUD, this.azureClouds[0].name);
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
            for (const accountParamField of AzureAccountParamsKeys) {
                this.setControlValueSafely(accountParamField, credentials[accountParamField]);
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
                    const selectedRegion = this.regions.length === 1 ? this.regions[0].name : this.getSavedValue(AzureField.PROVIDER_REGION, '');
                    // setting the region value will trigger other data calls to the back end for resource groups, osimages, etc
                    this.setControlValueSafely(AzureField.PROVIDER_REGION, selectedRegion);
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
                    this.setControlValueSafely(AzureField.PROVIDER_REGION, '');
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
        if (option === ResourceGroupOption.EXISTING) {
            this.formGroup.controls[AzureField.PROVIDER_RESOURCEGROUPCUSTOM].clearValidators();
            this.formGroup.controls[AzureField.PROVIDER_RESOURCEGROUPCUSTOM].setValue('');
            this.formGroup.controls[AzureField.PROVIDER_RESOURCEGROUPEXISTING].setValidators([
                Validators.required
            ]);
            this.clearFieldSavedData(AzureField.PROVIDER_RESOURCEGROUPCUSTOM)
        } else if (option === ResourceGroupOption.CUSTOM) {
            this.formGroup.controls[AzureField.PROVIDER_RESOURCEGROUPEXISTING].clearValidators();
            this.formGroup.controls[AzureField.PROVIDER_RESOURCEGROUPEXISTING].setValue('');
            this.formGroup.controls[AzureField.PROVIDER_RESOURCEGROUPCUSTOM].setValidators([
                Validators.required,
                this.validationService.isValidResourceGroupName(),
                this.validationService.isUniqueResourceGroupName(this.resourceGroups),
            ]);
            this.clearFieldSavedData(AzureField.PROVIDER_RESOURCEGROUPEXISTING)
        } else {
            console.log('WARNING: showResourceGroup() received unrecognized value of ' + option);
        }
        this.formGroup.controls[AzureField.PROVIDER_RESOURCEGROUPCUSTOM].updateValueAndValidity();
        this.formGroup.controls[AzureField.PROVIDER_RESOURCEGROUPEXISTING].updateValueAndValidity();
    }

    /**
     * Event handler when AzureField.PROVIDER_REGION selection has changed
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
            payload: this.formGroup.get(AzureField.PROVIDER_RESOURCEGROUPCUSTOM).value
        });
    }

    private handleIfSavedCustomResourceGroupIsNowExisting(savedGroupCustom: string): boolean {
        // handle case where user originally created a new (custom) resource group (and value was either saved
        // to local storage or to config file as a custom resource group), but now when the data is restored,
        // the resource group exists (so we should move the custom value over to the existing data slot).
        const customIsNowExisting = this.resourceGroupContains(savedGroupCustom);
        if (customIsNowExisting) {
            this.clearFieldSavedData(AzureField.PROVIDER_RESOURCEGROUPCUSTOM);
            this.saveFieldData(AzureField.PROVIDER_RESOURCEGROUPEXISTING, savedGroupCustom);
            return true;
        }
        return false;
    }

    private resourceGroupContains(resourceGroupName: string) {
        return this.resourceGroups.find( resourceGroup => { return resourceGroup.name === resourceGroupName; });
    }
    protected dynamicDescription(): string {
        const tenant = this.getFieldValue('tenantId', true);
        return tenant ? `Azure tenant: ${tenant}` : 'Validate the Azure provider credentials for Tanzu';
    }
}
