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
import { Messenger, TkgEventType } from '../../../../shared/service/Messenger';
import { FormMetaDataStore } from '../../wizard/shared/FormMetaDataStore';

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
const PUBLIC_SUBNETS = ['vpcPublicSubnet1', 'vpcPublicSubnet2', 'vpcPublicSubnet3'];
const PRIVATE_SUBNET = ['vpcPrivateSubnet1', 'vpcPrivateSubnet2', 'vpcPrivateSubnet3'];
const VPC_SUBNETS = [...PUBLIC_SUBNETS, ...PRIVATE_SUBNET];
@Component({
    selector: 'app-node-setting-step',
    templateUrl: './node-setting-step.component.html',
    styleUrls: ['./node-setting-step.component.scss']
})
export class NodeSettingStepComponent extends StepFormDirective implements OnInit {

    nodeTypes: Array<string> = [];
    nodeType: string;
    vpcType: string;
    nodeAzs: Array<AWSNodeAz>;
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
        workerNodeInstanceType: [Validators.required],
        bastionHostEnabled: [],
        sshKeyName: [Validators.required],
        createCloudFormation: [],
        clusterName: [this.validationService.isValidClusterName()],
        awsNodeAz1: [Validators.required],
        awsNodeAz2: [Validators.required],
        awsNodeAz3: [Validators.required],
        vpcPublicSubnet1: [],
        vpcPrivateSubnet1: [],
        vpcPublicSubnet2: [],
        vpcPrivateSubnet2: [],
        vpcPublicSubnet3: [],
        vpcPrivateSubnet3: [],
    };

    constructor(private validationService: ValidationService,
        public messenger: Messenger,
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

        this.formGroup.get("bastionHostEnabled").setValue(BASTION_HOST_ENABLED);
        this.formGroup.addControl(
            'machineHealthChecksEnabled',
            new FormControl(true, [])
        );
    }

    ngOnInit() {
        super.ngOnInit();
        this.buildForm();

        this.messenger.getSubject(TkgEventType.AWS_AIRGAPPED_VPC_CHANGE).subscribe(event => {
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
        this.messenger.getSubject(TkgEventType.AWS_REGION_CHANGED)
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

        this.messenger.getSubject(TkgEventType.AWS_VPC_TYPE_CHANGED)
            .subscribe(event => {
                this.vpcType = event.payload.vpcType;
                if (this.vpcType !== 'existing') {
                    this.clearSubnets();
                }

                this.updateVpcSubnets();

                // clear az selection
                this.clearAzs();
                [...AZS, ...VPC_SUBNETS].forEach(attr => this.formGroup.get(attr).updateValueAndValidity());
            });

        this.messenger.getSubject(TkgEventType.AWS_VPC_CHANGED)
            .subscribe(event => {
                this.clearAzs();
                this.clearSubnets();
            });

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
                        this.formGroup.get('devInstanceType').value);
                } else {
                    this.resurrectField('prodInstanceType',
                        [Validators.required, this.validationService.isValidNameInList(this.nodeTypes)],
                        this.formGroup.get('prodInstanceType').value);
                }
                this.resurrectField('workerNodeInstanceType',
                    [Validators.required, this.validationService.isValidNameInList(this.nodeTypes)],
                    this.formGroup.get('workerNodeInstanceType').value);
            });

        AZS.forEach(az => {
            this.formGroup.get(az).valueChanges
                .pipe(
                    takeUntil(this.unsubscribe)
                ).subscribe((val) => {
                    this.filterSubnets(az, val);
                });
        });

        this.registerOnValueChange('controlPlaneSetting', data => {
            if (data === 'dev') {
                this.nodeType = 'dev';
                const prodFields = ['awsNodeAz2', 'awsNodeAz3', 'prodInstanceType'];
                prodFields.forEach(attr => this.disarmField(attr, true));

                this.resurrectField('devInstanceType',
                    [Validators.required, this.validationService.isValidNameInList(this.nodeTypes)],
                    this.formGroup.get('devInstanceType').value);
                this.resurrectField('awsNodeAz1', [Validators.required]);
            } else if (data === 'prod') {
                this.nodeType = 'prod';

                this.disarmField('devInstanceType', true);
                this.resurrectField('prodInstanceType',
                    [Validators.required, this.validationService.isValidNameInList(this.nodeTypes)],
                    this.formGroup.get('prodInstanceType').value);
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
            }
            this.resurrectField('workerNodeInstanceType',
                [Validators.required, this.validationService.isValidNameInList(this.nodeTypes)],
                this.formGroup.get('workerNodeInstanceType').value);

            this.updateVpcSubnets();
        });

        setTimeout(_ => {
            this.displayForm = true;

            const existingVpcId = FormMetaDataStore.getMetaDataItem('vpcForm', 'existingVpcId');
            if (existingVpcId && existingVpcId.displayValue) {
                this.messenger.publish({
                    type: TkgEventType.AWS_GET_SUBNETS,
                    payload: { vpcId: existingVpcId.displayValue }
                });
            }
        });
    }

    setSavedDataAfterLoad() {
        const resetFields = [
            'devInstanceType',
            'prodInstanceType',
            'workerNodeInstanceType'
        ];
        this.cardClick(this.getSavedValue('devInstanceType', '') === '' ? 'prod' : 'dev');
        super.setSavedDataAfterLoad();
        resetFields.forEach(field => this.formGroup.get(field).setValue(''));
    }

    get devInstanceTypeValue() {
        return this.formGroup.controls['devInstanceType'].value;
    }

    get prodInstanceTypeValue() {
        return this.formGroup.controls['prodInstanceType'].value;
    }

    get workerNodeInstanceTypeValue() {
        return this.formGroup.controls['workerNodeInstanceType'].value;
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
