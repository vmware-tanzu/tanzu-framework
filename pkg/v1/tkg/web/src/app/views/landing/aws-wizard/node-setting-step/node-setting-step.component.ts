// Angular imports
import { Component, OnInit } from '@angular/core';
import { Validators } from '@angular/forms';
import { takeUntil } from 'rxjs/operators';
// App imports
import { APIClient } from '../../../../swagger/api-client.service';
import { AppEdition } from 'src/app/shared/constants/branding.constants';
import AppServices from '../../../../shared/service/appServices';
import { AwsField, VpcType } from "../aws-wizard.constants";
import { AwsNodeSettingStepMapping } from './node-setting-step.fieldmapping';
import { AWSNodeAz } from '../../../../swagger/models/aws-node-az.model';
import { AWSSubnet } from '../../../../swagger/models/aws-subnet.model';
import { AzRelatedFieldsArray } from '../aws-wizard.component';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { TanzuEventType } from '../../../../shared/service/Messenger';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { NodeSettingStepDirective } from '../../wizard/shared/components/steps/node-setting-step/node-setting-step.component';
import { NodeSettingField } from '../../wizard/shared/components/steps/node-setting-step/node-setting-step.fieldmapping';

export interface AzNodeTypes {
    awsNodeAz1: Array<string>,
    awsNodeAz2: Array<string>,
    awsNodeAz3: Array<string>
}

export interface FilteredAzs {
    awsNodeAz1: {
        publicSubnets: Array<AWSSubnet>;
        privateSubnets: Array<AWSSubnet>;
    },
    awsNodeAz2: {
        publicSubnets: Array<AWSSubnet>;
        privateSubnets: Array<AWSSubnet>;
    },
    awsNodeAz3: {
        publicSubnets: Array<AWSSubnet>;
        privateSubnets: Array<AWSSubnet>;
    }
}

const AZS = [
    AwsField.NODESETTING_AZ_1,
    AwsField.NODESETTING_AZ_2,
    AwsField.NODESETTING_AZ_3,
];
const WORKER_NODE_INSTANCE_TYPES = [
    NodeSettingField.WORKER_NODE_INSTANCE_TYPE,
    AwsField.NODESETTING_WORKERTYPE_2,
    AwsField.NODESETTING_WORKERTYPE_3
];
const PUBLIC_SUBNETS = [
    AwsField.NODESETTING_VPC_PUBLIC_SUBNET_1,
    AwsField.NODESETTING_VPC_PUBLIC_SUBNET_2,
    AwsField.NODESETTING_VPC_PUBLIC_SUBNET_3
];
const PRIVATE_SUBNET = [
    AwsField.NODESETTING_VPC_PRIVATE_SUBNET_1,
    AwsField.NODESETTING_VPC_PRIVATE_SUBNET_2,
    AwsField.NODESETTING_VPC_PRIVATE_SUBNET_3,
];
const VPC_SUBNETS = [...PUBLIC_SUBNETS, ...PRIVATE_SUBNET];

enum vpcType {
    EXISTING = 'existing'
}

@Component({
    selector: 'app-node-setting-step',
    templateUrl: './node-setting-step.component.html',
    styleUrls: ['./node-setting-step.component.scss']
})

export class NodeSettingStepComponent extends NodeSettingStepDirective<string> implements OnInit {
    APP_EDITION: any = AppEdition;

    vpcType: string;
    nodeAzs: Array<AWSNodeAz> = [];
    azNodeTypes: AzNodeTypes = {
        awsNodeAz1: [],
        awsNodeAz2: [],
        awsNodeAz3: []
    };

    publicSubnets: Array<AWSSubnet> = new Array<AWSSubnet>();
    privateSubnets: Array<AWSSubnet> = new Array<AWSSubnet>();

    filteredAzs: FilteredAzs = {
        awsNodeAz1: {
            publicSubnets: [],
            privateSubnets: []
        },
        awsNodeAz2: {
            publicSubnets: [],
            privateSubnets: []
        },
        awsNodeAz3: {
            publicSubnets: [],
            privateSubnets: []
        }
    };

