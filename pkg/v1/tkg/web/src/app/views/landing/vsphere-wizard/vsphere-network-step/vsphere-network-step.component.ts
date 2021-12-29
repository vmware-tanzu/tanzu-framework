// Angular imports
import { Component } from '@angular/core';
import { Validators } from '@angular/forms';
// Third party imports
import { takeUntil } from 'rxjs/operators';
// App imports
import AppServices from '../../../../shared/service/appServices';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { SharedNetworkStepComponent } from '../../wizard/shared/components/steps/network-step/network-step.component';
import { TkgEventType } from '../../../../shared/service/Messenger';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { VSphereNetwork } from '../../../../swagger/models';

declare var sortPaths: any;
@Component({
    selector: 'app-vsphere-network-step',
    templateUrl: '../../wizard/shared/components/steps/network-step/network-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/network-step/network-step.component.scss'],
})
export class VsphereNetworkStepComponent extends SharedNetworkStepComponent {
    static description = 'Specify how Tanzu Kubernetes Grid networking is provided and any global network settings';

    constructor(protected validationService: ValidationService,
                protected fieldMapUtilities: FieldMapUtilities) {
        super(validationService, fieldMapUtilities);
        this.enableNetworkName = true;
    }

    listenToEvents() {
        super.listenToEvents();
        AppServices.messenger.getSubject(TkgEventType.VSPHERE_VC_AUTHENTICATED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data) => {
                this.infraServiceAddress = data.payload;
            });
        AppServices.messenger.getSubject(TkgEventType.VSPHERE_DATACENTER_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                this.clearControlValue('networkName');
            });
    }

    protected subscribeToServices() {
        AppServices.dataServiceRegistrar.stepSubscribe<VSphereNetwork>(this, TkgEventType.VSPHERE_GET_VM_NETWORKS,
            this.onFetchedVmNetworks.bind(this));
    }

    private onFetchedVmNetworks(networks: Array<VSphereNetwork>) {
        this.vmNetworks = sortPaths(networks, function (item) { return item.name; }, '/');
        this.loadingNetworks = false;
        this.resurrectField('networkName',
            [Validators.required], networks.length === 1 ? networks[0].name : '',
            { onlySelf: true } // only for current form control
        );
    }

    protected onNoProxyChange(value: string) {
        this.hideNoProxyWarning = value.trim().split(',').includes(this.infraServiceAddress);
        super.onNoProxyChange(value);
    }

    /**
     * @method loadNetworks
     * helper method retrieves list of networks
     */
    loadNetworks() {
        this.loadingNetworks = true;
        AppServices.messenger.publish({
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
