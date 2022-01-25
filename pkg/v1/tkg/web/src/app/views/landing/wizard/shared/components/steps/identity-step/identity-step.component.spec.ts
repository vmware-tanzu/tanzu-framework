// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';

// Library imports
import { APIClient } from 'tanzu-mgmt-plugin-api-lib';

// App imports
import AppServices from 'src/app/shared/service/appServices';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { Messenger, TkgEventType } from 'src/app/shared/service/Messenger';
import { SharedIdentityStepComponent } from './identity-step.component';
import { SharedModule } from '../../../../../../../shared/shared.module';
import { ValidationService } from '../../../validation/validation.service';
import { IdentityManagementType, WizardForm } from '../../../constants/wizard.constants';

describe('IdentityStepComponent', () => {
  let component: SharedIdentityStepComponent;
  let fixture: ComponentFixture<SharedIdentityStepComponent>;

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
            declarations: [SharedIdentityStepComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        AppServices.messenger = new Messenger();
        fixture = TestBed.createComponent(SharedIdentityStepComponent);
        component = fixture.componentInstance;
        component.setInputs('BozoWizard', WizardForm.IDENTITY, new FormBuilder().group({}));

        fixture.detectChanges();
    });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should switch to ldap', () => {
    fixture.whenStable().then(() => {
      spyOn(component, 'unsetAllValidators').and.callThrough();
      spyOn(component, 'setLDAPValidators').and.callThrough();
      component.formGroup.get('identityType').setValue('ldap');
      expect(component.identityTypeValue).toEqual('ldap');
      expect(component.unsetAllValidators).toHaveBeenCalled();
      expect(component.setLDAPValidators).toHaveBeenCalled();
    });
  });

  it('should switch back to oidc', () => {
    fixture.whenStable().then(() => {
      component.formGroup.get('identityType').setValue('ldap');
      spyOn(component, 'unsetAllValidators').and.callThrough();
      spyOn(component, 'setOIDCValidators').and.callThrough();
      component.formGroup.get('identityType').setValue('oidc');
      expect(component.identityTypeValue).toEqual('oidc');
      expect(component.unsetAllValidators).toHaveBeenCalled();
      expect(component.setOIDCValidators).toHaveBeenCalled();
    });
  });

    it('should announce description change', () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        component.ngOnInit();
        const identityTypeControl = component.formGroup.get('identityType');
        const oidcIssuerControl = component.formGroup.get('issuerURL');
        const ldapEndpointIpControl = component.formGroup.get('endpointIp');
        const ldapEndpointPortControl = component.formGroup.get('endpointPort');

        expect(component.dynamicDescription()).toEqual(SharedIdentityStepComponent.description);

        // OIDC
        identityTypeControl.setValue(IdentityManagementType.OIDC);
        oidcIssuerControl.setValue('https://1.2.3.4');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TkgEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: WizardForm.IDENTITY,
                description: 'OIDC configured: https://1.2.3.4',
            }
        });

        // LDAP without port set
        identityTypeControl.setValue(IdentityManagementType.LDAP);
        ldapEndpointIpControl.setValue('https://5.6.7.8');
        ldapEndpointPortControl.setValue('');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TkgEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: WizardForm.IDENTITY,
                description: 'LDAP configured: https://5.6.7.8:',
            }
        });

        // LDAP with port set
        identityTypeControl.setValue(IdentityManagementType.LDAP);
        ldapEndpointIpControl.setValue('https://9.8.7.6');
        ldapEndpointPortControl.setValue('123');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TkgEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: WizardForm.IDENTITY,
                description: 'LDAP configured: https://9.8.7.6:123',
            }
        });
    });
});
