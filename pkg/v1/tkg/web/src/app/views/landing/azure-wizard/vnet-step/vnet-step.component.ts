import { AzureVirtualNetwork } from './../../../../swagger/models/azure-virtual-network.model';
import { TkgEventType } from 'src/app/shared/service/Messenger';
import { ValidationService } from './../../wizard/shared/validation/validation.service';
/**
 * Angular Modules
 */
import { Component, OnInit } from '@angular/core';
import {
    Validators,
    FormControl
} from '@angular/forms';

import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { takeUntil, distinctUntilChanged } from 'rxjs/operators';
import { AzureWizardFormService } from 'src/app/shared/service/azure-wizard-form.service';
import { AzureResourceGroup } from 'src/app/swagger/models';
import { APIClient } from 'src/app/swagger';
import { FormMetaDataStore } from '../../wizard/shared/FormMetaDataStore'
import Broker from 'src/app/shared/service/broker';

const CUSTOM = "CUSTOM";
export const EXISTING = "EXISTING";
@Component({
    selector: 'app-vnet-step',
    templateUrl: './vnet-step.component.html',
    styleUrls: ['./vnet-step.component.scss']
})
export class VnetStepComponent extends StepFormDirective implements OnInit {

    region = "";    // Current region selected
    showOption = CUSTOM;

    // An object maps vnet to subsets
    vnetSubnets = {};

    /** lists to be retrieved from the backend */
    resourceGroups = [];
    customResourceGroup = null;
    vnetNamesExisting = [];
    controlPlaneSubnets = [];
    workerNodeSubnets = [];

    defaultVnetCidr: string = '10.0.0.0/16';
    defaultControlPlaneCidr: string = '10.0.0.0/24';
    defaultWorkerNodeCidr: string = '10.0.1.0/24';

    cidrHolder = {}

    createPrivateCluster = false;

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
        this.requiredFields.forEach(key => this.formGroup.addControl(
            key,
            new FormControl('', [
                Validators.required
            ])
        ));

        this.defaultCidrFields.forEach(field => this.formGroup.addControl(
            field,
            new FormControl('', [
                Validators.required,
                this.validationService.noWhitespaceOnEnds(),
                this.validationService.isValidIpNetworkSegment()
            ])
        ));
        // special hidden field used to capture existing subnet cidr when user selects existing subnet
        this.formGroup.addControl(
            'controlPlaneSubnetCidr',
            new FormControl('', [])
        );

        this.formGroup.get('vnetOption').setValue(this.showOption);

        this.optionalFields.forEach(field => this.formGroup.addControl(
            field,
            new FormControl('', [])
        ));

        this.formGroup.get('resourceGroup').valueChanges
            .pipe(
                distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
                takeUntil(this.unsubscribe)
            ).subscribe((val) => {
                this.onResourceGroupChange(val);
            });

        this.formGroup.get('vnetNameExisting').valueChanges
            .pipe(
                takeUntil(this.unsubscribe)
            ).subscribe((val) => {
                this.onVnetChange(val)
            });