    config = {
        displayKey: 'description',
        search: true,
        height: 'auto',
        placeholder: 'Select',
        customComparator: () => { },
        moreText: 'more',
        noResultsFound: 'No results found!',
        searchPlaceholder: 'Search',
        searchOnKey: 'name',
        clearOnSelection: false,
        inputDirection: 'ltr'
    };

    airgappedVPC = false;

    constructor(protected validationService: ValidationService,
                private apiClient: APIClient) {
        super(validationService);
    }

    protected createStepMapping(): StepMapping {
        // AWS has a bunch of extra fields (related to AZs) over and above the default base class field mapping
        const result = { fieldMappings: [ ...super.createStepMapping().fieldMappings, ...AwsNodeSettingStepMapping.fieldMappings]};
        // The worker node instance in the base class is used as workerNodeInstance1 for AWS; it needs diff label and requiresBackendData
        const workerNodeInstanceType = AppServices.fieldMapUtilities.getFieldMapping(NodeSettingField.WORKER_NODE_INSTANCE_TYPE, result);
        workerNodeInstanceType.label = 'AZ1 WORKER NODE INSTANCE TYPE';
        workerNodeInstanceType.requiresBackendData = true;

        return result;
    }

    protected onProdInstanceTypeChange(prodNodeType: string) {
        // AWS sets the worker instance in response to the AZ selection, not the prod instance type selection
    }

    protected onDevInstanceTypeChange(devNodeType: string) {
        // AWS sets the worker instance in response to the AZ selection, not the dev instance type selection
    }

    protected listenToEvents() {
        super.listenToEvents();
        AppServices.messenger.subscribe<boolean>(TanzuEventType.AWS_AIRGAPPED_VPC_CHANGE, event => {
            this.airgappedVPC = event.payload;
            if (this.airgappedVPC) { // public subnet IDs shouldn't be provided
                PUBLIC_SUBNETS.forEach(f => {
                    const control = this.getControl(f);
                    control.setValue('');
                    control.disable();
                })
            } else {        // public subnet IDs are required
                PUBLIC_SUBNETS.forEach(f => {
                    this.getControl(f).enable();
                })
            }
        });

        AppServices.messenger.subscribe(TanzuEventType.AWS_REGION_CHANGED, () => {
            if (this.formGroup.get(AwsField.NODESETTING_AZ_1)) {
                this.publicSubnets = [];
                this.privateSubnets = [];

                this.clearSubnetData();
                this.clearAzs();
                this.clearSubnets();
            }
        }, this.unsubscribe);

        AppServices.messenger.subscribe<{ vpcType: string }>(TanzuEventType.AWS_VPC_TYPE_CHANGED, event => {
            this.vpcType = event.payload.vpcType;
            if (this.vpcType !== vpcType.EXISTING) {
                this.clearSubnets();
            }
            this.updateVpcSubnets();

            // clear az selection
            this.clearAzs();
            [...AZS, ...WORKER_NODE_INSTANCE_TYPES, ...VPC_SUBNETS].forEach(
                field => this.getControl(field).updateValueAndValidity()
            );
        });

        AppServices.messenger.subscribe(TanzuEventType.AWS_VPC_CHANGED, () => {
            this.clearAzs();
            this.clearSubnets();
        });

        AzRelatedFieldsArray.forEach(azRelatedFields => {
            this.registerOnValueChange(azRelatedFields.az, (newlySelectedAz) => {
                this.filterSubnetsByAZ(azRelatedFields.az, newlySelectedAz);
                this.setSubnetFieldsWithOnlyOneOption(azRelatedFields.az);
                this.updateWorkerNodeInstanceTypes(azRelatedFields.az, newlySelectedAz, azRelatedFields.workerNodeInstanceType);
            });
        });
    }

    protected subscribeToServices() {
        AppServices.dataServiceRegistrar.stepSubscribe<AWSSubnet>(this, TanzuEventType.AWS_GET_SUBNETS, this.onFetchedSubnets.bind(this));
        AppServices.dataServiceRegistrar.stepSubscribe<string>(this, TanzuEventType.AWS_GET_NODE_TYPES, this.onFetchedNodeTypes.bind(this));
        AppServices.dataServiceRegistrar.stepSubscribe<AWSNodeAz>(this, TanzuEventType.AWS_GET_AVAILABILITY_ZONES,
            this.onFetchedAzs.bind(this));
    }

