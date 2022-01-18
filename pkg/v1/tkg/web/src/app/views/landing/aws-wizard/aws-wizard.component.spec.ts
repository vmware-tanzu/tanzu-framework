// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { FormBuilder, FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';

// Library imports
import { APIClient } from 'tanzu-ui-api-lib';

// App imports
import AppServices from 'src/app/shared/service/appServices';
import { AwsForm } from './aws-wizard.constants';
import { AwsWizardComponent } from './aws-wizard.component';
import { ClusterType, WizardForm } from "../wizard/shared/constants/wizard.constants";
import { FieldMapUtilities } from '../wizard/shared/field-mapping/FieldMapUtilities';
import { Messenger } from 'src/app/shared/service/Messenger';
import { MetadataStepComponent } from '../wizard/shared/components/steps/metadata-step/metadata-step.component';
import { NodeSettingStepComponent } from './node-setting-step/node-setting-step.component';
import { SharedModule } from '../../../shared/shared.module';
import { SharedNetworkStepComponent } from '../wizard/shared/components/steps/network-step/network-step.component';
import { ValidationService } from '../wizard/shared/validation/validation.service';
import { VpcStepComponent } from './vpc-step/vpc-step.component';

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
                FieldMapUtilities,
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

    it('should return management cluster name', () => {
        const stepInstance = TestBed.createComponent(NodeSettingStepComponent).componentInstance;
        component.registerStep(AwsForm.NODESETTING, stepInstance);
        stepInstance.formGroup.addControl('clusterName', new FormControl(''));

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
            [AwsForm.NODESETTING, 'workerNodeInstanceType1', 't3.small'],
            [AwsForm.NODESETTING, 'createCloudFormation', true],
            [AwsForm.NODESETTING, 'machineHealthChecksEnabled', true],
            [WizardForm.METADATA, 'clusterDescription', 'DescriptionEXAMPLE'],
            [WizardForm.METADATA, 'clusterLabels', clusterLabels],
            [WizardForm.METADATA, 'clusterLocation', 'mylocation1'],
            ['networkForm', 'clusterPodCidr', '100.96.0.0/11'],
            ['networkForm', 'clusterServiceCidr', '100.64.0.0/13'],
            ['networkForm', 'cniType', 'antrea'],
            ['ceipOptInForm', 'ceipOptIn', true]
        ];
        mappings.forEach(mapping => {
            const formName = mapping[0] as string;
            const fieldName = mapping[1] as string;
            const desiredValue = mapping[2];

            const formGroup = component.form.get(formName) as FormGroup;
            expect(formGroup).toBeTruthy();
            formGroup.addControl(fieldName, new FormControl(desiredValue));
        });

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
