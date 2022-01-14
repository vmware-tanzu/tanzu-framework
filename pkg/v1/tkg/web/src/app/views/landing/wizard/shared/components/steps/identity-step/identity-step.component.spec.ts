// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { FormBuilder } from '@angular/forms';

// Library imports
import { APIClient } from 'tanzu-ui-api-lib';

// App imports
import AppServices from 'src/app/shared/service/appServices';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { Messenger } from 'src/app/shared/service/Messenger';
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
        const fb = new FormBuilder();
        fixture = TestBed.createComponent(SharedIdentityStepComponent);
        component = fixture.componentInstance;
        component.formGroup = fb.group({
        });

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
});