    private onFetchedAzs(availabilityZones: Array<AWSNodeAz>) {
        this.nodeAzs = availabilityZones;
    }

    private onFetchedSubnets(subnets: Array<AWSSubnet>) {
        this.publicSubnets = subnets.filter(obj => { return obj.isPublic });
        this.privateSubnets = subnets.filter(obj => { return !obj.isPublic });
        AZS.forEach(az => { this.filterSubnetsByAZ(az, this.getFieldValue(az)); });
        this.setSubnetFieldsFromSavedValues();
    }

    private onFetchedNodeTypes(nodeTypes: Array<string>) {
        this.nodeTypes = nodeTypes.sort();

        // The validation is based on the value of this.nodeTypes. Whenever we update this.nodeTypes,
        // the corresponding validation should be updated as well. e.g. the users came to the node-settings
        // step before the api responses. Then an empty array will be passed to the validation isValidNameInList.
        // It will cause the selected option to be invalid all the time.
        if (this.isClusterPlanDev) {
            const devInstanceType = this.nodeTypes.length === 1 ? this.nodeTypes[0] :
                this.formGroup.get(AwsField.NODESETTING_INSTANCE_TYPE_DEV).value;
            this.resurrectField(AwsField.NODESETTING_INSTANCE_TYPE_DEV,
            [Validators.required, this.validationService.isValidNameInList(this.nodeTypes)],
            devInstanceType);
        } else {
            const prodInstanceType = this.nodeTypes.length === 1 ? this.nodeTypes[0] :
                this.formGroup.get(AwsField.NODESETTING_INSTANCE_TYPE_PROD).value;
            this.resurrectField(AwsField.NODESETTING_INSTANCE_TYPE_PROD,
                [Validators.required, this.validationService.isValidNameInList(this.nodeTypes)],
                prodInstanceType);
        }
    }

    protected setControlPlaneToProd() {
        super.setControlPlaneToProd();
        this.updateVpcSubnets();

        for (let i = 0; i < AZS.length; i++) {
            const thisAZ = AZS[i];
            const otherAZs = this.otherAZs(thisAZ);
            const thisAZcontrol = this.getControl(thisAZ);
            thisAZcontrol.setValidators([
                Validators.required,
                this.validationService.isUniqueAz([
                    this.getControl(otherAZs[0]),
                    this.getControl(otherAZs[1]) ])
            ]);
            this.setFieldWithStoredValue(thisAZ, this.supplyStepMapping());
        }
        if (!this.modeClusterStandalone) {
            WORKER_NODE_INSTANCE_TYPES.forEach((field, index) => {
                // only populated the worker node instance type if the associated AZ has a value
                if (this.getFieldValue(AZS[index])) {
                    this.resurrectFieldWithStoredValue(field.toString(), this.supplyStepMapping(), [Validators.required]);
                } else {
                    this.resurrectField(field.toString(), [Validators.required]);
                }
            });
        }
    }

    protected setControlPlaneToDev() {
        super.setControlPlaneToDev();
        this.updateVpcSubnets();
        const prodFields = [
            AwsField.NODESETTING_AZ_2,
            AwsField.NODESETTING_AZ_3,
            AwsField.NODESETTING_WORKERTYPE_2,
            AwsField.NODESETTING_WORKERTYPE_3,
            AwsField.NODESETTING_INSTANCE_TYPE_PROD
        ];
        prodFields.forEach(attr => this.disarmField(attr.toString(), true));
        if (this.nodeAzs && this.nodeAzs.length === 1) {
            this.setControlValueSafely(AwsField.NODESETTING_AZ_1, this.nodeAzs[0].name);
        } else {
            this.setFieldWithStoredValue(AwsField.NODESETTING_AZ_1, this.supplyStepMapping());
        }
    }

    // returns an array of the other two AZs (used to populate a validator that ensures unique AZs are selected)
    private otherAZs(targetAz: AwsField): AwsField[] {
        return AZS.filter((field, index, arr) => { return field !== targetAz });
    }

    get workerNodeInstanceType1Value() {
        return this.getFieldValue(NodeSettingField.WORKER_NODE_INSTANCE_TYPE);
    }

    get workerNodeInstanceType2Value() {
        return this.getFieldValue(AwsField.NODESETTING_WORKERTYPE_2);
    }

