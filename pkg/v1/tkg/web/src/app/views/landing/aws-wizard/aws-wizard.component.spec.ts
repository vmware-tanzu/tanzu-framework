// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { FormBuilder, FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
// App imports
import { APIClient } from '../../../swagger/api-client.service';
import AppServices from 'src/app/shared/service/appServices';
import { AwsField, AwsForm, VpcType } from './aws-wizard.constants';
import { AwsWizardComponent } from './aws-wizard.component';
import { CeipField } from '../wizard/shared/components/steps/ceip-step/ceip-step.fieldmapping';
import { ClusterType, WizardForm } from "../wizard/shared/constants/wizard.constants";
import { FieldMapUtilities } from '../wizard/shared/field-mapping/FieldMapUtilities';
import { Messenger } from 'src/app/shared/service/Messenger';
import { MetadataField } from '../wizard/shared/components/steps/metadata-step/metadata-step.fieldmapping';
import { NetworkField } from '../wizard/shared/components/steps/network-step/network-step.fieldmapping';
import { NodeSettingStepComponent } from './node-setting-step/node-setting-step.component';
import { SharedModule } from '../../../shared/shared.module';
import { ValidationService } from '../wizard/shared/validation/validation.service';
import { NodeSettingField } from '../wizard/shared/components/steps/node-setting-step/node-setting-step.fieldmapping';

