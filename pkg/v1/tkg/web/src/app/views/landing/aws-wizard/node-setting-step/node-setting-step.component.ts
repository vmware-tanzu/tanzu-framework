/**
 * Angular Modules
 */
import { Component, OnInit } from '@angular/core';
import { FormControl, Validators } from '@angular/forms';
import { takeUntil } from 'rxjs/operators';
/**
 * App imports
 */
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { AWSNodeAz } from '../../../../swagger/models/aws-node-az.model';
import { AWSSubnet } from '../../../../swagger/models/aws-subnet.model';
import { AwsWizardFormService } from '../../../../shared/service/aws-wizard-form.service';
import { TkgEventType } from '../../../../shared/service/Messenger';
import { FormMetaDataStore } from '../../wizard/shared/FormMetaDataStore';
import { APIClient } from '../../../../swagger/api-client.service';
import Broker from 'src/app/shared/service/broker';
import { AppEdition } from 'src/app/shared/constants/branding.constants';
import { FormUtils } from '../../wizard/shared/utils/form-utils';
import { AwsField, AwsForm } from "../aws-wizard.constants";
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { AwsNodeSettingStepMapping } from './node-setting-step.fieldmapping';

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

enum NodeType {
    DEV = 'dev',
    PROD = 'prod'
}

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
    nodeType: string;
    vpcType: string;
    nodeAzs: Array<AWSNodeAz>;
    azNodeTypes: AzNodeTypes = {
        awsNodeAz1: [],
        awsNodeAz2: [],
        awsNodeAz3: []
    };

    publicSubnets: Array<AWSSubnet>;
    privateSubnets: Array<AWSSubnet>;

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

    constructor(private validationService: ValidationService,
        private fieldMapUtilities: FieldMapUtilities,
        private apiClient: APIClient,
        public awsWizardFormService: AwsWizardFormService) {
        super();
    }

    ngOnInit() {
        super.ngOnInit();
        this.fieldMapUtilities.buildForm(this.formGroup, this.formName, AwsNodeSettingStepMapping);

        Broker.messenger.getSubject(TkgEventType.AWS_AIRGAPPED_VPC_CHANGE).subscribe(event => {
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
        Broker.messenger.getSubject(TkgEventType.AWS_REGION_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                if (this.formGroup.get(AwsField.NODESETTING_AZ_1)) {
                    this.publicSubnets = [];
                    this.privateSubnets = [];

                    this.clearSubnetData();
                    this.clearAzs();
                    this.clearSubnets();
                }
            });

        Broker.messenger.getSubject(TkgEventType.AWS_VPC_TYPE_CHANGED)
            .subscribe(event => {
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

        Broker.messenger.getSubject(TkgEventType.AWS_VPC_CHANGED)
            .subscribe(event => {
                this.clearAzs();
                this.clearSubnets();
            });

        if (this.edition !== AppEdition.TKG) {
            this.resurrectField(AwsField.NODESETTING_CLUSTER_NAME,
                [Validators.required, this.validationService.isValidClusterName()],
                this.getFieldValue(AwsField.NODESETTING_CLUSTER_NAME));
        }

        this.awsWizardFormService.getErrorStream(TkgEventType.AWS_GET_AVAILABILITY_ZONES)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(error => {
                this.errorNotification = error;
            });

        this.awsWizardFormService.getErrorStream(TkgEventType.AWS_GET_SUBNETS)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(error => {
                this.errorNotification = error;
            });

        this.awsWizardFormService.getErrorStream(TkgEventType.AWS_GET_NODE_TYPES)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(error => {
                this.errorNotification = error;
            });

        this.awsWizardFormService.getDataStream(TkgEventType.AWS_GET_AVAILABILITY_ZONES)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((availabilityZones: Array<AWSNodeAz>) => {
                this.nodeAzs = availabilityZones;
            });

        this.awsWizardFormService.getDataStream(TkgEventType.AWS_GET_SUBNETS)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((subnets: Array<AWSSubnet>) => {
                this.publicSubnets = subnets.filter(obj => {
                    return obj.isPublic === true
                });
                this.privateSubnets = subnets.filter(obj => {
                    return obj.isPublic === false
                });
                AZS.forEach(az => {
                    this.filterSubnets(az, this.getFieldValue(az));
                });
                this.setSubnetFieldsFromSavedValues();
            });

        this.awsWizardFormService.getDataStream(TkgEventType.AWS_GET_NODE_TYPES)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((nodeTypes: Array<string>) => {
                this.nodeTypes = nodeTypes.sort();

                // The validation is based on the value of this.nodeTypes. Whenever we update this.nodeTypes,
                // the corresponding validation should be updated as well. e.g. the users came to the node-settings
                // step before the api responses. Then an empty array will be passed to the validation isValidNameInList.
                // It will cause the selected option to be invalid all the time.

                if (this.nodeType === NodeType.DEV) {
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
            });

        AZS.forEach((az, index) => {
            this.registerOnValueChange(az, (val) => {
                this.filterSubnets(az, val);
                this.setSubnetFieldsWithOnlyOneOption(az);
                this.updateWorkerNodeInstanceTypes(az, val, index);
            });
        });

        this.registerOnValueChange(AwsField.NODESETTING_CONTROL_PLANE_SETTING, data => {
            if (data === NodeType.DEV) {
                this.setControlPlaneToDev();
            } else if (data === NodeType.PROD) {
                this.setControlPlaneToProd();
            }
            this.updateVpcSubnets();
        });

        setTimeout(_ => {
            this.displayForm = true;
            const existingVpcId = FormMetaDataStore.getMetaDataItem(AwsForm.VPC, 'existingVpcId');
            if (existingVpcId && existingVpcId.displayValue) {
                Broker.messenger.publish({
                    type: TkgEventType.AWS_GET_SUBNETS,
                    payload: { vpcId: existingVpcId.displayValue }
                });
            }
        });
        this.initFormWithSavedData();
    }

    private setControlPlaneToProd() {
        this.nodeType = NodeType.PROD;

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
            WORKER_NODE_INSTANCE_TYPES.forEach(field => this.resurrectFieldWithSavedValue(field.toString(), [Validators.required]));
        }
    }

    private setControlPlaneToDev() {
        this.nodeType = NodeType.DEV;
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
            this.resurrectFieldWithSavedValue(AwsField.NODESETTING_WORKERTYPE_1, [Validators.required],
                this.azNodeTypes.awsNodeAz1.length === 1 ? this.azNodeTypes.awsNodeAz1[0] : '');
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
        const prodInstanceType = this.getSavedValue(AwsField.NODESETTING_INSTANCE_TYPE_PROD, '');
        const isProdInstanceType = devInstanceType === '';
        this.cardClick(isProdInstanceType ? NodeType.PROD : NodeType.DEV);
        super.initFormWithSavedData();
        // because it's in its own component, the enable audit logging field does not get initialized in the above call to
        // super.initFormWithSavedData()
        setTimeout( () => {
            this.setControlWithSavedValue('enableAuditLogging', false);
        })

        if (isProdInstanceType) {
            const nodeType = this.nodeTypes.length === 1 ? this.nodeTypes[0] : prodInstanceType;
            this.clearControlValue(AwsField.NODESETTING_INSTANCE_TYPE_DEV);
            this.setControlValueSafely(AwsField.NODESETTING_INSTANCE_TYPE_PROD, nodeType);
        } else {
            const nodeType = this.nodeTypes.length === 1 ? this.nodeTypes[0] : devInstanceType;
            this.setControlValueSafely(AwsField.NODESETTING_INSTANCE_TYPE_DEV, nodeType);
            this.clearControlValue(AwsField.NODESETTING_INSTANCE_TYPE_PROD);
        }
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
     * @method filterSubnets
     * helper method that filters larger lists of public and private subnets and returns filtered
     * lists based on match of availability zone name
     * @param $event
     */
    filterSubnets(azControlName, az): void {
        if (this.vpcType === vpcType.EXISTING && azControlName !== '' && az !== '') {
            this.filteredAzs[azControlName].publicSubnets = this.publicSubnets.filter(obj => {
                return obj.availabilityZoneName === az;
            });
            this.filteredAzs[azControlName].privateSubnets = this.privateSubnets.filter(obj => {
                return obj.availabilityZoneName === az;
            });
        }
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

    setSavedWorkerNodeInstanceTypes(): void {
        WORKER_NODE_INSTANCE_TYPES.forEach(field => {
            const instanceType = this.getSavedValue(field.toString(), '');
            this.formGroup.get(field.toString()).setValue(instanceType);
        });
    }

    /**
     * @method updateWorkerNodeInstanceTypes
     * @param azWorkerNodeKey - the key of the worker node list in the azNodeTypes list to update
     * @param availabilityZone - the availability zone name to retrieve node types against
     * Updates available worker node instance type list per availability zone. API takes the availability zone name
     * and returns list of node instance types available to that zone.
     */
    private updateWorkerNodeInstanceTypes(azWorkerNodeKey, availabilityZone, index) {
        if (availabilityZone) {
            this.apiClient.getAWSNodeTypes({
                az: availabilityZone
            })
                .pipe(takeUntil(this.unsubscribe))
                .subscribe(
                    ((nodeTypes) => {
                        this.azNodeTypes[azWorkerNodeKey] = nodeTypes;
                        if (nodeTypes.length === 1) {
                            this.formGroup.get(WORKER_NODE_INSTANCE_TYPES[index].toString()).setValue(nodeTypes[0]);
                        }
                    }),
                    ((err) => {
                        const error = err.error.message || err.message || JSON.stringify(err);
                        this.errorNotification = `Unable to retrieve worker node instance types. ${error}`;
                    })
                );
        } else {
            this.azNodeTypes[azWorkerNodeKey] = [];
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

        if (this.nodeType === NodeType.PROD) {
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
        } else if (this.nodeType === NodeType.DEV) {
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

    protected dynamicDescription(): string {
        const ctlPlaneFlavor = this.getFieldValue('controlPlaneSetting', true);
        if (ctlPlaneFlavor) {
            let mode = 'Development cluster selected: 1 node control plane';
            if (ctlPlaneFlavor === 'prod') {
                mode = 'Production cluster selected: 3 node control plane';
            }
            return mode;
        }
        return `Specify the resources backing the ${this.clusterTypeDescriptor} cluster`;
    }
}
