import { Component } from '@angular/core';
import { SharedNetworkStepComponent } from '../../wizard/shared/components/steps/network-step/network-step.component';
import Broker from '../../../../shared/service/broker';
import { TkgEventType } from '../../../../shared/service/Messenger';
import { distinctUntilChanged, takeUntil } from 'rxjs/operators';
import { VSphereNetwork } from '../../../../swagger/models';
import { Validators } from '@angular/forms';
import { VSphereWizardFormService } from '../../../../shared/service/vsphere-wizard-form.service';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';

declare var sortPaths: any;
@Component({
    selector: 'app-vsphere-network-step',
    templateUrl: '../../wizard/shared/components/steps/network-step/network-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/network-step/network-step.component.scss'],
})
export class VsphereNetworkStepComponent extends SharedNetworkStepComponent {
    constructor(private wizardFormService: VSphereWizardFormService,
                protected validationService: ValidationService,
                protected fieldMapUtilities: FieldMapUtilities) {
        super(validationService, fieldMapUtilities);
        this.enableNetworkName = true;
    }

    listenToEvents() {
        super.listenToEvents();
        Broker.messenger.getSubject(TkgEventType.VSPHERE_VC_AUTHENTICATED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data) => {
                this.infraServiceAddress = data.payload;
            });
        Broker.messenger.getSubject(TkgEventType.VSPHERE_DATACENTER_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                this.clearControlValue('networkName');
            });
        this.wizardFormService.getErrorStream(TkgEventType.VSPHERE_GET_VM_NETWORKS)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(error => {
                this.errorNotification = error;
            });
        this.wizardFormService.getDataStream(TkgEventType.VSPHERE_GET_VM_NETWORKS)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((networks: Array<VSphereNetwork>) => {
                this.vmNetworks = sortPaths(networks, function (item) { return item.name; }, '/');
                this.loadingNetworks = false;
                this.resurrectField('networkName',
                    [Validators.required], networks.length === 1 ? networks[0].name : '',
                    { onlySelf: true } // only for current form control
                );
            });
    }

    protected onNoProxyChange(value: string) {
        this.hideNoProxyWarning = value.trim().split(',').includes(this.infraServiceAddress);
        super.onNoProxyChange(value);
    }

    /**
     * @method loadVSphereNetworks
     * helper method retrieves list of vsphere networks
     */
    loadNetworks() {
        this.loadingNetworks = true;
        Broker.messenger.publish({
            type: TkgEventType.VSPHERE_GET_VM_NETWORKS
        });
    }

    initFormWithSavedData() {
        super.initFormWithSavedData();
        const fieldNetworkName = this.formGroup.get('networkName');
        if (fieldNetworkName) {
            const savedNetworkName = this.getSavedValue('networkName', '');
            fieldNetworkName.setValue(
                this.vmNetworks.length === 1 ? this.vmNetworks[0].name : savedNetworkName,
                { onlySelf: true } // avoid step error message when networkName is empty
            );
        }
    }
}