    get workerNodeInstanceType3Value() {
        return this.getFieldValue(AwsField.NODESETTING_WORKERTYPE_3);
    }

    // public for testing
    clearAzs() {
        AZS.forEach(az => this.clearControlValue(az));
    }

    // public for testing
    clearSubnets() {
        VPC_SUBNETS.forEach(vpcSubnet => this.clearControlValue(vpcSubnet));
    }

    // public for testing
    clearSubnetData() {
        AZS.forEach(az => {
            this.filteredAzs[az.toString()].publicSubnets = [];
            this.filteredAzs[az.toString()].privateSubnets = [];
        });
    }

    filterSubnetsByAZ(azControlName, az): void {
        if (this.vpcType === vpcType.EXISTING && azControlName !== '' && az !== '') {
            this.filteredAzs[azControlName].publicSubnets = this.filterSubnetArrayByAZ(az, this.publicSubnets);
            this.filteredAzs[azControlName].privateSubnets = this.filterSubnetArrayByAZ(az, this.privateSubnets);
        }
    }

    private filterSubnetArrayByAZ(az: string, subnets: AWSSubnet[]): AWSSubnet[] {
        return (!subnets) ? [] : subnets.filter(subnet => { return subnet.availabilityZoneName === az; });
    }

    private setSubnetFieldsWithOnlyOneOption(azControlName) {
        if (this.vpcType === vpcType.EXISTING && azControlName !== '') {
            const filteredPublicSubnets = this.filteredAzs[azControlName].publicSubnets;
            if (filteredPublicSubnets.length === 1) {
                this.setControlValueSafely(this.getPublicSubnetFromAz(azControlName), filteredPublicSubnets[0].id);
            }
            const filteredPrivateSubnets = this.filteredAzs[azControlName].privateSubnets;
            if (filteredPrivateSubnets.length === 1) {
                this.setControlValueSafely(this.getPrivateSubnetFromAz(azControlName), filteredPrivateSubnets[0].id);
            }
        }
    }

    private getPublicSubnetFromAz(azControlName: AwsField): AwsField {
        const indexAZ = AZS.indexOf(azControlName);
        if (indexAZ < 0) {
            console.log('WARNING: getPrivateSubnetFieldNameFromAzName() received unrecognized azControlName of ' + azControlName);
            return null;
        }
        return PUBLIC_SUBNETS[indexAZ];
    }

    private getPrivateSubnetFromAz(azControlName: AwsField): AwsField {
        const indexAZ = AZS.indexOf(azControlName);
        if (indexAZ < 0) {
            console.log('WARNING: getPrivateSubnetFieldNameFromAzName() received unrecognized azControlName of ' + azControlName);
            return null;
        }
        return PRIVATE_SUBNET[indexAZ];
    }

    // updateWorkerNodeInstanceTypes() is called when the user has selected a new value (newlySelectedAz) for an azField.
    // We need to get the worker node types available on that AZ and use them to populate our data structure that holds them.
    // If there is only one worker node type, then we want to set the value of the workerNodeField to that type (rather than
    // make the user "select it" from a list of only one element
    private updateWorkerNodeInstanceTypes(azField: string, newlySelectedAz: string, workerNodeField: string) {
        if (newlySelectedAz) {
            this.apiClient.getAWSNodeTypes({
                az: newlySelectedAz
            })
                .pipe(takeUntil(this.unsubscribe))
                .subscribe(
                    ((nodeTypes) => {
                        this.azNodeTypes[azField] = nodeTypes?.sort();
                        if (nodeTypes.length === 1) {
                            this.setControlValueSafely(workerNodeField, nodeTypes[0]);
                        } else {
                            // we default to the same instance type as the management cluster
                            const mgmtClusterInstanceType = this.isClusterPlanProd ? this.prodInstanceTypeValue : this.devInstanceTypeValue;
                            // ...but the stored value has precedence
                            const instanceType = this.getStoredValue(workerNodeField, this.supplyStepMapping(), mgmtClusterInstanceType);
                            this.setControlValueSafely(workerNodeField, instanceType);
                        }
                    }),
                    ((err) => {
                        const error = err.error.message || err.message || JSON.stringify(err);
                        this.errorNotification = `Unable to retrieve worker node instance types. ${error}`;
                    })
                );
        } else {
            this.azNodeTypes[newlySelectedAz] = [];
        }
    }

