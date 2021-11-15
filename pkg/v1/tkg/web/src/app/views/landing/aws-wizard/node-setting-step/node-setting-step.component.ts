/**
 * Angular Modules
 */
import { Component, OnInit } from '@angular/core';
import {
    Validators,
    FormControl
} from '@angular/forms';
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
import {AwsField, AwsForm} from "../aws-wizard.constants";

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

    // TODO: modify this to use aws-wizard.constants.ts' AwsField enum
    commonFieldMap: { [key: string]: Array<any> } = {
        controlPlaneSetting: [Validators.required],
        devInstanceType: [Validators.required],
        prodInstanceType: [Validators.required],
        bastionHostEnabled: [],
        sshKeyName: [Validators.required],
        clusterName: [this.validationService.isValidClusterName()],
        awsNodeAz1: [Validators.required],
        awsNodeAz2: [Validators.required],
        awsNodeAz3: [Validators.required],
        workerNodeInstanceType1: [],
        vpcPublicSubnet1: [],
        vpcPrivateSubnet1: [],
        workerNodeInstanceType2: [],
        vpcPublicSubnet2: [],
        vpcPrivateSubnet2: [],
        workerNodeInstanceType3: [],
        vpcPublicSubnet3: [],
        vpcPrivateSubnet3: [],
    };

    constructor(private validationService: ValidationService,
        private apiClient: APIClient,
        public awsWizardFormService: AwsWizardFormService) {
        super();
    }

    buildForm() {
        // key is field name, value is validation rules
        for (const key in this.commonFieldMap) {
            if (key) {
                this.formGroup.addControl(
                    key,
                    new FormControl('', this.commonFieldMap[key])
                );
            }
        }

        this.formGroup.get(AwsField.NODESETTING_BASTION_HOST_ENABLED).setValue(BASTION_HOST_ENABLED);
        this.formGroup.addControl(
            AwsField.NODESETTING_MACHINE_HEALTH_CHECKS_ENABLED,
            new FormControl(true, [])
        );
        this.formGroup.addControl(
            AwsField.NODESETTING_CREATE_CLOUD_FORMATION,
            new FormControl(true, [])
        );
    }

    ngOnInit() {
        super.ngOnInit();
        this.buildForm();

        Broker.messenger.getSubject(TkgEventType.AWS_AIRGAPPED_VPC_CHANGE).subscribe(event => {
            this.airgappedVPC = event.payload;
            if (this.airgappedVPC) { // public subnet IDs shouldn't be provided
                PUBLIC_SUBNETS.forEach(f => {
                    this.formGroup.controls[f.toString()].setValue('');
                    this.formGroup.controls[f.toString()].disable();
                })
            } else {        // public subnet IDs are required
                PUBLIC_SUBNETS.forEach(f => {
                    this.formGroup.controls[f.toString()].enable();
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
                [...AZS, ...WORKER_NODE_INSTANCE_TYPES, ...VPC_SUBNETS].forEach(attr => this.formGroup.get(attr.toString()).updateValueAndValidity());
            });

        Broker.messenger.getSubject(TkgEventType.AWS_VPC_CHANGED)
            .subscribe(event => {
                this.clearAzs();
                this.clearSubnets();
            });

        if (this.edition !== AppEdition.TKG) {
            this.resurrectField(AwsField.NODESETTING_CLUSTER_NAME,
                [Validators.required, this.validationService.isValidClusterName()],
                this.formGroup.get(AwsField.NODESETTING_CLUSTER_NAME).value);
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
                AZS.forEach(az => this.filterSubnets(az, this.formGroup.get(az.toString()).value));
                this.setSavedSubnets();
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
            this.formGroup.get(az.toString()).valueChanges
                .pipe(
                    takeUntil(this.unsubscribe)
                ).subscribe((val) => {
                    this.filterSubnets(az, val);
                    this.updateWorkerNodeInstanceTypes(az, val, index);
                });
        });

        this.registerOnValueChange(AwsField.NODESETTING_CONTROL_PLANE_SETTING, data => {
            if (data === NodeType.DEV) {
                this.nodeType = NodeType.DEV;
                const prodFields = [AwsField.NODESETTING_AZ_2, AwsField.NODESETTING_AZ_3, AwsField.NODESETTING_WORKERTYPE_2, AwsField.NODESETTING_WORKERTYPE_3,
                    AwsField.NODESETTING_INSTANCE_TYPE_PROD];
                prodFields.forEach(attr => this.disarmField(attr.toString(), true));
                if (this.nodeAzs && this.nodeAzs.length === 1) {
                    this.formGroup.get(AwsField.NODESETTING_AZ_1).setValue(this.nodeAzs[0].name);
                }
                if (!this.modeClusterStandalone) {
                    this.resurrectField(AwsField.NODESETTING_WORKERTYPE_1, [Validators.required],
                        this.azNodeTypes.awsNodeAz1.length === 1 ? this.azNodeTypes.awsNodeAz1[0] : '');
                }
                this.resurrectField(AwsField.NODESETTING_INSTANCE_TYPE_DEV,
                    [Validators.required, this.validationService.isValidNameInList(this.nodeTypes)],
                    this.nodeTypes.length === 1 ? this.nodeTypes[0] : this.formGroup.get(AwsField.NODESETTING_INSTANCE_TYPE_DEV).value);
            } else if (data === NodeType.PROD) {
                this.nodeType = NodeType.PROD;

                this.disarmField(AwsField.NODESETTING_INSTANCE_TYPE_DEV, true);
                this.resurrectField(AwsField.NODESETTING_INSTANCE_TYPE_PROD,
                    [Validators.required, this.validationService.isValidNameInList(this.nodeTypes)],
                    this.nodeTypes.length === 1 ? this.nodeTypes[0] : this.formGroup.get(AwsField.NODESETTING_INSTANCE_TYPE_PROD).value);
                const azNew = [...AZS];
                for (let i = 0; i < AZS.length; i++) {
                    swap(azNew, i, 0);
                    this.formGroup.get(azNew[0].toString()).setValidators([
                        Validators.required,
                        this.validationService.isUniqueAz([
                            this.formGroup.get(azNew[1].toString()),
                            this.formGroup.get(azNew[2].toString())])
                    ]);
                }
                if (!this.modeClusterStandalone) {
                    WORKER_NODE_INSTANCE_TYPES.forEach(field => this.resurrectField(field.toString(), [Validators.required]));
                }
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

    initFormWithSavedData() {
        const devInstanceType = this.getSavedValue(AwsField.NODESETTING_INSTANCE_TYPE_DEV, '');
        const prodInstanceType = this.getSavedValue(AwsField.NODESETTING_INSTANCE_TYPE_PROD, '');
        const isProdInstanceType = devInstanceType === '';
        this.cardClick(isProdInstanceType ? NodeType.PROD : NodeType.DEV);
        super.initFormWithSavedData();
        if (isProdInstanceType) {
            this.formGroup.get(AwsField.NODESETTING_INSTANCE_TYPE_DEV).setValue('');
            this.formGroup.get(AwsField.NODESETTING_INSTANCE_TYPE_PROD).setValue(this.nodeTypes.length === 1 ? this.nodeTypes[0] : prodInstanceType);
        } else {
            this.formGroup.get(AwsField.NODESETTING_INSTANCE_TYPE_DEV).setValue(this.nodeTypes.length === 1 ? this.nodeTypes[0] : devInstanceType);
            this.formGroup.get(AwsField.NODESETTING_INSTANCE_TYPE_PROD).setValue('');
        }
    }

    get devInstanceTypeValue() {
        return this.formGroup.controls[AwsField.NODESETTING_INSTANCE_TYPE_DEV].value;
    }

    get prodInstanceTypeValue() {
        return this.formGroup.controls[AwsField.NODESETTING_INSTANCE_TYPE_PROD].value;
    }

    get workerNodeInstanceType1Value() {
        return this.formGroup.controls[AwsField.NODESETTING_WORKERTYPE_1].value;
    }

    get workerNodeInstanceType2Value() {
        return this.formGroup.controls[AwsField.NODESETTING_WORKERTYPE_2].value;
    }

    get workerNodeInstanceType3Value() {
        return this.formGroup.controls[AwsField.NODESETTING_WORKERTYPE_3].value;
    }

    /**
     * @method cardClick
     * sets control plane setting value depending on whether NodeType.DEV or NodeType.PROD
     * card was clicked
     * @param envType
     */
    cardClick(envType: string) {
        this.formGroup.controls[AwsField.NODESETTING_CONTROL_PLANE_SETTING].setValue(envType);
    }

    /**
     * @method getEnvType
     * returns selected control plane setting
     * @returns {string} NodeType.DEV or NodeType.PROD
     */
    getEnvType(): string {
        return this.formGroup.controls[AwsField.NODESETTING_CONTROL_PLANE_SETTING].value;
    }

    /**
     * @method clearAzs
     * helper method used to clear selected AZs from UI controls
     */
    clearAzs() {
        AZS.forEach(az => this.formGroup.get(az.toString()).setValue(''));
    }

    /**
     * @method clearSubnets
     * helper method used to clear selected subnets from UI controls
     */
    clearSubnets() {
        VPC_SUBNETS.forEach(vpcSubnet => this.formGroup.get(vpcSubnet.toString()).setValue(''));
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
    updateWorkerNodeInstanceTypes(azWorkerNodeKey, availabilityZone, index) {
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

                        if (this.vpcType === vpcType.EXISTING) {
                            if (this.filteredAzs[AZS[index].toString()].publicSubnets.length === 1) {
                                this.formGroup.get(PUBLIC_SUBNETS[index].toString()).setValue(this.filteredAzs[AZS[index].toString()].publicSubnets[0].id);
                            }
                            if (this.filteredAzs[AZS[index].toString()].privateSubnets.length === 1) {
                                this.formGroup.get(PRIVATE_SUBNET[index].toString()).setValue(this.filteredAzs[AZS[index].toString()].privateSubnets[0].id);
                            }
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

    setSavedSubnets(): void {
        VPC_SUBNETS.forEach(vpcSubnet => {
            const typeOfSubnet = vpcSubnet.toString().indexOf('public') > -1 ? 'publicSubnets' : 'privateSubnets';
            const subnet = this[typeOfSubnet].find(x => x.cidr === this.getSavedValue(vpcSubnet.toString(), ''));
            this.formGroup.get(vpcSubnet.toString()).setValue(subnet ? subnet.id : '');
        });
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
                this.disarmField(field.toString(), true);
            });
            return;
        }

        // Start by enabling validators on all fields
        [
            AwsField.NODESETTING_VPC_PRIVATE_SUBNET_1,
            AwsField.NODESETTING_VPC_PRIVATE_SUBNET_2,
            AwsField.NODESETTING_VPC_PRIVATE_SUBNET_3,
            AwsField.NODESETTING_VPC_PUBLIC_SUBNET_1,
            AwsField.NODESETTING_VPC_PUBLIC_SUBNET_2,
            AwsField.NODESETTING_VPC_PUBLIC_SUBNET_3
        ].forEach(field => {
            this.resurrectField(field.toString(), [Validators.required]);
        });

        // in DEV deployments, only one subnet is used
        if (this.nodeType === NodeType.DEV) {   // 2 & 3 should be disarmed
            [
                AwsField.NODESETTING_VPC_PRIVATE_SUBNET_2,
                AwsField.NODESETTING_VPC_PRIVATE_SUBNET_3,
                AwsField.NODESETTING_VPC_PUBLIC_SUBNET_2,
                AwsField.NODESETTING_VPC_PUBLIC_SUBNET_3
            ].forEach(field => {
                this.disarmField(field.toString(), true);
            });
        }

        if (this.airgappedVPC) {
            [
                AwsField.NODESETTING_VPC_PUBLIC_SUBNET_1,
                AwsField.NODESETTING_VPC_PUBLIC_SUBNET_2,
                AwsField.NODESETTING_VPC_PUBLIC_SUBNET_3
            ].forEach(field => {
                this.disarmField(field.toString(), true);
            });
        }
    }
}
