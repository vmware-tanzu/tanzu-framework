// Angular imports
import { Component, OnInit } from '@angular/core';
import { Validators } from '@angular/forms';
import { takeUntil } from 'rxjs/operators';
// App imports
import { APIClient } from '../../../../swagger/api-client.service';
import { AppEdition } from 'src/app/shared/constants/branding.constants';
import AppServices from '../../../../shared/service/appServices';
import { AwsField, AwsForm, VpcType } from "../aws-wizard.constants";
import { AwsNodeSettingStepMapping } from './node-setting-step.fieldmapping';
import { AWSNodeAz } from '../../../../swagger/models/aws-node-az.model';
import { AWSSubnet } from '../../../../swagger/models/aws-subnet.model';
import { AzRelatedFieldsArray } from '../aws-wizard.component';
import { ClusterPlan } from '../../wizard/shared/constants/wizard.constants';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { TanzuEventType } from '../../../../shared/service/Messenger';
import { ValidationService } from '../../wizard/shared/validation/validation.service';

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

export const BASTION_HOST_ENABLED = 'yes';
export const BASTION_HOST_DISABLED = 'no';
const swap = (arr, index1, index2) => { [arr[index1], arr[index2]] = [arr[index2], arr[index1]] }

const AZS = [
    AwsField.NODESETTING_AZ_1,
    AwsField.NODESETTING_AZ_2,
    AwsField.NODESETTING_AZ_3,
];
const WORKER_NODE_INSTANCE_TYPES = [
    AwsField.NODESETTING_WORKERTYPE_1,
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

export class NodeSettingStepComponent extends StepFormDirective implements OnInit {
    APP_EDITION: any = AppEdition;

    nodeTypes: Array<string> = [];
    clusterPlan: string;
    vpcType: string;
    nodeAzs: Array<AWSNodeAz>;
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

    displayForm = false;

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
    clusterNameInstruction: string;

    constructor(private validationService: ValidationService,
                private apiClient: APIClient) {
        super();
    }

    private supplyStepMapping(): StepMapping {
        AppServices.fieldMapUtilities.getFieldMapping(AwsField.NODESETTING_CLUSTER_NAME, AwsNodeSettingStepMapping).required =
            AppServices.appDataService.isClusterNameRequired();
        const mapping = AwsNodeSettingStepMapping;
        // dynamically modify the cluster name label based on the type descriptor and whether the cluster name is required
        const clusterNameMapping = AppServices.fieldMapUtilities.getFieldMapping('clusterName', mapping);
        let clusterNameLabel = this.clusterTypeDescriptor.toUpperCase() + ' CLUSTER NAME';
        if (!AppServices.appDataService.isClusterNameRequired()) {
            clusterNameLabel += ' (OPTIONAL)';
        }
        clusterNameMapping.label = clusterNameLabel;

        return mapping;
    }

    private customizeForm() {
        this.registerStepDescriptionTriggers({clusterTypeDescriptor: true, fields: ['controlPlaneSetting']});
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

        /**
         * Whenever aws region selection changes, update AZ subregion
         */
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

        this.registerOnValueChange(AwsField.NODESETTING_CONTROL_PLANE_SETTING, data => {
            if (data === ClusterPlan.DEV) {
                this.setControlPlaneToDev();
            } else if (data === ClusterPlan.PROD) {
                this.setControlPlaneToProd();
            }
            this.updateVpcSubnets();
            this.triggerStepDescriptionChange();
        });
    }

    private subscribeToServices() {
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

    ngOnInit() {
        super.ngOnInit();
        AppServices.fieldMapUtilities.buildForm(this.formGroup, this.wizardName, this.formName, this.supplyStepMapping());
        this.htmlFieldLabels = AppServices.fieldMapUtilities.getFieldLabelMap(this.supplyStepMapping());
        this.storeDefaultLabels(this.supplyStepMapping());
        this.setClusterNameInstruction();
        this.subscribeToServices();
        this.customizeForm();
        this.registerDefaultFileImportedHandler(TanzuEventType.AWS_CONFIG_FILE_IMPORTED, this.supplyStepMapping());
        this.registerDefaultFileImportErrorHandler(TanzuEventType.AWS_CONFIG_FILE_IMPORT_ERROR);

        setTimeout(_ => {
            this.displayForm = true;
            // SHIMON TODO: I'm not sure we need to go get these right here, since existingVpcId should be a trigger field
            const identifier = { wizard: this.wizardName, step: AwsForm.VPC, field: 'existingVpcId' };
            const existingVpcId = AppServices.userDataService.retrieve(identifier);
            if (existingVpcId && existingVpcId.value) {
                AppServices.messenger.publish({
                    type: TanzuEventType.AWS_GET_SUBNETS,
                    payload: { vpcId: existingVpcId.value }
                });
            }
        });
        this.initFormWithSavedData();
    }

    private setClusterNameInstruction() {
        if (AppServices.appDataService.isClusterNameRequired()) {
            this.clusterNameInstruction = 'Specify a name for the ' + this.clusterTypeDescriptor + ' cluster.';
        } else {
            this.clusterNameInstruction = 'Optionally specify a name for the ' + this.clusterTypeDescriptor + ' cluster. ' +
                'If left blank, the installer names the cluster automatically.';
        }
    }

    private setControlPlaneToProd() {
        this.clusterPlan = ClusterPlan.PROD;

        this.disarmField(AwsField.NODESETTING_INSTANCE_TYPE_DEV, true);
        this.resurrectFieldWithSavedValue(AwsField.NODESETTING_INSTANCE_TYPE_PROD,
            [Validators.required, this.validationService.isValidNameInList(this.nodeTypes)],
            this.nodeTypes.length === 1 ? this.nodeTypes[0] : this.formGroup.get(AwsField.NODESETTING_INSTANCE_TYPE_PROD).value,
            { onlySelf: true }
        );
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
            this.setControlWithSavedValue(thisAZ);
        }
        if (!this.modeClusterStandalone) {
            WORKER_NODE_INSTANCE_TYPES.forEach((field, index) => {
                // only populated the worker node instance type if the associated AZ has a value
                if (this.getFieldValue(AZS[index])) {
                    this.resurrectFieldWithSavedValue(field.toString(), [Validators.required]);
                } else {
                    this.resurrectField(field.toString(), [Validators.required]);
                }
            });
        }
    }

    private setControlPlaneToDev() {
        this.clusterPlan = ClusterPlan.DEV;
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
            this.setControlWithSavedValue(AwsField.NODESETTING_AZ_1);
        }
        if (!this.modeClusterStandalone) {
            const hasAz = this.getFieldValue(AwsField.NODESETTING_AZ_1);
            // only set the worker node instance type if the AZ has a value
            if (hasAz) {
                this.resurrectFieldWithSavedValue(AwsField.NODESETTING_WORKERTYPE_1, [Validators.required],
                    this.azNodeTypes.awsNodeAz1.length === 1 ? this.azNodeTypes.awsNodeAz1[0] : '');
            } else {
                this.resurrectField(AwsField.NODESETTING_WORKERTYPE_1, [Validators.required],
                    this.azNodeTypes.awsNodeAz1.length === 1 ? this.azNodeTypes.awsNodeAz1[0] : '');
            }
        }
        this.resurrectFieldWithSavedValue(AwsField.NODESETTING_INSTANCE_TYPE_DEV,
            [Validators.required, this.validationService.isValidNameInList(this.nodeTypes)],
            this.nodeTypes.length === 1 ? this.nodeTypes[0] : this.formGroup.get(AwsField.NODESETTING_INSTANCE_TYPE_DEV).value);
    }

    // returns an array of the other two AZs (used to populate a validator that ensures unique AZs are selected)
    private otherAZs(targetAz: AwsField): AwsField[] {
        return AZS.filter((field, index, arr) => { return field !== targetAz });
    }

    initFormWithSavedData() {
        const devInstanceType = this.getSavedValue(AwsField.NODESETTING_INSTANCE_TYPE_DEV, '');
        const isProdInstanceType = devInstanceType === '';
        this.cardClick(isProdInstanceType ? ClusterPlan.PROD : ClusterPlan.DEV);
        // NOTE: by clicking the right card, the stored values will be used to populate the rest of the fields
    }

    get devInstanceTypeValue() {
        return this.getFieldValue(AwsField.NODESETTING_INSTANCE_TYPE_DEV);
    }

    get prodInstanceTypeValue() {
        return this.getFieldValue(AwsField.NODESETTING_INSTANCE_TYPE_PROD);
    }

    get workerNodeInstanceType1Value() {
        return this.getFieldValue(AwsField.NODESETTING_WORKERTYPE_1);
    }

    get workerNodeInstanceType2Value() {
        return this.getFieldValue(AwsField.NODESETTING_WORKERTYPE_2);
    }

    get workerNodeInstanceType3Value() {
        return this.getFieldValue(AwsField.NODESETTING_WORKERTYPE_3);
    }

    /**
     * @method cardClick
     * sets control plane setting value depending on whether NodeType.DEV or NodeType.PROD
     * card was clicked
     * @param envType
     */
    cardClick(envType: string) {
        this.setControlValueSafely(AwsField.NODESETTING_CONTROL_PLANE_SETTING, envType);
    }

    /**
     * @method getEnvType
     * returns selected control plane setting
     * @returns {string} NodeType.DEV or NodeType.PROD
     */
    getEnvType(): string {
        return this.getFieldValue(AwsField.NODESETTING_CONTROL_PLANE_SETTING);
    }

    /**
     * @method clearAzs
     * helper method used to clear selected AZs from UI controls
     */
    clearAzs() {
        AZS.forEach(az => this.clearControlValue(az));
    }

    /**
     * @method clearSubnets
     * helper method used to clear selected subnets from UI controls
     */
    clearSubnets() {
        VPC_SUBNETS.forEach(vpcSubnet => this.clearControlValue(vpcSubnet));
    }

    /**
     * @method clearSubnetData
     * FilteredAzs does not have iterator, so manually clear subnets
     */
    clearSubnetData() {
        AZS.forEach(az => {
            this.filteredAzs[az.toString()].publicSubnets = [];
            this.filteredAzs[az.toString()].privateSubnets = [];
        });
    }

    /**
     * @method filterSubnetsByAZ
     * helper method that filters larger lists of public and private subnets and returns filtered
     * lists based on match of availability zone name
     * @param $event
     */
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
        const savedValue = this.getSavedValue(subnetField, '');
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
                this.resurrectFieldWithSavedValue(field.toString(), [Validators.required]);
            });
        } else if (this.isClusterPlanDev) {
            // in DEV deployments, only one subnet is used
            [
                AwsField.NODESETTING_VPC_PRIVATE_SUBNET_1,
                AwsField.NODESETTING_VPC_PUBLIC_SUBNET_1,
            ].forEach(field => {
                this.resurrectFieldWithSavedValue(field.toString(), [Validators.required]);
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

    dynamicDescription(): string {
        if (this.isClusterPlanProd) {
            return 'Production cluster selected: 3 node control plane';
        }
        if (this.isClusterPlanDev) {
            return 'Development cluster selected: 1 node control plane';
        }
        return 'Specify the resources backing the ' + this.clusterTypeDescriptor + ' cluster';
    }

    get isClusterPlanProd(): boolean {
        return this.clusterPlan === ClusterPlan.PROD;
    }

    get isClusterPlanDev(): boolean {
        return this.clusterPlan === ClusterPlan.DEV;
    }

    get isVpcTypeExisting(): boolean {
        return this.vpcType === VpcType.EXISTING;
    }

    protected storeUserData() {
        this.storeUserDataFromMapping(this.supplyStepMapping());
        this.storeDefaultDisplayOrder(this.supplyStepMapping());
    }
}
