import { AzureVirtualNetwork } from './../../../../swagger/models/azure-virtual-network.model';
import { TkgEventType } from 'src/app/shared/service/Messenger';
import { ValidationService } from './../../wizard/shared/validation/validation.service';
/**
 * Angular Modules
 */
import { Component, OnInit } from '@angular/core';
import { FormControl, Validators } from '@angular/forms';

import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { distinctUntilChanged, takeUntil } from 'rxjs/operators';
import { AzureWizardFormService } from 'src/app/shared/service/azure-wizard-form.service';
import { AzureResourceGroup } from 'src/app/swagger/models';
import { APIClient } from 'src/app/swagger';
import { FormMetaDataStore } from '../../wizard/shared/FormMetaDataStore'
import Broker from 'src/app/shared/service/broker';
import { AzureForm } from '../azure-wizard.constants';
import { FormUtils } from '../../wizard/shared/utils/form-utils';

const CUSTOM = "CUSTOM";
export const EXISTING = "EXISTING";

enum VnetField {
    EXISTING_OR_CUSTOM = 'vnetOption',
    RESOURCE_GROUP = 'vnetResourceGroup',
    EXISTING_NAME = 'vnetNameExisting',
    CUSTOM_NAME = 'vnetNameCustom',
    CUSTOM_CIDR = 'vnetCidrBlock',
    PRIVATE_CLUSTER = 'privateAzureCluster',
    PRIVATE_IP = 'privateIP',
    // subnet fields:
    CONTROLPLANE_SUBNET_CIDR = 'controlPlaneSubnetCidr',
    CONTROLPLANE_SUBNET_NAME = 'controlPlaneSubnet',
    CONTROLPLANE_NEWSUBNET_CIDR = 'controlPlaneSubnetCidrNew',
    CONTROLPLANE_NEWSUBNET_NAME = 'controlPlaneSubnetNew',
    WORKER_SUBNET_NAME = 'workerNodeSubnet',
    WORKER_NEWSUBNET_CIDR = 'workerNodeSubnetCidrNew',
    WORKER_NEWSUBNET_NAME = 'workerNodeSubnetNew',
}

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
    defaultCidrFields: Array<string> = [];
    optionalFields: Array<string> = [];
    requiredFields: Array<string> = [];
    vnetFieldsExisting: Array<string> = [];
    vnetFieldsNew: Array<string> = [];

    constructor(private apiClient: APIClient,
        private validationService: ValidationService,
        private wizardFormService: AzureWizardFormService) {
        super();
    }

    /**
     * Create the initial form
     */
    private buildForm() {
        this.requiredFields.forEach(key => FormUtils.addControl(
            this.formGroup,
            key,
            new FormControl('', [
                Validators.required
            ])
        ));

        this.defaultCidrFields.forEach(field => FormUtils.addControl(
            this.formGroup,
            field,
            new FormControl('', [
                Validators.required,
                this.validationService.noWhitespaceOnEnds(),
                this.validationService.isValidIpNetworkSegment()
            ])
        ));
        // special hidden field used to capture existing subnet cidr when user selects existing subnet
        FormUtils.addControl(
            this.formGroup,
            VnetField.CONTROLPLANE_SUBNET_CIDR,
            new FormControl('', [])
        );

        this.optionalFields.forEach(field => FormUtils.addControl(
            this.formGroup,
            field,
            new FormControl('', [])
        ));

        this.formGroup.get(VnetField.RESOURCE_GROUP).valueChanges
            .pipe(
                distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
                takeUntil(this.unsubscribe)
            ).subscribe((val) => {
                this.onResourceGroupChange(val);
            });

        this.formGroup.get(VnetField.EXISTING_NAME).valueChanges
            .pipe(
                distinctUntilChanged((prev, curr) => prev === curr),
                takeUntil(this.unsubscribe)
            ).subscribe(newValue => {
                this.onVnetChange(newValue);
            });

        this.formGroup.get(VnetField.CUSTOM_CIDR).valueChanges
            .pipe(
                distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
                takeUntil(this.unsubscribe)
            ).subscribe((cidr) => {
                Broker.messenger.publish({
                    type: TkgEventType.NETWORK_STEP_GET_NO_PROXY_INFO,
                    payload: { info: (cidr ? cidr + ',' : '') + '169.254.0.0/16,168.63.129.16' }
                });
            });
    }

    /**
     * Initialize the form with data from the backend
     */
    private initForm() {
        const customResourceGroup = FormMetaDataStore.getMetaDataItem("providerForm", "resourceGroupCustom");
        this.customResourceGroup = customResourceGroup ? customResourceGroup.displayValue : '';
    }

    ngOnInit() {
        super.ngOnInit();

        this.optionalFields = [VnetField.PRIVATE_CLUSTER, VnetField.PRIVATE_IP];
        if (this.modeClusterStandalone) {
            this.requiredFields =
                [
                    VnetField.EXISTING_OR_CUSTOM,
                    VnetField.RESOURCE_GROUP,
                    VnetField.CUSTOM_NAME,
                    VnetField.EXISTING_NAME,
                    VnetField.CONTROLPLANE_SUBNET_NAME,
                    VnetField.CONTROLPLANE_NEWSUBNET_NAME,
                    VnetField.CONTROLPLANE_NEWSUBNET_CIDR,
                ];
            this.defaultCidrFields = [VnetField.CUSTOM_CIDR, VnetField.CONTROLPLANE_NEWSUBNET_CIDR, ];
            this.vnetFieldsExisting = [VnetField.EXISTING_NAME, VnetField.CONTROLPLANE_SUBNET_NAME];
            this.vnetFieldsNew =
                [
                    VnetField.CUSTOM_NAME,
                    VnetField.CUSTOM_CIDR,
                    VnetField.CONTROLPLANE_NEWSUBNET_NAME,
                    VnetField.CONTROLPLANE_NEWSUBNET_CIDR,
                ];
        } else {
            this.requiredFields =
                [
                    VnetField.EXISTING_OR_CUSTOM,
                    VnetField.RESOURCE_GROUP,
                    VnetField.CUSTOM_NAME,
                    VnetField.EXISTING_NAME,
                    VnetField.CONTROLPLANE_SUBNET_NAME,
                    VnetField.WORKER_SUBNET_NAME,
                    VnetField.CONTROLPLANE_NEWSUBNET_NAME,
                    VnetField.CONTROLPLANE_NEWSUBNET_CIDR,
                    VnetField.WORKER_NEWSUBNET_NAME,
                    VnetField.WORKER_NEWSUBNET_CIDR
                ];
            this.defaultCidrFields = [VnetField.CUSTOM_CIDR, VnetField.CONTROLPLANE_NEWSUBNET_CIDR, VnetField.WORKER_NEWSUBNET_CIDR];
            this.vnetFieldsExisting = [VnetField.EXISTING_NAME, VnetField.CONTROLPLANE_SUBNET_NAME, VnetField.WORKER_SUBNET_NAME];
            this.vnetFieldsNew =
                [
                    VnetField.CUSTOM_NAME,
                    VnetField.CUSTOM_CIDR,
                    VnetField.CONTROLPLANE_NEWSUBNET_NAME,
                    VnetField.CONTROLPLANE_NEWSUBNET_CIDR,
                    VnetField.WORKER_NEWSUBNET_NAME,
                    VnetField.WORKER_NEWSUBNET_CIDR
                ];
        }
        this.buildForm();
        this.initForm();

        /**
         * Whenever Azure region selection changes...
         */
        Broker.messenger.getSubject(TkgEventType.AZURE_REGION_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                this.onRegionChange(event.payload);
            });

        Broker.messenger.getSubject(TkgEventType.AZURE_RESOURCEGROUP_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                this.customResourceGroup = event.payload;
            });

        this.wizardFormService.getErrorStream(TkgEventType.AZURE_GET_RESOURCE_GROUPS)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(error => {
                this.errorNotification = error;
            });

        this.wizardFormService.getDataStream(TkgEventType.AZURE_GET_RESOURCE_GROUPS)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((azureResourceGroups: AzureResourceGroup[]) => {
                this.vnetResourceGroups = azureResourceGroups;
                if (this.customResourceGroup) {
                    this.setControlValueSafely(VnetField.RESOURCE_GROUP, this.customResourceGroup);
                } else if (azureResourceGroups.length === 1) {
                    this.setControlValueSafely(VnetField.RESOURCE_GROUP, azureResourceGroups[0].name);
                } else {
                    this.setControlWithSavedValue(VnetField.RESOURCE_GROUP);
                }
            });

        this.registerOnValueChange(VnetField.PRIVATE_CLUSTER, this.onCreatePrivateAzureCluster.bind(this));
        this.registerOnValueChange(VnetField.CONTROLPLANE_NEWSUBNET_CIDR, this.onControlPlaneSubnetCidrNewChange.bind(this));
        this.registerOnValueChange(VnetField.CONTROLPLANE_SUBNET_NAME, this.onControlPlaneSubnetChange.bind(this));
        this.registerOnValueChange(VnetField.EXISTING_OR_CUSTOM, this.onExistingOrCustomOptionChange.bind(this));

        this.initFormWithSavedData();
    }

    onControlPlaneSubnetChange(name: string) {
        // when the user selects an existing subnet, we look up the associated CIDR and set a hidden field with the CIDR value,
        // which is later used in creating the AzureRegionalClusterParams payload object
        const subnetEntry = this.controlPlaneSubnets.find(subnet => subnet.name === name);
        const cidrOfSelectedControlPlaneSubnet = (subnetEntry) ? subnetEntry.cidr : '';

        this.setControlValueSafely(VnetField.CONTROLPLANE_SUBNET_CIDR, cidrOfSelectedControlPlaneSubnet);
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
            this.resurrectFieldWithSavedValue(VnetField.PRIVATE_IP,
                [Validators.required, this.validationService.isValidIpOrFqdn(),
                    cidrValidator]
            );
        } else {
            this.disarmField(VnetField.PRIVATE_IP, true);
        }
    }

    initFormWithSavedData() {
        this.setControlWithSavedValue(VnetField.PRIVATE_CLUSTER, false);

        // if the user did an import, then we expect there may be a custom vnet value to be stored in VnetField.CUSTOM_NAME slot.
        // however, since that custom vnet may have been created and now be an existing vnet,
        // we need to call a special method to handle that situation, which moves the saved name from the custom slot
        // into the existing slot. Note that we do this before deciding whether to show the custom or existing options
        this.modifySavedValuesIfVnetCustomNameIsNowExisting();
        const savedVnetCustom = this.getSavedValue(VnetField.CUSTOM_NAME, '');
        const optionValue = savedVnetCustom !== '' ? CUSTOM : EXISTING;
        // NOTE: setting the EXISTING_OR_CUSTOM value will trigger the display to update
        this.setControlWithSavedValue(VnetField.EXISTING_OR_CUSTOM, optionValue);
    }

    onRegionChange(region: string) {
        this.region = region;
        this.setControlWithSavedValue(VnetField.RESOURCE_GROUP);
    }

    onResourceGroupChange(resourceGroupName) {
        if (resourceGroupName && resourceGroupName !== this.customResourceGroup) {
            this.apiClient.getAzureVnets({ resourceGroupName, location: this.region })
                .pipe(takeUntil(this.unsubscribe))
                .subscribe(
                    (vnets: AzureVirtualNetwork[]) => { this.setVnets(vnets); },
                    err => { this.errorNotification = err.message; },
                    () => { }
                );
        }
    }

    private setVnets(vnets: AzureVirtualNetwork[]) {
        this.vnetSubnets = vnets.reduce((accu, vnet) => { accu[vnet.name] = vnet.subnets; return accu; }, {});
        this.vnetNamesExisting = Object.keys(this.vnetSubnets);

        this.modifySavedValuesIfVnetCustomNameIsNowExisting();

        // NOTE: setting the EXISTING_NAME value will cause an update to the subnets
        const savedVnet = this.getSavedValue(VnetField.EXISTING_NAME, '');
        if (savedVnet && this.vnetNamesExisting.includes(savedVnet)) {
            this.setControlValueSafely(VnetField.EXISTING_NAME, savedVnet);
        } else if (this.vnetNamesExisting.length === 1) {
            this.setControlValueSafely(VnetField.EXISTING_NAME, this.vnetNamesExisting[0]);
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
        this.initSubnetField(VnetField.CONTROLPLANE_SUBNET_NAME, this.controlPlaneSubnets);
        if (!this.modeClusterStandalone) {
            this.initSubnetField(VnetField.WORKER_SUBNET_NAME, this.workerNodeSubnets);
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
                this.setControlValueSafely(VnetField.RESOURCE_GROUP, this.vnetResourceGroups[0].name);
            } else {
                this.setControlWithSavedValue(VnetField.RESOURCE_GROUP);
            }
        } else if (option === CUSTOM) {
            this.showVnetFieldsCustom(clearSavedData);
        }
        this.onCreatePrivateAzureCluster(this.createPrivateCluster);
    }

    private showVnetFieldsCustom(clearSavedData: boolean) {
        this.resurrectFieldWithSavedValue(VnetField.CUSTOM_NAME, [Validators.required]);
        this.resurrectFieldWithSavedValue(VnetField.CONTROLPLANE_NEWSUBNET_NAME, [Validators.required]);

        this.resurrectFieldWithSavedValue(VnetField.CUSTOM_CIDR, [Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.validationService.isValidIpNetworkSegment()
        ], this.defaultVnetCidr);
        this.resurrectFieldWithSavedValue(VnetField.CONTROLPLANE_NEWSUBNET_CIDR, [Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.validationService.isValidIpNetworkSegment()
        ], this.defaultControlPlaneCidr);

        if (!this.modeClusterStandalone) {
            this.resurrectFieldWithSavedValue(VnetField.WORKER_NEWSUBNET_NAME, [Validators.required]);
            this.resurrectFieldWithSavedValue(VnetField.WORKER_NEWSUBNET_CIDR, [Validators.required,
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
                    this.setControlValueSafely(VnetField.RESOURCE_GROUP, this.vnetResourceGroups[0].name);
                }
                if (this.vnetNamesExisting.length === 1) {
                    this.setControlValueSafely(VnetField.EXISTING_NAME, this.vnetNamesExisting[0]);
                }
                if (this.controlPlaneSubnets.length === 1) {
                    this.setControlValueSafely(VnetField.CONTROLPLANE_SUBNET_NAME, this.controlPlaneSubnets[0].name)
                }
                if (this.workerNodeSubnets.length === 1) {
                    this.setControlValueSafely(VnetField.WORKER_SUBNET_NAME, this.workerNodeSubnets[0].name)
                }
            });
        }
        Broker.messenger.publish({
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
        const savedVnetCustom = this.getSavedValue(VnetField.CUSTOM_NAME, '');
        const customIsNowExisting = this.vnetNamesExisting.indexOf(savedVnetCustom) >= 0;
        if (customIsNowExisting) {
            this.saveFieldData(VnetField.EXISTING_NAME, savedVnetCustom);
            this.clearFieldSavedData(VnetField.CUSTOM_NAME);
        }
    }

    protected dynamicDescription(): string {
        const vnetCidrBlock = this.getFieldValue("vnetCidrBlock", true);
        return vnetCidrBlock ? `Subnet: ${vnetCidrBlock}` : 'Specify a Azure VNET CIDR';
    }
}
