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
import { AppDataService } from 'src/app/shared/service/app-data.service';

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
const swap = (arr, index1, index2) => { [arr[index1], arr[index2]] = [arr[index2], arr[index1]] }

const AZS = ['awsNodeAz1', 'awsNodeAz2', 'awsNodeAz3'];
const WORKER_NODE_INSTANCE_TYPES = ['workerNodeInstanceType1', 'workerNodeInstanceType2', 'workerNodeInstanceType3'];
const PUBLIC_SUBNETS = ['vpcPublicSubnet1', 'vpcPublicSubnet2', 'vpcPublicSubnet3'];
const PRIVATE_SUBNET = ['vpcPrivateSubnet1', 'vpcPrivateSubnet2', 'vpcPrivateSubnet3'];
const VPC_SUBNETS = [...PUBLIC_SUBNETS, ...PRIVATE_SUBNET];
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

    constructor(private validationService: ValidationService, appDataService: AppDataService,
        private apiClient: APIClient,
        public awsWizardFormService: AwsWizardFormService) {
        super(appDataService);
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

        this.formGroup.get("bastionHostEnabled").setValue(BASTION_HOST_ENABLED);
        this.formGroup.addControl(
            'machineHealthChecksEnabled',
            new FormControl(true, [])
        );
        this.formGroup.addControl(
            'createCloudFormation',
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
                    this.formGroup.controls[f].setValue('');
                    this.formGroup.controls[f].disable();
                })
            } else {        // public subnet IDs are required
                PUBLIC_SUBNETS.forEach(f => {
                    this.formGroup.controls[f].enable();
                })
            }
        });

        /**
         * Whenever aws region selection changes, update AZ subregion
         */
        Broker.messenger.getSubject(TkgEventType.AWS_REGION_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                if (this.formGroup.get('awsNodeAz1')) {
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
                if (this.vpcType !== 'existing') {
                    this.clearSubnets();
                }

                this.updateVpcSubnets();

                // clear az selection
                this.clearAzs();
                [...AZS, ...WORKER_NODE_INSTANCE_TYPES, ...VPC_SUBNETS].forEach(attr => this.formGroup.get(attr).updateValueAndValidity());
            });

        Broker.messenger.getSubject(TkgEventType.AWS_VPC_CHANGED)
            .subscribe(event => {
                this.clearAzs();
                this.clearSubnets();
            });

        if (this.edition !== AppEdition.TKG) {
            this.resurrectField('clusterName',
                [Validators.required, this.validationService.isValidClusterName()],
                this.formGroup.get('clusterName').value);
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
                AZS.forEach(az => this.filterSubnets(az, this.formGroup.get(az).value));
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

                if (this.nodeType === 'dev') {
                    this.resurrectField('devInstanceType',
                        [Validators.required, this.validationService.isValidNameInList(this.nodeTypes)],
                        this.nodeTypes.length === 1 ? this.nodeTypes[0] : this.formGroup.get('devInstanceType').value);
                } else {
                    this.resurrectField('prodInstanceType',
                        [Validators.required, this.validationService.isValidNameInList(this.nodeTypes)],
                        this.nodeTypes.length === 1 ? this.nodeTypes[0] : this.formGroup.get('prodInstanceType').value);
                }
            });

        AZS.forEach((az, index) => {
            this.formGroup.get(az).valueChanges
                .pipe(
                    takeUntil(this.unsubscribe)
                ).subscribe((val) => {
                    this.filterSubnets(az, val);
                    this.updateWorkerNodeInstanceTypes(az, val, index);
                });
        });

        this.registerOnValueChange('controlPlaneSetting', data => {
            if (data === 'dev') {
                this.nodeType = 'dev';
                const prodFields = ['awsNodeAz2', 'awsNodeAz3', 'workerNodeInstanceType2', 'workerNodeInstanceType3', 'prodInstanceType'];

                prodFields.forEach(attr => this.disarmField(attr, true));
                if (this.nodeAzs && this.nodeAzs.length === 1) {
                    this.formGroup.get('awsNodeAz1').setValue(this.nodeAzs[0].name);
                }

                if (!this.modeClusterStandalone) {
                    this.resurrectField('workerNodeInstanceType1', [Validators.required],
                        this.azNodeTypes.awsNodeAz1.length === 1 ? this.azNodeTypes.awsNodeAz1[0] : '');
                }

                this.resurrectField('devInstanceType',
                    [Validators.required, this.validationService.isValidNameInList(this.nodeTypes)],
                    this.nodeTypes.length === 1 ? this.nodeTypes[0] : this.formGroup.get('devInstanceType').value);
            } else if (data === 'prod') {
                this.nodeType = 'prod';

                this.disarmField('devInstanceType', true);
                this.resurrectField('prodInstanceType',
                    [Validators.required, this.validationService.isValidNameInList(this.nodeTypes)],
                    this.nodeTypes.length === 1 ? this.nodeTypes[0] : this.formGroup.get('prodInstanceType').value);
                const azNew = [...AZS];
                for (let i = 0; i < AZS.length; i++) {
                    swap(azNew, i, 0);
                    this.formGroup.get(azNew[0]).setValidators([
                        Validators.required,
                        this.validationService.isUniqueAz([
                            this.formGroup.get(azNew[1]),
                            this.formGroup.get(azNew[2])])
                    ]);
                }
                if (!this.modeClusterStandalone) {
                    WORKER_NODE_INSTANCE_TYPES.forEach(field => this.resurrectField(field, [Validators.required]));
                }
            }

            this.updateVpcSubnets();
        });

        setTimeout(_ => {
            this.displayForm = true;

            const existingVpcId = FormMetaDataStore.getMetaDataItem('vpcForm', 'existingVpcId');
            if (existingVpcId && existingVpcId.displayValue) {
                Broker.messenger.publish({
                    type: TkgEventType.AWS_GET_SUBNETS,
                    payload: { vpcId: existingVpcId.displayValue }
                });
            }
        });
    }

    setSavedDataAfterLoad() {
        this.cardClick(this.getSavedValue('devInstanceType', '') === '' ? 'prod' : 'dev');
        super.setSavedDataAfterLoad();
        if (this.getSavedValue('devInstanceType', '') === '') { // prod
            this.formGroup.get('devInstanceType').setValue('');
            this.formGroup.get('prodInstanceType').setValue(this.nodeTypes.length === 1 ? this.nodeTypes[0] : '');
        } else {
            this.formGroup.get('devInstanceType').setValue(this.nodeTypes.length === 1 ? this.nodeTypes[0] : '');
            this.formGroup.get('prodInstanceType').setValue('');
        }
    }

    get devInstanceTypeValue() {
        return this.formGroup.controls['devInstanceType'].value;
    }

    get prodInstanceTypeValue() {
        return this.formGroup.controls['prodInstanceType'].value;
    }

    get workerNodeInstanceType1Value() {
        return this.formGroup.controls['workerNodeInstanceType1'].value;
    }

    get workerNodeInstanceType2Value() {
        return this.formGroup.controls['workerNodeInstanceType2'].value;
    }

    get workerNodeInstanceType3Value() {
        return this.formGroup.controls['workerNodeInstanceType3'].value;
    }

    /**
     * @method cardClick
     * sets control plane setting value depending on whether 'dev' or 'prod'
     * card was clicked
     * @param envType
     */
    cardClick(envType: string) {
        this.formGroup.controls['controlPlaneSetting'].setValue(envType);
    }

    /**
     * @method getEnvType
     * returns selected control plane setting
     * @returns {string} 'dev' or 'prod'
     */
    getEnvType(): string {
        return this.formGroup.controls['controlPlaneSetting'].value;
    }

    /**
     * @method clearAzs
     * helper method used to clear selected AZs from UI controls
     */
    clearAzs() {
        AZS.forEach(az => this.formGroup.get(az).setValue(''));
    }

    /**
     * @method clearSubnets
     * helper method used to clear selected subnets from UI controls
     */
    clearSubnets() {
        VPC_SUBNETS.forEach(vpcSubnet => this.formGroup.get(vpcSubnet).setValue(''));
    }

    /**
     * @method clearSubnetData
     * FilteredAzs does not have iterator, so manually clear subnets
     */
    clearSubnetData() {
        AZS.forEach(az => {
            this.filteredAzs[az].publicSubnets = [];
            this.filteredAzs[az].privateSubnets = [];
        });
    }

    /**
     * @method filterSubnets
     * helper method that filters larger lists of public and private subnets and returns filtered
     * lists based on match of availability zone name
     * @param $event
     */
    filterSubnets(azControlName, az): void {
        if (this.vpcType === 'existing' && azControlName !== '' && az !== '') {
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
            const instanceType = this.getSavedValue(field, '');
            this.formGroup.get(field).setValue(instanceType);
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
                            this.formGroup.get(WORKER_NODE_INSTANCE_TYPES[index]).setValue(nodeTypes[0]);
                        }

                        if (this.vpcType === 'existing') {
                            if (this.filteredAzs[AZS[index]].publicSubnets.length === 1) {
                                this.formGroup.get(PUBLIC_SUBNETS[index]).setValue(this.filteredAzs[AZS[index]].publicSubnets[0].id);
                            }
                            if (this.filteredAzs[AZS[index]].privateSubnets.length === 1) {
                                this.formGroup.get(PRIVATE_SUBNET[index]).setValue(this.filteredAzs[AZS[index]].privateSubnets[0].id);
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
            const typeOfSubnet = vpcSubnet.indexOf('public') > -1 ? 'publicSubnets' : 'privateSubnets';
            const subnet = this[typeOfSubnet].find(x => x.cidr === this.getSavedValue(vpcSubnet, ''));
            this.formGroup.get(vpcSubnet).setValue(subnet ? subnet.id : '');
        });
    }

    updateVpcSubnets() {
        if (this.vpcType !== "existing") {   // validations should be disabled for all public/private subnets
            [1, 2, 3].forEach(i => {
                this.disarmField('vpcPublicSubnet' + i, true);
                this.disarmField('vpcPrivateSubnet' + i, true);
            });
            return;
        }

        // First enable validators on all fields
        [1, 2, 3].forEach(i => {
            this.resurrectField('vpcPublicSubnet' + i, [Validators.required]);
            this.resurrectField('vpcPrivateSubnet' + i, [Validators.required]);
        });

        // both private and public fields will be shown
        if (this.nodeType === "dev") {   // 2 & 3 should be disarmed
            [2, 3].forEach(i => {
                this.disarmField('vpcPublicSubnet' + i, true);
                this.disarmField('vpcPrivateSubnet' + i, true);
            });
        }

        if (this.airgappedVPC) {
            [1, 2, 3].forEach(i => {
                this.disarmField('vpcPublicSubnet' + i, true);
            });
        }
    }

}