        this.formGroup.get('vnetCidrBlock').valueChanges
            .pipe(
                distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
                takeUntil(this.unsubscribe)
            ).subscribe((cidr) => {
                Broker.messenger.publish({
                    type: TkgEventType.AWS_GET_NO_PROXY_INFO,
                    payload: { info: (cidr ? cidr + ',' : '') + '169.254.0.0/16,168.63.129.16' }
                });
            });
    }

    /**
     * Initialize the form with data from the backend
     */
    private initForm() {
        const customResourceGroup = FormMetaDataStore.getMetaDataItem("providerForm", "resourceGroupCustom");
        this.customResourceGroup = customResourceGroup ? customResourceGroup.displayValue : ''
    }

    ngOnInit() {
        super.ngOnInit();

        this.requiredFields = this.modeClusterStandalone ?
            [
                "vnetOption",
                "resourceGroup",
                "vnetNameCustom",
                "vnetNameExisting",
                "controlPlaneSubnet",
                "controlPlaneSubnetNew",
                "controlPlaneSubnetCidrNew",
            ] :
            [
                "vnetOption",
                "resourceGroup",
                "vnetNameCustom",
                "vnetNameExisting",
                "controlPlaneSubnet",
                "workerNodeSubnet",
                "controlPlaneSubnetNew",
                "controlPlaneSubnetCidrNew",
                "workerNodeSubnetNew",
                "workerNodeSubnetCidrNew"
            ];

        this.optionalFields = ['privateAzureCluster', 'privateIP'];

        this.defaultCidrFields = this.modeClusterStandalone ?
            ["vnetCidrBlock", "controlPlaneSubnetCidrNew", ] :
            ["vnetCidrBlock", "controlPlaneSubnetCidrNew", "workerNodeSubnetCidrNew"];

        this.vnetFieldsExisting = this.modeClusterStandalone ?
            ["vnetNameExisting", "controlPlaneSubnet"] :
            ["vnetNameExisting", "controlPlaneSubnet", "workerNodeSubnet"];

        this.vnetFieldsNew = this.modeClusterStandalone ?
            [
                "vnetNameCustom",
                "vnetCidrBlock",
                "controlPlaneSubnetNew",
                "controlPlaneSubnetCidrNew",
            ] :
            [
                "vnetNameCustom",
                "vnetCidrBlock",
                "controlPlaneSubnetNew",
                "controlPlaneSubnetCidrNew",
                "workerNodeSubnetNew",
                "workerNodeSubnetCidrNew"
            ];

        this.buildForm();
        this.initForm();
        this.show(this.showOption, false);

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
            .subscribe((rgs: AzureResourceGroup[]) => {
                this.resourceGroups = rgs;
                if (this.customResourceGroup) {
                    this.formGroup.get('resourceGroup').setValue(this.customResourceGroup);
                } else if (rgs.length === 1) {
                    this.formGroup.get('resourceGroup').setValue(rgs[0].name);
                }
            });

        this.registerOnValueChange('privateAzureCluster', this.onCreatePrivateAzureCluster.bind(this));
        this.registerOnValueChange('controlPlaneSubnetCidrNew', this.onControlPlaneSubnetCidrNewChange.bind(this));
        this.registerOnValueChange('controlPlaneSubnet', this.onControlPlaneSubnetChange.bind(this));
        this.registerOnValueChange('vnetOption', this.show.bind(this));
    }

    onControlPlaneSubnetChange(name: string) {
        // when the user selects an existing subnet, we look up the associated CIDR and set a hidden field with the CIDR value,
        // which is later used in creating the AzureRegionalClusterParams payload object
        const subnetEntry = this.controlPlaneSubnets.find(subnet => subnet.name === name);
        const cidrOfSelectedControlPlaneSubnet = (subnetEntry) ? subnetEntry.cidr : '';

        const control = this.formGroup.controls['controlPlaneSubnetCidr'];
        if (control) {
            control.setValue(cidrOfSelectedControlPlaneSubnet);
        }
        // Leaving cidrHolder assignment in place, but unable to see how it is useful
        this.cidrHolder[EXISTING] = cidrOfSelectedControlPlaneSubnet;
    }

    onControlPlaneSubnetCidrNewChange(value: string) {
        this.cidrHolder[CUSTOM] = value;
    }

    onCreatePrivateAzureCluster(createPrivateCluster: boolean) {
        if (createPrivateCluster) {      // private azure cluster
            this.createPrivateCluster = true;
            const cidrValidator = this.validationService.isIpInSubnet2(this.cidrHolder, "" + this.showOption);

            this.resurrectField('privateIP', [Validators.required, this.validationService.isValidIpOrFqdn(), cidrValidator])
        } else {
            this.createPrivateCluster = false;
            this.disarmField('privateIP', true);
        }
    }

    setSavedDataAfterLoad() {
        if (!this.hasSavedData() || this.getSavedValue('vnetNameExisting', '') === '') {
            this.show(CUSTOM, true);
        } else {
            this.show(EXISTING, true)
        }
        super.setSavedDataAfterLoad();
    }

    onRegionChange(region: string) {
        this.region = region;
        this.formGroup.get("resourceGroup").setValue('');
    }

    onResourceGroupChange(resourceGroupName) {
        if (resourceGroupName && resourceGroupName !== this.customResourceGroup) {
            this.apiClient.getAzureVnets({ resourceGroupName, location: this.region })
                .pipe(takeUntil(this.unsubscribe))
                .subscribe(
                    (vnets: AzureVirtualNetwork[]) => {
                        this.vnetSubnets = vnets.reduce((accu, vnet) => { accu[vnet.name] = vnet.subnets; return accu; }, {});
                        this.vnetNamesExisting = Object.keys(this.vnetSubnets);
                        this.formGroup.get('vnetNameExisting').setValue(this.vnetNamesExisting.length === 1 ?
                            this.vnetNamesExisting[0] : this.getSavedValue('vnetNameExisting', ''))
                    },
                    err => { this.errorNotification = err.message; },
                    () => { }
                );
        }

        this.vnetFieldsExisting.forEach(field => this.formGroup.get(field).setValue(''));
    }

    onVnetChange(vnetName) {
        // Use the same source
        this.workerNodeSubnets = this.vnetSubnets[vnetName] || [];
        this.controlPlaneSubnets = this.workerNodeSubnets;

        // set child fields with local storage data if available
        let filteredSubnets = this.controlPlaneSubnets.filter(s => s.name === this.getSavedValue('controlPlaneSubnet', ''));
        if (filteredSubnets.length > 0) {
            this.formGroup.get('controlPlaneSubnet').setValue(filteredSubnets[0].cidr)
        } else if (this.controlPlaneSubnets.length === 1) {
            this.formGroup.get('controlPlaneSubnet').setValue(this.controlPlaneSubnets[0].name)
        }
        if (!this.modeClusterStandalone) {
            filteredSubnets = this.workerNodeSubnets.filter(s => s.name === this.getSavedValue('workerNodeSubnet', ''));
            if (filteredSubnets.length > 0) {
                this.formGroup.get('workerNodeSubnet').setValue(filteredSubnets[0].cidr);
            } else if (this.workerNodeSubnets.length === 1) {
                this.formGroup.get('workerNodeSubnet').setValue(this.workerNodeSubnets[0].name)
            }
        }
    }

    show(option: string, clearSavedData = true) {
        this.showOption = option;

        if (option === EXISTING) {
            this.vnetFieldsExisting.forEach(field => this.resurrectField(field, [Validators.required]));
            this.vnetFieldsNew.forEach(field => this.disarmField(field, clearSavedData));
            this.formGroup.get('vnetOption').setValue(EXISTING);
            if ( this.resourceGroups.length === 1 ) {
                setTimeout( _ => {
                    if (this.resourceGroups.length === 1) {
                        this.formGroup.get('resourceGroup').setValue(this.resourceGroups[0].name);
                    }
                    if (this.vnetNamesExisting.length === 1) {
                        this.formGroup.get('vnetNameExisting').setValue(this.vnetNamesExisting[0]);
                    }
                    if (this.controlPlaneSubnets.length === 1) {
                        this.formGroup.get('controlPlaneSubnet').setValue(this.controlPlaneSubnets[0].name)
                    }
                    if (this.workerNodeSubnets.length === 1) {
                        this.formGroup.get('workerNodeSubnet').setValue(this.workerNodeSubnets[0].name)
                    }
                });
            }
            Broker.messenger.publish({
                type: TkgEventType.AWS_GET_NO_PROXY_INFO,
                payload: { info: '169.254.0.0/16,168.63.129.16' }
            });
        } else if (option === CUSTOM) {
            this.resurrectField("vnetNameCustom", [Validators.required]);
            this.resurrectField("controlPlaneSubnetNew", [Validators.required]);

            this.resurrectField("vnetCidrBlock", [Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.validationService.isValidIpNetworkSegment()
            ], this.defaultVnetCidr);
            this.resurrectField("controlPlaneSubnetCidrNew", [Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.validationService.isValidIpNetworkSegment()
            ], this.defaultControlPlaneCidr);

            if (!this.modeClusterStandalone) {
                this.resurrectField("workerNodeSubnetNew", [Validators.required]);
                this.resurrectField("workerNodeSubnetCidrNew", [Validators.required,
                    this.validationService.noWhitespaceOnEnds(),
                    this.validationService.isValidIpNetworkSegment()
                ], this.defaultWorkerNodeCidr);
            }

            this.vnetFieldsExisting.forEach(field => this.disarmField(field, clearSavedData));
            this.formGroup.get('vnetOption').setValue(CUSTOM);
        }

        this.onCreatePrivateAzureCluster(this.createPrivateCluster);
    }
}
