// Angular imports
import { Component, OnInit } from '@angular/core';
import { Validators } from '@angular/forms';

// Library imports
import { distinctUntilChanged, takeUntil } from 'rxjs/operators';
import { AzureResourceGroup, AzureVirtualNetwork } from 'tanzu-ui-api-lib';

// App imports
import AppServices from 'src/app/shared/service/appServices';
import { AzureField } from '../azure-wizard.constants';
import { AzureVnetStandaloneStepMapping, AzureVnetStepMapping } from './vnet-step.fieldmapping';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { FormMetaDataStore } from '../../wizard/shared/FormMetaDataStore'
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { TkgEventType } from 'src/app/shared/service/Messenger';
import { ValidationService } from './../../wizard/shared/validation/validation.service';

const CUSTOM = "CUSTOM";
export const EXISTING = "EXISTING";

@Component({
    selector: 'app-vnet-step',
    templateUrl: './vnet-step.component.html',
    styleUrls: ['./vnet-step.component.scss']
})
export class VnetStepComponent extends StepFormDirective implements OnInit {
    region = '';    // Current region selected
    showVnetFieldsOption = EXISTING;
    customResourceGroup = null;

    // An object maps vnet to subsets; data retrieved from backend
    vnetSubnets = {};

    /** lists to be retrieved from the backend */
    vnetResourceGroups = [];
    vnetNamesExisting = [];
    controlPlaneSubnets = [];
    workerNodeSubnets = [];

    defaultVnetCidr: string = '10.0.0.0/16';
    defaultControlPlaneCidr: string = '10.0.0.0/24';
    defaultWorkerNodeCidr: string = '10.0.1.0/24';

    createPrivateCluster = false;
    // cidrForPrivateCluster holds two CIDR values, one for EXISTING and one for CUSTOM;
    // used to validate private cluster IP address (if nec)
    // also displayed on the page as instruction, if user chooses to create a private cluster
    cidrForPrivateCluster = {};

    /** UI fields to expose per edition */
    vnetFieldsExisting: Array<string> = [];
    vnetFieldsNew: Array<string> = [];

    constructor(private fieldMapUtilities: FieldMapUtilities,
                private validationService: ValidationService) {
        super();
    }

    private supplyStepMapping(): StepMapping {
        return this.modeClusterStandalone ? AzureVnetStandaloneStepMapping : AzureVnetStepMapping;
    }

    /**
     * Create the initial form
     */
    private subscribeToServices() {
        AppServices.dataServiceRegistrar.stepSubscribe(this, TkgEventType.AZURE_GET_VNETS, this.setVnets.bind(this))
        AppServices.dataServiceRegistrar.stepSubscribe(this, TkgEventType.AZURE_GET_RESOURCE_GROUPS,
            this.onFetchedResourceGroups.bind(this));
    }

    private onFetchedResourceGroups(azureResourceGroups: AzureResourceGroup[]) {
        this.vnetResourceGroups = azureResourceGroups;
        if (this.customResourceGroup) {
            this.setControlValueSafely(AzureField.VNET_RESOURCE_GROUP, this.customResourceGroup);
        } else if (azureResourceGroups.length === 1) {
            this.setControlValueSafely(AzureField.VNET_RESOURCE_GROUP, azureResourceGroups[0].name);
        } else {
            this.setControlWithSavedValue(AzureField.VNET_RESOURCE_GROUP);
        }
    }

    private customizeForm() {
        // TODO: consider morphing these .valueChanges calls into registerOnValueChange() calls
        this.formGroup.get(AzureField.VNET_RESOURCE_GROUP).valueChanges
            .pipe(
                distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
                takeUntil(this.unsubscribe)
            ).subscribe((val) => {
                this.onResourceGroupChange(val);
            });

        this.formGroup.get(AzureField.VNET_EXISTING_NAME).valueChanges
            .pipe(
                distinctUntilChanged((prev, curr) => prev === curr),
                takeUntil(this.unsubscribe)
            ).subscribe(newValue => {
                this.onVnetChange(newValue);
            });

        this.formGroup.get(AzureField.VNET_CUSTOM_CIDR).valueChanges
            .pipe(
                distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
                takeUntil(this.unsubscribe)
            ).subscribe((cidr) => {
                AppServices.messenger.publish({
                    type: TkgEventType.NETWORK_STEP_GET_NO_PROXY_INFO,
                    payload: { info: (cidr ? cidr + ',' : '') + '169.254.0.0/16,168.63.129.16' }
                });
                this.triggerStepDescriptionChange();
            });
        /**
         * Whenever Azure region selection changes...
         */
        AppServices.messenger.getSubject(TkgEventType.AZURE_REGION_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                this.onRegionChange(event.payload);
            });