describe('AwsWizardComponent', () => {
    let component: AwsWizardComponent;
    let fixture: ComponentFixture<AwsWizardComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule,
                ReactiveFormsModule,
                BrowserAnimationsModule,
                RouterTestingModule,
                SharedModule
            ],
            providers: [
                APIClient,
                FormBuilder,
                ValidationService
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            declarations: [ AwsWizardComponent ]
        })
        .compileComponents();
    }));

    beforeEach(() => {
        AppServices.messenger = new Messenger();
        const fb = new FormBuilder();
        fixture = TestBed.createComponent(AwsWizardComponent);
        component = fixture.componentInstance;
        component.form = fb.group({
            awsProviderForm: fb.group({
                accessKeyID: [''],
                region: [''],
                secretAccessKey: [''],
            }),
            vpcForm: fb.group({
                vpc: [''],
                publicNodeCidr: [''],
                privateNodeCidr: [''],
                awsNodeAz: [''],
                vpcType: ['']
            }),
            awsNodeSettingForm: fb.group({
                awsNodeAz1: [''],
                awsNodeAz2: [''],
                awsNodeAz3: [''],
                bastionHostEnabled: [false],
                controlPlaneSetting: [''],
                devInstanceType: [''],
                machineHealthChecksEnabled: [false],
                createCloudFormation: [false],
                workerNodeInstanceType: [''],
                workerNodeInstanceType2: [''],
                workerNodeInstanceType3: [''],
                clusterName: [''],
                sshKeyName: ['']
            }),
            metadataForm: fb.group({
                clusterDescription: [''],
                clusterLabels: [new Map()],
                clusterLocation: [''],
            }),
            networkForm: fb.group({
                clusterPodCidr: [''],
                clusterServiceCidr: [''],
                cniType: ['']
            }),
            identityForm: fb.group({
            }),
            amiOrgIdForm: fb.group({
            }),
            ceipOptInForm: fb.group({
                ceipOptIn: [true]
            }),
            osImageForm: fb.group({
            })
        });
        component.clusterTypeDescriptor = '' + ClusterType.Management;
        component.title = 'Tanzu';
        fixture.detectChanges();
    });

    afterEach(() => {
        fixture.destroy();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should return management cluster name', () => {
        const stepInstance = TestBed.createComponent(NodeSettingStepComponent).componentInstance;
        component.registerStep(AwsForm.NODESETTING, stepInstance);
        stepInstance.formGroup.addControl('clusterName', new FormControl(''));

        component.form.get(AwsForm.NODESETTING).get('clusterName').setValue('mylocalTestName');
        expect(component.getMCName()).toBe('mylocalTestName');
    });

    it('should create API payload', () => {
        const mappings = [
            [AwsForm.PROVIDER, AwsField.PROVIDER_ACCESS_KEY, 'aws-access-key-id-12345'],
            [AwsForm.PROVIDER, AwsField.PROVIDER_REGION, 'US-WEST'],
            [AwsForm.PROVIDER, AwsField.PROVIDER_SECRET_ACCESS_KEY, 'My-AWS-Secret-Access-Key'],
            [AwsForm.VPC, AwsField.VPC_NEW_CIDR, '10.0.0.0/16'],
            [AwsForm.VPC, AwsField.VPC_TYPE, VpcType.NEW],
            [AwsForm.NODESETTING, AwsField.NODESETTING_AZ_1, 'us-west-a'],
            [AwsForm.NODESETTING, AwsField.NODESETTING_BASTION_HOST_ENABLED, true],
            [AwsForm.NODESETTING, AwsField.NODESETTING_CONTROL_PLANE_SETTING, 'dev'],
            [AwsForm.NODESETTING, AwsField.NODESETTING_INSTANCE_TYPE_DEV, 't3.medium'],
            [AwsForm.NODESETTING, AwsField.NODESETTING_SSH_KEY_NAME, 'default'],
            [AwsForm.NODESETTING, NodeSettingField.WORKER_NODE_INSTANCE_TYPE, 't3.small'],
            [AwsForm.NODESETTING, AwsField.NODESETTING_CREATE_CLOUD_FORMATION, true],
            [AwsForm.NODESETTING, AwsField.NODESETTING_MACHINE_HEALTH_CHECKS_ENABLED, true],
            [WizardForm.METADATA, MetadataField.CLUSTER_DESCRIPTION, 'DescriptionEXAMPLE'],
            [WizardForm.METADATA, MetadataField.CLUSTER_LOCATION, 'mylocation1'],
            [WizardForm.NETWORK, NetworkField.CLUSTER_POD_CIDR, '100.96.0.0/11'],
            [WizardForm.NETWORK, NetworkField.CLUSTER_SERVICE_CIDR, '100.64.0.0/13'],
            [WizardForm.NETWORK, NetworkField.CNI_TYPE, 'antrea'],
            [WizardForm.CEIP, CeipField.OPTIN, true]
        ];
        mappings.forEach(mapping => {
            const formName = mapping[0] as string;
            const fieldName = mapping[1] as string;
            const desiredValue = mapping[2];

            const formGroup = component.form.get(formName) as FormGroup;
            expect(formGroup).toBeTruthy();
            formGroup.addControl(fieldName, new FormControl(desiredValue));
        });
        // NOTE: because cluster labels are pulled from storage (not a DOM control) we have to put the test values in storage
        const clusterLabels = new Map<string, string>([['key1', 'value1']]);
        const identifier = { wizard: component.wizardName, step: WizardForm.METADATA, field: 'clusterLabels'};
        AppServices.userDataService.storeMap(identifier, clusterLabels);

        const payload = component.getPayload();
        expect(payload.awsAccountParams).toEqual({
            region: 'US-WEST',
            accessKeyID: 'aws-access-key-id-12345',
            secretAccessKey: 'My-AWS-Secret-Access-Key',
            profileName: '',
            sessionToken: ''
        });

        expect(payload.networking).toEqual({
            networkName: '',
            clusterDNSName: '',
            clusterNodeCIDR: '',
            clusterServiceCIDR: '100.64.0.0/13',
            clusterPodCIDR: '100.96.0.0/11',
            cniType: 'antrea'
        });
        expect(payload.labels).toEqual({
            key1: 'value1'
        });
        expect(payload.annotations).toEqual({
            description: 'DescriptionEXAMPLE',
            location: 'mylocation1'
        });
        expect(payload.createCloudFormationStack).toBe(true);
        expect(payload.controlPlaneNodeType).toBe('t3.medium');
        expect(payload.sshKeyName).toBe('default');
        expect(payload.controlPlaneFlavor).toBe('dev');
        expect(payload.vpc.azs[0].workerNodeType).toBe('t3.small');
        expect(payload.bastionHostEnabled).toBe(true);
        expect(payload.machineHealthCheckEnabled).toBe(true);
        expect(payload.ceipOptIn).toBe(true);
    });

    it('should generate cli', () => {
        const path = '/testPath/xyz.yaml';
        const payload = component.getPayload();
        if (payload.createCloudFormationStack) {
            expect(component.getCli(path)).toBe(`tanzu management-cluster permissions aws set && tanzu management-cluster create --file ${path} -v 6`);
        } else {
            expect(component.getCli(path)).toBe(`tanzu management-cluster create --file ${path} -v 6`);
        }
    });

    it('should call api to create aws regional cluster', () => {
        const apiSpy = spyOn(component['apiClient'], 'createAWSRegionalCluster').and.callThrough();
        component.createRegionalCluster({});
        expect(apiSpy).toHaveBeenCalled();
    });

    it('should apply TKG config for aws', () => {
        const apiSpy = spyOn(component['apiClient'], 'applyTKGConfigForAWS').and.callThrough();
        component.applyTkgConfig();
        expect(apiSpy).toHaveBeenCalled();
    });
});
