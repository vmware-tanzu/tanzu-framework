import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { FormBuilder } from '@angular/forms';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { RouterTestingModule } from '@angular/router/testing';
import { APIClient } from '../../../swagger/api-client.service';
import { AwsWizardComponent } from './aws-wizard.component';
import { SharedModule } from '../../../shared/shared.module';
import Broker from 'src/app/shared/service/broker';
import { Messenger } from 'src/app/shared/service/Messenger';
import { ClusterType, WizardForm } from "../wizard/shared/constants/wizard.constants";
import { AwsForm } from './aws-wizard.constants';

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
                FormBuilder
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            declarations: [ AwsWizardComponent ]
        })
        .compileComponents();
    }));

    beforeEach(() => {
        Broker.messenger = new Messenger();
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
                bastionHostEnabled: [''],
                controlPlaneSetting: [''],
                devInstanceType: [''],
                machineHealthChecksEnabled: [false],
                createCloudFormation: [false],
                workerNodeInstanceType1: [''],
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

    describe('should return correct description', () => {
        it('is for provider step', () => {
            const description = component.describeStep(component.AwsProviderForm.name, component.AwsProviderForm.description);
            expect(description).toBe('Validate the AWS provider account for Tanzu');
        });

        it('is for vpc step', () => {
            let description = component.describeStep(component.AwsVpcForm.name, component.AwsVpcForm.description);
            expect(description).toBe('Specify VPC settings for AWS');

            component.form.get('vpcForm').get('vpc').setValue('10.0.0.0/16');
            component.form.get('vpcForm').get('publicNodeCidr').setValue('1.1.1.1/23');
            component.form.get('vpcForm').get('privateNodeCidr').setValue('2.2.2.2/23');
            component.form.get('vpcForm').get('awsNodeAz').setValue('awsNodeAz1');
            description = component.describeStep(component.AwsVpcForm.name, component.AwsVpcForm.description);
            expect(description).toBe('VPC CIDR: 10.0.0.0/16, Public Node CIDR: 1.1.1.1/23, ' +
                'Private Node CIDR: 2.2.2.2/23, Node AZ: awsNodeAz1');
        });

        it('is for nodeSetting step', () => {
            const controlPlaneField = component.form.get(component.AwsNodeSettingForm.name).get('controlPlaneSetting');
            let description = component.describeStep(component.AwsNodeSettingForm.name, component.AwsNodeSettingForm.description);
            expect(description).toBe('Specify the resources backing the management cluster');

            controlPlaneField.setValue('prod');
            description = component.describeStep(component.AwsNodeSettingForm.name, component.AwsNodeSettingForm.description);
            expect(description).toBe('Production cluster selected: 3 node control plane');

            controlPlaneField.setValue('dev');
            description = component.describeStep(component.AwsNodeSettingForm.name, component.AwsNodeSettingForm.description);
            expect(description).toBe('Development cluster selected: 1 node control plane');
        });

        it('is for network step', () => {
            let description = component.describeStep(component.NetworkForm.name, component.NetworkForm.description);
            expect(description).toBe('Specify the cluster Pod CIDR');
            component.form.get('networkForm').get('clusterPodCidr').setValue('10.10.10.10/23');
            description = component.describeStep(component.NetworkForm.name, component.NetworkForm.description);
            expect(description).toBe('Cluster Pod CIDR: 10.10.10.10/23');
        });

        it('is for metadata step', () => {
            let description = component.describeStep(component.MetadataForm.name, component.MetadataForm.description);
            expect(description).toBe('Specify metadata for the management cluster');
            component.form.get(WizardForm.METADATA).get('clusterLocation').setValue('us-west');
            description = component.describeStep(component.MetadataForm.name, component.MetadataForm.description);
            expect(description).toBe('Location: us-west');
        });
    });

    it('should return management cluster name', () => {
        component.form.get(AwsForm.NODESETTING).get('clusterName').setValue('mylocalTestName');
        expect(component.getMCName()).toBe('mylocalTestName');
    });

    it('should create API payload', () => {
        const clusterLabels = new Map();
        clusterLabels.set('key1', 'value1');
        const mappings = [
            ['awsProviderForm', 'accessKeyID', 'aws-access-key-id-12345'],
            ['awsProviderForm', 'region', 'US-WEST'],
            ['awsProviderForm', 'secretAccessKey', 'My-AWS-Secret-Access-Key'],
            ['vpcForm', 'vpc', '10.0.0.0/16'],
            ['vpcForm', 'vpcType', 'new'],
            [AwsForm.NODESETTING, 'awsNodeAz1', 'us-west-a'],
            [AwsForm.NODESETTING, 'bastionHostEnabled', 'yes'],
            [AwsForm.NODESETTING, 'controlPlaneSetting', 'dev'],
            [AwsForm.NODESETTING, 'devInstanceType', 't3.medium'],
            [AwsForm.NODESETTING, 'sshKeyName', 'default'],
            // [AwsForm.NODESETTING, 'machineHealthChecksEnabled', true],
            [AwsForm.NODESETTING, 'workerNodeInstanceType1', 't3.small'],
            [WizardForm.METADATA, 'clusterDescription', 'DescriptionEXAMPLE'],
            // [WizardForm.METADATA, 'clusterLabels', clusterLabels],
            [WizardForm.METADATA, 'clusterLocation', 'mylocation1'],
            ['networkForm', 'clusterPodCidr', '100.96.0.0/11'],
            ['networkForm', 'clusterServiceCidr', '100.64.0.0/13'],
            ['networkForm', 'cniType', 'antrea'],
            // ['ceipOptInForm', 'ceipOptIn', true]
        ];
        mappings.forEach(attr => component.form.get(attr[0]).get(attr[1]).setValue(attr[2]));

        component.form.get(AwsForm.NODESETTING).get('createCloudFormation').setValue(true);
        component.form.get(AwsForm.NODESETTING).get('machineHealthChecksEnabled').setValue(true);
        component.form.get('ceipOptInForm').get('ceipOptIn').setValue(true);
        component.form.get(WizardForm.METADATA).get('clusterLabels').setValue(clusterLabels);

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
