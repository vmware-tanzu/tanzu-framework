// Angular modules
import { TestBed, async, ComponentFixture } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';

// Library imports
import { APIClient } from 'tanzu-management-cluster-ng-api';

// App imports
import { ConfirmComponent } from './confirm.component';
import { SharedModule } from './../../../shared/shared.module';
import { VmwCopyToClipboardButtonComponent } from '../../../shared/components/copy-to-clipboard-button/copy-to-clipboard-button.component';
import { PreviewConfigComponent } from '../../../shared/components/preview-config/preview-config.component';
import { FormMetaDataStore } from '../wizard/shared/FormMetaDataStore';

describe('ConfirmComponent', () => {
    let component: ConfirmComponent;
    let fixture: ComponentFixture<ConfirmComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule,
                HttpClientTestingModule,
                SharedModule
            ],
            providers: [
                APIClient
            ],
            declarations: [
                ConfirmComponent,
                PreviewConfigComponent,
                VmwCopyToClipboardButtonComponent
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ConfirmComponent);
        component = fixture.componentInstance;
    });

    it('should exist', () => {
        const landingComponent = fixture.debugElement.componentInstance;
        expect(landingComponent).toBeTruthy();
    });

    it('should retrieve data from meta data store', () => {
        const stepListSpy = spyOn(FormMetaDataStore, 'getStepList').and.returnValues([{title: 'test', description: 'testDescription'}]);
        const formListSpy = spyOn(FormMetaDataStore, 'getFormList').and.returnValues(['provider-step']);
        const metaDataSpy = spyOn(FormMetaDataStore, 'getMetaData').and.callThrough();
        component.ngOnInit();
        expect(stepListSpy).toHaveBeenCalled();
        expect(formListSpy).toHaveBeenCalled();
        expect(metaDataSpy).toHaveBeenCalled();
    });

    it('should return all entries of data', () => {
        const data = {
            key1: 'value1',
            key2: 'value2'
        };
        expect(component.entries(data)).toEqual(['value1', 'value2']);
        expect(component.entries(null)).toBeNull();
    });
});
