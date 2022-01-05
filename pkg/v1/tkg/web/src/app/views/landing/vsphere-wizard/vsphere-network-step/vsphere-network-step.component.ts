// Angular imports
import { Component } from '@angular/core';
import { Validators } from '@angular/forms';
// Third party imports
import { takeUntil } from 'rxjs/operators';
// App imports
import AppServices from '../../../../shared/service/appServices';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { NetworkField } from '../../wizard/shared/components/steps/network-step/network-step.fieldmapping';
import { SharedNetworkStepComponent } from '../../wizard/shared/components/steps/network-step/network-step.component';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { TanzuEventType } from '../../../../shared/service/Messenger';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { VSphereNetwork } from '../../../../swagger/models';
import { VsphereNetworkFieldMappings } from './vsphere-network-step.fieldmapping';
import { VsphereField } from '../vsphere-wizard.constants';

declare var sortPaths: any;
@Component({
    selector: 'app-vsphere-network-step',
    templateUrl: '../../wizard/shared/components/steps/network-step/network-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/network-step/network-step.component.scss'],
})
export class VsphereNetworkStepComponent extends SharedNetworkStepComponent {
    static description = 'Specify how Tanzu Kubernetes Grid networking is provided and any global network settings';

    vmNetworks: VSphereNetwork[] = [];

    constructor(protected validationService: ValidationService) {
        super(validationService);
    }

    listenToEvents() {
        super.listenToEvents();
        AppServices.messenger.getSubject(TanzuEventType.VSPHERE_VC_AUTHENTICATED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data) => {
                this.infraServiceAddress = data.payload;
            });
        AppServices.messenger.getSubject(TanzuEventType.VSPHERE_DATACENTER_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                this.clearControlValue(VsphereField.NETWORK_NAME);
            });
    }

    protected subscribeToServices() {
        AppServices.dataServiceRegistrar.stepSubscribe<VSphereNetwork>(this, TanzuEventType.VSPHERE_GET_VM_NETWORKS,
            this.onFetchedVmNetworks.bind(this));
    }

    private onFetchedVmNetworks(networks: Array<VSphereNetwork>) {
        this.vmNetworks = sortPaths(networks, function (item) { return item.name; }, '/');
        this.loadingNetworks = false;
        this.resurrectField(VsphereField.NETWORK_NAME,
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
            type: TanzuEventType.VSPHERE_GET_VM_NETWORKS
        });
    }

    initFormWithSavedData() {
        super.initFormWithSavedData();
        const fieldNetworkName = this.formGroup.get(VsphereField.NETWORK_NAME);
        if (fieldNetworkName) {
            const savedNetworkName = this.getSavedValue(VsphereField.NETWORK_NAME, '');
            fieldNetworkName.setValue(
                this.vmNetworks.length === 1 ? this.vmNetworks[0].name : savedNetworkName,
                { onlySelf: true } // avoid step error message when networkName is empty
            );
        }
    }

    protected supplyFieldsAffectingStepDescription(): string[] {
        const fields = super.supplyFieldsAffectingStepDescription();
        fields.push(VsphereField.NETWORK_NAME);
        return fields;
    }

    protected supplyEnablesNetworkName(): boolean {
        return true;
    }

    protected supplyNetworkNameInstruction(): string {
        return 'Select a vSphere network to use as the Kubernetes service network.';
    }

    protected supplyEnablesNoProxyWarning(): boolean {
        return true;
    }

    protected supplyNetworks(): { displayName?: string }[] {
        return this.vmNetworks;
    }

    protected supplyStepMapping(): StepMapping {
        const fieldMappings = [...VsphereNetworkFieldMappings, ...super.supplyStepMapping().fieldMappings];
        return { fieldMappings };
    }

    dynamicDescription(): string {
        // NOTE: even though this is a common wizard form, vSphere has a different way of describing it
        // because vSphere allows for the user to select a network name
        const networkName = this.getFieldValue(VsphereField.NETWORK_NAME);
        let result = '';
        if (networkName) {
            result = 'Network: ' + networkName + ' ';
        }
        const serviceCidr = this.getFieldValue(NetworkField.CLUSTER_SERVICE_CIDR, true);
        if (serviceCidr) {
            result +=  'Cluster Service CIDR: ' + serviceCidr + ' ';
        }
        const podCidr = this.getFieldValue(NetworkField.CLUSTER_POD_CIDR, true);
        if (podCidr) {
            result +=  'Cluster Pod CIDR: ' + podCidr;
        }
        return result ? result.trim() : VsphereNetworkStepComponent.description;
    }
}
