// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { By } from '@angular/platform-browser';

// Library imports
import { APIClient, AWSSubnet } from 'tanzu-management-cluster-api';

// App imports
import AppServices from 'src/app/shared/service/appServices';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { NodeSettingStepComponent, NodeType } from './node-setting-step.component';
import { Messenger, TkgEventType } from 'src/app/shared/service/Messenger';
import { SharedModule } from '../../../../shared/shared.module';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { DataServiceRegistrarTestExtension } from '../../../../testing/data-service-registrar.testextension';
import { AwsField, AwsForm } from '../aws-wizard.constants';

describe('NodeSettingStepComponent', () => {
    let component: NodeSettingStepComponent;
    let fixture: ComponentFixture<NodeSettingStepComponent>;
    const vpcSubnets = ['vpcPublicSubnet1', 'vpcPrivateSubnet1', 'vpcPublicSubnet2',
        'vpcPrivateSubnet2', 'vpcPublicSubnet3', 'vpcPrivateSubnet3'];
    const azs = ['awsNodeAz1', 'awsNodeAz2', 'awsNodeAz3'];

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                ReactiveFormsModule,
                SharedModule
            ],
            providers: [
                ValidationService,
                FormBuilder,
                FieldMapUtilities,
                APIClient
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            declarations: [NodeSettingStepComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        AppServices.messenger = new Messenger();
        AppServices.dataServiceRegistrar = new DataServiceRegistrarTestExtension();
        fixture = TestBed.createComponent(NodeSettingStepComponent);
        component = fixture.componentInstance;
        component.setInputs('SquashWizard', AwsForm.NODESETTING, new FormBuilder().group({}));

        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    afterEach(() => {
        fixture.destroy();
    });

    it('should be invalid when cluster name has leading/trailing spaces', () => {
        fixture.whenStable().then(() => {
            component.formGroup.get('clusterName').setValue(" test");
            expect(component.formGroup.valid).toBeFalsy();
            component.formGroup.get('clusterName').setValue("test   ");
            expect(component.formGroup.valid).toBeFalsy();
        });
    });

    xit('Dev instance type should be reset if control plane is prod', () => {
        fixture.whenStable().then(() => {
            component.formGroup.get("devInstanceType").setValue("t3.small");
            const cards = fixture.debugElement.queryAll(By.css("a.card"));
            cards[1].triggerEventHandler('click', {});
            expect(component.formGroup.get("devInstanceType").value).toBeFalsy();
        });
    });

    xit('Prod instance type should be reset if control plane is dev', () => {
        fixture.whenStable().then(() => {
            component.formGroup.get("prodInstanceType").setValue("t3.small");
            const cards = fixture.debugElement.queryAll(By.css("a.card"));
            cards[0].triggerEventHandler('click', {});
            expect(component.formGroup.get("prodInstanceType").value).toBeFalsy();
        });
    });

    it('should return dev instance type', () => {
        component.formGroup.get("devInstanceType").setValue('t3.small');
        expect(component.devInstanceTypeValue).toBe('t3.small');
    });

    it('should return pro instance type', () => {
        component.formGroup.get("prodInstanceType").setValue('t3.small');
        expect(component.prodInstanceTypeValue).toBe('t3.small');
    });

    it('should return worker node instance type', () => {
        component.formGroup.get("workerNodeInstanceType1").setValue('t3.small');
        expect(component.formGroup.get('workerNodeInstanceType1').value).toBe('t3.small');
    });

    it('should return environment type', () => {
        component.formGroup.get("controlPlaneSetting").setValue('dev');
        expect(component.getEnvType()).toBe('dev');
    });

    it('should clear availability zone', () => {
        component.formGroup.get('awsNodeAz1').setValue('us-west-a');
        component.formGroup.get('awsNodeAz2').setValue('us-west-b');
        component.formGroup.get('awsNodeAz3').setValue('us-west-c');
        component.clearAzs();
        azs.forEach(az => expect(component.formGroup.get(az).value).toBe(''));
    });

    it('should clear subsets', () => {
        component.formGroup.get('vpcPublicSubnet1').setValue('100.63.0.0/14');
        component.formGroup.get('vpcPrivateSubnet1').setValue('100.54.0.0/14');
        component.formGroup.get('vpcPublicSubnet2').setValue('100.63.0.0/14');
        component.formGroup.get('vpcPrivateSubnet2').setValue('100.54.0.0/14');
        component.formGroup.get('vpcPublicSubnet3').setValue('100.63.0.0/14');
        component.formGroup.get('vpcPrivateSubnet3').setValue('100.63.0.0/14');
        component.clearSubnets();
        vpcSubnets.forEach(subnet => expect(component.formGroup.get(subnet).value).toBe(''));
    });

    it('should clear subnet data', () => {
        component.filteredAzs = {
            awsNodeAz1: {
                publicSubnets: [{cidr: '100.63.0.0/14', isPublic:  true}],
                privateSubnets: [{cidr: '100.63.0.0/14', isPublic: false}]
            },
            awsNodeAz2: {
                publicSubnets: [{cidr: '100.63.0.0/14', isPublic:  true}],
                privateSubnets: [{cidr: '100.63.0.0/14', isPublic: false}]
            },
            awsNodeAz3: {
                publicSubnets: [{cidr: '100.63.0.0/14', isPublic:  true}],
                privateSubnets: [{cidr: '100.63.0.0/14', isPublic: false}]
            }
        }
        component.clearSubnetData();
        expect(component.filteredAzs).toEqual({
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
        });
    });

    it('should fiter subnet', () => {
        component.vpcType = 'existing';
        component.publicSubnets = [{
            availabilityZoneId: 'us-west-a',
            availabilityZoneName: 'us-west-a',
            cidr: '100.64.0.0/13',
            id: 'sn1',
            isPublic: true,
        }];
        component.privateSubnets = [{
            availabilityZoneId: 'us-west-a',
            availabilityZoneName: 'us-west-a',
            cidr: '100.64.0.0/24',
            id: 'sn4',
            isPublic: false
        }];

        component.filterSubnetsByAZ('awsNodeAz1', 'us-west-a');
        expect(component.filteredAzs['awsNodeAz1']).toEqual({
            publicSubnets: [{
                availabilityZoneId: 'us-west-a',
                availabilityZoneName: 'us-west-a',
                cidr: '100.64.0.0/13',
                id: 'sn1',
                isPublic: true,
            }],
            privateSubnets: [{
                availabilityZoneId: 'us-west-a',
                availabilityZoneName: 'us-west-a',
                cidr: '100.64.0.0/24',
                id: 'sn4',
                isPublic: false
            }]
        })
    });

    it('should reset fields when aws region changed', () => {
        component.publicSubnets = [{
            availabilityZoneId: 'us-west-a',
            availabilityZoneName: 'us-west-a',
            cidr: '100.64.0.0/13',
            id: 'sn1',
            isPublic: true,
        }];
        component.privateSubnets = [{
            availabilityZoneId: 'us-west-a',
            availabilityZoneName: 'us-west-a',
            cidr: '100.64.0.0/24',
            id: 'sn4',
            isPublic: false
        }];
        component.filteredAzs = {
            awsNodeAz1: {
                publicSubnets: [{cidr: '100.63.0.0/14', isPublic:  true}],
                privateSubnets: [{cidr: '100.63.0.0/14', isPublic: false}]
            },
            awsNodeAz2: {
                publicSubnets: [{cidr: '100.63.0.0/14', isPublic:  true}],
                privateSubnets: [{cidr: '100.63.0.0/14', isPublic: false}]
            },
            awsNodeAz3: {
                publicSubnets: [{cidr: '100.63.0.0/14', isPublic:  true}],
                privateSubnets: [{cidr: '100.63.0.0/14', isPublic: false}]
            }
        }
        component.formGroup.get('awsNodeAz1').setValue('us-west-a');
        component.formGroup.get('awsNodeAz2').setValue('us-west-b');
        component.formGroup.get('awsNodeAz3').setValue('us-west-c');

        component.formGroup.get('vpcPublicSubnet1').setValue('100.63.0.0/14');
        component.formGroup.get('vpcPrivateSubnet1').setValue('100.54.0.0/14');
        component.formGroup.get('vpcPublicSubnet2').setValue('100.63.0.0/14');
        component.formGroup.get('vpcPrivateSubnet2').setValue('100.54.0.0/14');
        component.formGroup.get('vpcPublicSubnet3').setValue('100.63.0.0/14');
        component.formGroup.get('vpcPrivateSubnet3').setValue('100.63.0.0/14');

        AppServices.messenger.publish({ type: TkgEventType.AWS_REGION_CHANGED});
        expect(component.publicSubnets).toEqual([]);
        expect(component.privateSubnets).toEqual([]);
        expect(component.filteredAzs).toEqual({
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
        });
        azs.forEach(az => expect(component.formGroup.get(az).value).toBe(''));
        vpcSubnets.forEach(subnet => expect(component.formGroup.get(subnet).value).toBe(''));
    });

    it('should handle aws vpc type change', () => {
        component.formGroup.get('vpcPublicSubnet1').setValue('100.63.0.0/14');
        component.formGroup.get('vpcPrivateSubnet1').setValue('100.54.0.0/14');

        const spySubnets = [];
        vpcSubnets.forEach(vpcSubnet => spySubnets.push(spyOn(component.formGroup.get(vpcSubnet), 'setValidators').and.callThrough()));
        const spyAzs = spyOn(component, 'clearAzs').and.callThrough();

        AppServices.messenger.publish({ type: TkgEventType.AWS_VPC_TYPE_CHANGED, payload: { vpcType: 'existing'}});

        spySubnets.forEach(subnet => expect(subnet).toHaveBeenCalledTimes(1));
        expect(spyAzs).toHaveBeenCalled();
    });

    it('should handle aws vpc change', () => {
        const spyAzs = spyOn(component, 'clearAzs').and.callThrough();
        const spySubnets = spyOn(component, 'clearSubnets').and.callThrough();
        AppServices.messenger.publish({ type: TkgEventType.AWS_VPC_CHANGED});
        expect(spyAzs).toHaveBeenCalled();
        expect(spySubnets).toHaveBeenCalled();
    });

    it('should handle AWS_GET_SUBNETS event', () => {
        const dataServiceRegistrar = AppServices.dataServiceRegistrar as DataServiceRegistrarTestExtension;
        // we expect wizard to have registered this event
        dataServiceRegistrar.simulateRegistration<AWSSubnet>(TkgEventType.AWS_GET_SUBNETS);

        component.ngOnInit();

        const spySavedSubnet = spyOn(component, 'setSubnetFieldsFromSavedValues').and.callThrough();
        dataServiceRegistrar.simulateData(TkgEventType.AWS_GET_SUBNETS, [
            {cidr: '100.63.0.0/14', isPublic:  true},
            {cidr: '100.64.0.0/14', isPublic:  false}
        ]);
        expect(component.publicSubnets).toEqual([{cidr: '100.63.0.0/14', isPublic:  true}]);
        expect(component.privateSubnets).toEqual([{cidr: '100.64.0.0/14', isPublic:  false}]);
        expect(spySavedSubnet).toHaveBeenCalled();
    });

    it('should announce description change', () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();

        component.ngOnInit();
        component.nodeType = '';
        const description = component.dynamicDescription();
        expect(description).toEqual('Specify the resources backing the  cluster');

        component.setClusterTypeDescriptor('CARAMEL');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TkgEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'SquashWizard',
                step: AwsForm.NODESETTING,
                description: 'Specify the resources backing the CARAMEL cluster',
            }
        });

        const controlPlaneSettingControl = component.formGroup.controls[AwsField.NODESETTING_CONTROL_PLANE_SETTING];
        controlPlaneSettingControl.setValue(NodeType.DEV);
        expect(msgSpy).toHaveBeenCalledWith({
            type: TkgEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'SquashWizard',
                step: AwsForm.NODESETTING,
                description: 'Development cluster selected: 1 node control plane',
            }
        });

        controlPlaneSettingControl.setValue(NodeType.PROD);
        expect(msgSpy).toHaveBeenCalledWith({
            type: TkgEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'SquashWizard',
                step: AwsForm.NODESETTING,
                description: 'Production cluster selected: 3 node control plane',
            }
        });
    });
});