        AppServices.messenger.getSubject(TkgEventType.AZURE_RESOURCEGROUP_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                this.customResourceGroup = event.payload;
            });

        this.registerOnValueChange(AzureField.VNET_PRIVATE_CLUSTER, this.onCreatePrivateAzureCluster.bind(this));
        this.registerOnValueChange(AzureField.VNET_CONTROLPLANE_NEWSUBNET_CIDR, this.onControlPlaneSubnetCidrNewChange.bind(this));
        this.registerOnValueChange(AzureField.VNET_CONTROLPLANE_SUBNET_NAME, this.onControlPlaneSubnetChange.bind(this));
        this.registerOnValueChange(AzureField.VNET_EXISTING_OR_CUSTOM, this.onExistingOrCustomOptionChange.bind(this));
    }

    ngOnInit() {
        super.ngOnInit();
        this.fieldMapUtilities.buildForm(this.formGroup, this.formName, this.supplyStepMapping());
        this.subscribeToServices();
        this.customizeForm();

        this.vnetFieldsExisting = [AzureField.VNET_EXISTING_NAME, AzureField.VNET_CONTROLPLANE_SUBNET_NAME];
        this.vnetFieldsNew =
            [
                AzureField.VNET_CUSTOM_NAME,
                AzureField.VNET_CUSTOM_CIDR,
                AzureField.VNET_CONTROLPLANE_NEWSUBNET_NAME,
                AzureField.VNET_CONTROLPLANE_NEWSUBNET_CIDR,
            ];
        if (!this.modeClusterStandalone) {
            this.vnetFieldsExisting.push(AzureField.VNET_WORKER_SUBNET_NAME);
            this.vnetFieldsNew.push(AzureField.VNET_WORKER_NEWSUBNET_NAME, AzureField.VNET_WORKER_NEWSUBNET_CIDR);
        }

        this.initFormWithSavedData();
    }

    onControlPlaneSubnetChange(name: string) {
        // when the user selects an existing subnet, we look up the associated CIDR and set a hidden field with the CIDR value,
        // which is later used in creating the AzureRegionalClusterParams payload object
        const subnetEntry = this.controlPlaneSubnets.find(subnet => subnet.name === name);
        const cidrOfSelectedControlPlaneSubnet = (subnetEntry) ? subnetEntry.cidr : '';

        this.setControlValueSafely(AzureField.VNET_CONTROLPLANE_SUBNET_CIDR, cidrOfSelectedControlPlaneSubnet);
        this.cidrForPrivateCluster[EXISTING] = cidrOfSelectedControlPlaneSubnet;
    }

    onControlPlaneSubnetCidrNewChange(value: string) {
        this.cidrForPrivateCluster[CUSTOM] = value;
    }

    onCreatePrivateAzureCluster(createPrivateCluster: boolean) {
        this.createPrivateCluster = createPrivateCluster;
        if (createPrivateCluster) {
            const cidr = this.cidrForPrivateCluster['' + this.showVnetFieldsOption];
            const cidrValidator = this.validationService.isIpInSubnet2(cidr);
            this.formGroup.markAsPending();
            this.resurrectFieldWithSavedValue(AzureField.VNET_PRIVATE_IP,
                [Validators.required, this.validationService.isValidIpOrFqdn(),
                    cidrValidator]
            );
        } else {
            this.disarmField(AzureField.VNET_PRIVATE_IP, true);
        }
    }

    initFormWithSavedData() {
        const customResourceGroup1 = FormMetaDataStore.getMetaDataItem("providerForm", "resourceGroupCustom");
        this.customResourceGroup = customResourceGroup1 ? customResourceGroup1.displayValue : '';

        this.setControlWithSavedValue(AzureField.VNET_PRIVATE_CLUSTER, false);

        // if the user did an import, then we expect there may be a custom vnet value to be stored in AzureField.VNET_CUSTOM_NAME slot.
        // however, since that custom vnet may have been created and now be an existing vnet,
        // we need to call a special method to handle that situation, which moves the saved name from the custom slot
        // into the existing slot. Note that we do this before deciding whether to show the custom or existing options
        this.modifySavedValuesIfVnetCustomNameIsNowExisting();
        const savedVnetCustom = this.getSavedValue(AzureField.VNET_CUSTOM_NAME, '');
        const optionValue = savedVnetCustom !== '' ? CUSTOM : EXISTING;
        // NOTE: setting the EXISTING_OR_CUSTOM value will trigger the display to update
        this.setControlWithSavedValue(AzureField.VNET_EXISTING_OR_CUSTOM, optionValue);
    }

    onRegionChange(region: string) {
        this.region = region;
        this.setControlWithSavedValue(AzureField.VNET_RESOURCE_GROUP);
    }

    onResourceGroupChange(resourceGroupName) {
        if (resourceGroupName && resourceGroupName !== this.customResourceGroup) {
            AppServices.dataServiceRegistrar.trigger([TkgEventType.AZURE_GET_VNETS], { resourceGroupName, location: this.region })
        }
    }

    private setVnets(vnets: AzureVirtualNetwork[]) {
        this.vnetSubnets = vnets.reduce((accu, vnet) => { accu[vnet.name] = vnet.subnets; return accu; }, {});
        this.vnetNamesExisting = Object.keys(this.vnetSubnets);

        this.modifySavedValuesIfVnetCustomNameIsNowExisting();

        // NOTE: setting the EXISTING_NAME value will cause an update to the subnets
        const savedVnet = this.getSavedValue(AzureField.VNET_EXISTING_NAME, '');
        if (savedVnet && this.vnetNamesExisting.includes(savedVnet)) {
            this.setControlValueSafely(AzureField.VNET_EXISTING_NAME, savedVnet);
        } else if (this.vnetNamesExisting.length === 1) {
            this.setControlValueSafely(AzureField.VNET_EXISTING_NAME, this.vnetNamesExisting[0]);
        }
    }

    onVnetChange(vnetName) {
        // Use the same set of subnets for control plane and worker plane
        this.workerNodeSubnets = this.vnetSubnets[vnetName] || [];
        this.controlPlaneSubnets = this.workerNodeSubnets;

        if (this.workerNodeSubnets.length === 0) {
            const warning = 'WARNING: vnet ' + vnetName + ' appears to have no subnets available! vnetSubnets=' +
                JSON.stringify(this.vnetSubnets);
            console.log(warning);
        }
        this.initSubnetField(AzureField.VNET_CONTROLPLANE_SUBNET_NAME, this.controlPlaneSubnets);
        if (!this.modeClusterStandalone) {
            this.initSubnetField(AzureField.VNET_WORKER_SUBNET_NAME, this.workerNodeSubnets);
        }
    }

    // set subnet field with saved data if available, or default if only one subnet
    private initSubnetField(fieldName: string, subnets: any[]) {
        // set subnet fields with local storage data if available, or default if only one subnet
        const nameSavedSubnet = this.getSavedValue(fieldName, '');
        const savedControlPlaneSubnet = nameSavedSubnet ? this.findSubnetByName(nameSavedSubnet, subnets) : null;
        if (savedControlPlaneSubnet) {
            this.setControlValueSafely(fieldName, nameSavedSubnet);
        } else if (subnets.length === 1) {
            this.setControlValueSafely(fieldName, subnets[0].name);
        }
    }

    private findSubnetByName(subnetName: string, subnets: any[]): any {
        return subnets.find(s => s.name === subnetName);
    }

    private onExistingOrCustomOptionChange(option: string, clearSavedData = false) {
        this.showVnetFieldsOption = option;
        if (option === EXISTING) {
            this.initVnetFieldsExisting(clearSavedData);
            if (this.vnetResourceGroups.length === 1) {
                this.setControlValueSafely(AzureField.VNET_RESOURCE_GROUP, this.vnetResourceGroups[0].name);
            } else {
                this.setControlWithSavedValue(AzureField.VNET_RESOURCE_GROUP);
            }
        } else if (option === CUSTOM) {
            this.showVnetFieldsCustom(clearSavedData);
        }
        this.onCreatePrivateAzureCluster(this.createPrivateCluster);
    }

    private showVnetFieldsCustom(clearSavedData: boolean) {
        this.resurrectFieldWithSavedValue(AzureField.VNET_CUSTOM_NAME, [Validators.required]);
        this.resurrectFieldWithSavedValue(AzureField.VNET_CONTROLPLANE_NEWSUBNET_NAME, [Validators.required]);

        this.resurrectFieldWithSavedValue(AzureField.VNET_CUSTOM_CIDR, [Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.validationService.isValidIpNetworkSegment()
        ], this.defaultVnetCidr);
        this.resurrectFieldWithSavedValue(AzureField.VNET_CONTROLPLANE_NEWSUBNET_CIDR, [Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.validationService.isValidIpNetworkSegment()
        ], this.defaultControlPlaneCidr);

        if (!this.modeClusterStandalone) {
            this.resurrectFieldWithSavedValue(AzureField.VNET_WORKER_NEWSUBNET_NAME, [Validators.required]);
            this.resurrectFieldWithSavedValue(AzureField.VNET_WORKER_NEWSUBNET_CIDR, [Validators.required,
                this.validationService.noWhitespaceOnEnds(),
                this.validationService.isValidIpNetworkSegment()
            ], this.defaultWorkerNodeCidr);
        }

        this.vnetFieldsExisting.forEach(field => this.disarmField(field, clearSavedData));
    }

    private initVnetFieldsExisting(clearSavedData: boolean) {
        this.vnetFieldsExisting.forEach(field => {
            this.resurrectField(field, [Validators.required]);
        });
        if (this.vnetResourceGroups.length === 1) {
            setTimeout(_ => {
                if (this.vnetResourceGroups.length === 1) {
                    this.setControlValueSafely(AzureField.VNET_RESOURCE_GROUP, this.vnetResourceGroups[0].name);
                }
                if (this.vnetNamesExisting.length === 1) {
                    this.setControlValueSafely(AzureField.VNET_EXISTING_NAME, this.vnetNamesExisting[0]);
                }
                if (this.controlPlaneSubnets.length === 1) {
                    this.setControlValueSafely(AzureField.VNET_CONTROLPLANE_SUBNET_NAME, this.controlPlaneSubnets[0].name)
                }
                if (this.workerNodeSubnets.length === 1) {
                    this.setControlValueSafely(AzureField.VNET_WORKER_SUBNET_NAME, this.workerNodeSubnets[0].name)
                }
            });
        }
        AppServices.messenger.publish({
            type: TkgEventType.NETWORK_STEP_GET_NO_PROXY_INFO,
            payload: {info: '169.254.0.0/16,168.63.129.16'}
        });

        this.vnetFieldsNew.forEach(field => this.disarmField(field, clearSavedData));
    }

    // modifySavedValuesIfVnetCustomNameIsNowExisting() handles the case where user originally created a new (custom) vnet
    // (and value was either saved to local storage or to config file as a custom vnet name), but now when the data is restored,
    // the vnet name exists (so we should move the custom value over to the existing data slot).
    // In doing that, we create a side-effect of changing local storage values.
    private modifySavedValuesIfVnetCustomNameIsNowExisting() {
        const savedVnetCustom = this.getSavedValue(AzureField.VNET_CUSTOM_NAME, '');
        const customIsNowExisting = this.vnetNamesExisting.indexOf(savedVnetCustom) >= 0;
        if (customIsNowExisting) {
            this.saveFieldData(AzureField.VNET_EXISTING_NAME, savedVnetCustom);
            this.clearFieldSavedData(AzureField.VNET_CUSTOM_NAME);
        }
    }

    dynamicDescription(): string {
        const vnetCidrBlock = this.getFieldValue(AzureField.VNET_CUSTOM_CIDR, true);
        return vnetCidrBlock ? `Subnet: ${vnetCidrBlock}` : 'Specify an Azure VNET CIDR';
    }
}