    setSubnetFieldsFromSavedValues(): void {
        PUBLIC_SUBNETS.forEach(vpcSubnet => {
            const subnet = this.findSubnetFromSavedValue(vpcSubnet, this['publicSubnets']);
            this.setControlValueSafely(vpcSubnet, subnet ? subnet.id : '');
        });
        PRIVATE_SUBNET.forEach(vpcSubnet => {
            const subnet = this.findSubnetFromSavedValue(vpcSubnet, this['privateSubnets']);
            this.setControlValueSafely(vpcSubnet, subnet ? subnet.id : '');
        });
    }

    // Given an array of subnet objects, find the one corresponding to the saved value of the given field
    private findSubnetFromSavedValue(subnetField: AwsField, subnets: AWSSubnet[]) {
        const savedValue = this.getStoredValue(subnetField, this.supplyStepMapping());
        // note that the saved value could either be the CIDR or the ID, so we find a match for either
        return subnets.find(x => { return x.cidr === savedValue || x.id === savedValue; });
    }

    updateVpcSubnets() {
        if (this.vpcType !== vpcType.EXISTING) {   // validations should be disabled for all public/private subnets
            [
                AwsField.NODESETTING_VPC_PRIVATE_SUBNET_1,
                AwsField.NODESETTING_VPC_PRIVATE_SUBNET_2,
                AwsField.NODESETTING_VPC_PRIVATE_SUBNET_3,
                AwsField.NODESETTING_VPC_PUBLIC_SUBNET_1,
                AwsField.NODESETTING_VPC_PUBLIC_SUBNET_2,
                AwsField.NODESETTING_VPC_PUBLIC_SUBNET_3
            ].forEach(field => {
                this.disarmField(field.toString(), false);
            });
            return;
        }

        if (this.isClusterPlanProd) {
            // in PROD deployments, all three subnets are used
            [
                AwsField.NODESETTING_VPC_PRIVATE_SUBNET_1,
                AwsField.NODESETTING_VPC_PRIVATE_SUBNET_2,
                AwsField.NODESETTING_VPC_PRIVATE_SUBNET_3,
                AwsField.NODESETTING_VPC_PUBLIC_SUBNET_1,
                AwsField.NODESETTING_VPC_PUBLIC_SUBNET_2,
                AwsField.NODESETTING_VPC_PUBLIC_SUBNET_3
            ].forEach(field => {
                this.resurrectFieldWithStoredValue(field.toString(), this.supplyStepMapping(), [Validators.required]);
            });
        } else if (this.isClusterPlanDev) {
            // in DEV deployments, only one subnet is used
            [
                AwsField.NODESETTING_VPC_PRIVATE_SUBNET_1,
                AwsField.NODESETTING_VPC_PUBLIC_SUBNET_1,
            ].forEach(field => {
                this.resurrectFieldWithStoredValue(field.toString(), this.supplyStepMapping(), [Validators.required]);
            });
            [
                AwsField.NODESETTING_VPC_PRIVATE_SUBNET_2,
                AwsField.NODESETTING_VPC_PRIVATE_SUBNET_3,
                AwsField.NODESETTING_VPC_PUBLIC_SUBNET_2,
                AwsField.NODESETTING_VPC_PUBLIC_SUBNET_3
            ].forEach(field => {
                this.disarmField(field.toString(), false);
            });
        }

        if (this.airgappedVPC) {
            [
                AwsField.NODESETTING_VPC_PUBLIC_SUBNET_1,
                AwsField.NODESETTING_VPC_PUBLIC_SUBNET_2,
                AwsField.NODESETTING_VPC_PUBLIC_SUBNET_3
            ].forEach(field => {
                this.disarmField(field.toString(), false);
            });
        }
    }

    get isVpcTypeExisting(): boolean {
        return this.vpcType === VpcType.EXISTING;
    }

    protected getKeyFromNodeInstance(nodeInstance: string): string {
        return nodeInstance;
    }

    protected getDisplayFromNodeInstance(nodeInstance: string): string {
        return nodeInstance;
    }
}
