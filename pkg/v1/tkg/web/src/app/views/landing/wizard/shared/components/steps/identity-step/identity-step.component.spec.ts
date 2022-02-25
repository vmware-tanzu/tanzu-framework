// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
// App imports
import { APIClient } from '../../../../../../../swagger/api-client.service';
import AppServices from 'src/app/shared/service/appServices';
import { IdentityField } from './identity-step.fieldmapping';
import { IdentityManagementType, WizardForm } from '../../../constants/wizard.constants';
import { Messenger, TanzuEventType } from 'src/app/shared/service/Messenger';
import { SharedIdentityStepComponent } from './identity-step.component';
import { SharedModule } from '../../../../../../../shared/shared.module';
import { ValidationService } from '../../../validation/validation.service';

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
        // NOTE: using Azure file import stuff just to test
        component.setStepRegistrantData({ wizard: 'BozoWizard', step: WizardForm.IDENTITY, formGroup: new FormBuilder().group({}),
            eventFileImported: TanzuEventType.AZURE_CONFIG_FILE_IMPORTED,
            eventFileImportError: TanzuEventType.AZURE_CONFIG_FILE_IMPORT_ERROR});
        fixture.detectChanges();
    });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should switch to ldap', () => {
    fixture.whenStable().then(() => {
      spyOn(component, 'unsetValidators').and.callThrough();
      spyOn(component, 'setLDAPValidators').and.callThrough();
      component.formGroup.get(IdentityField.IDENTITY_TYPE).setValue(IdentityManagementType.LDAP);
      expect(component.isIdentityManagementLdap).toBeTrue();
      expect(component.unsetValidators).toHaveBeenCalled();
      expect(component.setLDAPValidators).toHaveBeenCalled();
    });
  });

  it('should switch back to oidc', () => {
    fixture.whenStable().then(() => {
      component.formGroup.get(IdentityField.IDENTITY_TYPE).setValue(IdentityManagementType.LDAP);
      spyOn(component, 'unsetValidators').and.callThrough();
      spyOn(component, 'setOIDCValidators').and.callThrough();
      component.formGroup.get(IdentityField.IDENTITY_TYPE).setValue(IdentityManagementType.OIDC);
      expect(component.isIdentityManagementOidc).toBeTrue();
      expect(component.unsetValidators).toHaveBeenCalled();
      expect(component.setOIDCValidators).toHaveBeenCalled();
    });
  });

    it('should announce description change', () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        component.ngOnInit();
        const identityTypeControl = component.formGroup.get(IdentityField.IDENTITY_TYPE);
        const oidcIssuerControl = component.formGroup.get(IdentityField.ISSUER_URL);
        const ldapEndpointIpControl = component.formGroup.get(IdentityField.ENDPOINT_IP);
        const ldapEndpointPortControl = component.formGroup.get(IdentityField.ENDPOINT_PORT);

        expect(component.dynamicDescription()).toEqual(SharedIdentityStepComponent.description);

        // OIDC
        identityTypeControl.setValue(IdentityManagementType.OIDC);
        oidcIssuerControl.setValue('https://1.2.3.4');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
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
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
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
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: WizardForm.IDENTITY,
                description: 'LDAP configured: https://9.8.7.6:123',
            }
        });
    });
});
