/**
 * Angular Modules
 */
import { Component, Input, OnInit } from '@angular/core';
import {
        FormGroup,
        Validators,
        FormControl
} from '@angular/forms';

/**
 * App imports
 */
import { IAAS_DEFAULT_CIDRS, IpFamilyEnum } from '../../../../../../../shared/constants/app.constants';
import { ValidationService } from '../../../validation/validation.service';
import { StepFormDirective } from '../../../step-form/step-form';
import { FormMetaDataStore, FormMetaData } from '../../../FormMetaDataStore';
import { TkgEventType } from 'src/app/shared/service/Messenger';
import { VSphereWizardFormService } from 'src/app/shared/service/vsphere-wizard-form.service';
import { distinctUntilChanged, takeUntil } from 'rxjs/operators';
import { VSphereNetwork } from 'src/app/swagger/models/v-sphere-network.model';
import Broker from 'src/app/shared/service/broker';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { NetworkIpv4StepMapping, NetworkIpv6StepMapping } from './network-step.fieldmapping';
import { managementClusterPlugin, WizardForm } from "../../../constants/wizard.constants";
import { FormUtils } from '../../../utils/form-utils';
import { StepMapping } from '../../../field-mapping/FieldMapping';

declare var sortPaths: any;
@Component({
    selector: 'app-shared-network-step',
    templateUrl: './network-step.component.html',
    styleUrls: ['./network-step.component.scss']
})
export class SharedNetworkStepComponent extends StepFormDirective implements OnInit {
    @Input() enableNetworkName: boolean;
    @Input() enableNoProxyWarning: boolean;

    form: FormGroup;
    cniType: string;
    loadingNetworks: boolean = false;
    vmNetworks: Array<VSphereNetwork>;
    additionalNoProxyInfo: string;
    fullNoProxy: string;
    infraServiceAddress: string = '';
    hideWarning: boolean = true;

    constructor(private validationService: ValidationService,
                protected fieldMapUtilities: FieldMapUtilities,
        private wizardFormService: VSphereWizardFormService) {
        super(fieldMapUtilities);
    }

    protected supplyStepMapping(): StepMapping {
        return this.ipFamily === IpFamilyEnum.IPv4 ? NetworkIpv4StepMapping : NetworkIpv6StepMapping;
    }

    protected customizeForm() {
        if (!this.enableNetworkName) {
            this.clearFieldSavedData('networkName');
            this.formGroup.removeControl('networkName');
        }

        const cidrs = ['clusterServiceCidr', 'clusterPodCidr'];
        cidrs.forEach(cidr => {
            this.registerOnIpFamilyChange(cidr, [
                this.validationService.isValidIpNetworkSegment()], [
                this.validationService.isValidIpv6NetworkSegment(),
                this.setCidrs
            ]);
        });

        this.setValidators();
    }

    ngOnInit() {
        super.ngOnInit();
        this.listenToEvents();

        const cniTypeData = {
            label: 'CNI PROVIDER',
            displayValue: this.cniType,
        } as FormMetaData;
        FormMetaDataStore.saveMetaDataEntry(this.formName, 'cniType', cniTypeData);
        // TODO: guessing we don't need this line (due to initFormWithSavedData() below)
        this.formGroup.get('cniType').setValue(this.cniType, { onlySelf: true });
        this.initFormWithSavedData();
    }

    setValidators() {
        const configuredCni = Broker.appDataService.getPluginFeature(managementClusterPlugin, 'cni');
        if (configuredCni && ['antrea', 'calico', 'none'].includes(configuredCni)) {
            this.cniType = configuredCni;
        } else {
            this.cniType = 'antrea';
        }

        if (this.cniType === 'none') {
            ['clusterServiceCidr', 'clusterPodCidr'].forEach( field => this.disarmField(field, false));
        } else {
            if (this.cniType === 'calico') {
                this.disarmField('clusterServiceCidr', false);
            }
            this.setCidrs();

            if (this.enableNetworkName) {
                this.resurrectField('networkName', [
                    Validators.required
                ], '', { onlySelf: true }); // only for current form control
            }
        }
    }
    setCidrs = () => {
        if (this.cniType === 'antrea') {
            this.resurrectField('clusterServiceCidr', [
                Validators.required,
                this.validationService.noWhitespaceOnEnds(),
                this.ipFamily === IpFamilyEnum.IPv4 ?
                    this.validationService.isValidIpNetworkSegment() : this.validationService.isValidIpv6NetworkSegment(),
                this.validationService.isIpUnique([this.formGroup.get('clusterPodCidr')])
            ], this.ipFamily === IpFamilyEnum.IPv4 ?
                IAAS_DEFAULT_CIDRS.CLUSTER_SVC_CIDR : IAAS_DEFAULT_CIDRS.CLUSTER_SVC_IPV6_CIDR, { onlySelf: true });
        }

        this.resurrectField('clusterPodCidr', [
            Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.ipFamily === IpFamilyEnum.IPv4 ?
                this.validationService.isValidIpNetworkSegment() : this.validationService.isValidIpv6NetworkSegment(),
            this.validationService.isIpUnique([this.formGroup.get('clusterServiceCidr')])
        ], this.ipFamily === IpFamilyEnum.IPv4 ?
            IAAS_DEFAULT_CIDRS.CLUSTER_POD_CIDR : IAAS_DEFAULT_CIDRS.CLUSTER_POD_IPV6_CIDR, { onlySelf: true });
    }
    listenToEvents() {
        /**
         * Whenever data center selection changes, reset the relevant fields
        */
        Broker.messenger.getSubject(TkgEventType.DATACENTER_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                this.resetFieldsUponDCChange();
            });
        if (this.enableNoProxyWarning) {
            Broker.messenger.getSubject(TkgEventType.VC_AUTHENTICATED)
                .pipe(takeUntil(this.unsubscribe))
                .subscribe((data) => {
                    this.infraServiceAddress = data.payload;
                });
        }

        this.wizardFormService.getErrorStream(TkgEventType.GET_VM_NETWORKS)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(error => {
                this.errorNotification = error;
            });
        this.wizardFormService.getDataStream(TkgEventType.GET_VM_NETWORKS)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((networks: Array<VSphereNetwork>) => {
                this.vmNetworks = sortPaths(networks, function (item) { return item.name; }, '/');
                this.loadingNetworks = false;
                this.resurrectField('networkName',
                    [Validators.required], networks.length === 1 ? networks[0].name : '',
                    { onlySelf: true } // only for current form control
                );
            });

        const noProxyFieldChangeMap = ['noProxy', 'clusterServiceCidr', 'clusterPodCidr'];

        noProxyFieldChangeMap.forEach((field) => {
            this.formGroup.get(field).valueChanges.pipe(
                distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
                takeUntil(this.unsubscribe)
            ).subscribe((value) => {
                if (this.enableNoProxyWarning && field === 'noProxy') {
                    this.hideWarning = value.trim().split(',').includes(this.infraServiceAddress);
                }
                this.generateFullNoProxy();
            });
        });

        Broker.messenger.getSubject(TkgEventType.NETWORK_STEP_GET_NO_PROXY_INFO)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
               this.additionalNoProxyInfo = event.payload.info;
               this.generateFullNoProxy();
            });
    }

    generateFullNoProxy() {
        const noProxy = this.formGroup.get('noProxy');
        if (noProxy && !noProxy.value) {
            this.fullNoProxy = '';
            return;
        }
        const clusterServiceCidr = this.formGroup.get('clusterServiceCidr');
        const clusterPodCidr = this.formGroup.get('clusterPodCidr');

        const noProxyList = [
            ...noProxy.value.split(','),
            this.additionalNoProxyInfo,
            clusterServiceCidr && clusterServiceCidr.value,
            clusterPodCidr && clusterPodCidr.value,
            'localhost',
            '127.0.0.1',
            '.svc',
            '.svc.cluster.local'
        ];

        this.fullNoProxy = noProxyList.filter(elem => elem).join(',');
    }

    toggleProxySetting(fromSavedData?: boolean) {
        const proxySettingFields = [
            'httpProxyUrl',
            'httpProxyUsername',
            'httpProxyPassword',
            'isSameAsHttp',
            'httpsProxyUrl',
            'httpsProxyUsername',
            'httpsProxyPassword',
            'noProxy'
        ];

        if (!fromSavedData) {
            this.formGroup.markAsPending();
        }

        if (this.formGroup.value['proxySettings']) {
            this.resurrectField('httpProxyUrl', [
                Validators.required,
                this.validationService.isHttpOrHttps()
            ], this.formGroup.value['httpProxyUrl'],
                { onlySelf: true }
            );
            this.resurrectField('noProxy', [],
                this.formGroup.value['noProxy'] || this.infraServiceAddress,
                { onlySelf: true }
            );
            if (!this.formGroup.value['isSameAsHttp']) {
                this.resurrectField('httpsProxyUrl', [
                    Validators.required,
                    this.validationService.isHttpOrHttps()
                ], this.formGroup.value['httpsProxyUrl'],
                    { onlySelf: true }
                );
            } else {
                const httpsFields = [
                    'httpsProxyUrl',
                    'httpsProxyUsername',
                    'httpsProxyPassword',
                ];
                httpsFields.forEach((field) => {
                    this.disarmField(field, true);
                });
            }
        } else {
            proxySettingFields.forEach((field) => {
                this.disarmField(field, true);
            });
        }
    }

    getCniTypeLabel() {
        if (this.cniType === "none") {
            return "None";
        } else if (this.cniType === "calico") {
            return "Calico";
        } else {
            return "Antrea"
        }
    }

    /**
     * @method loadVSphereNetworks
     * helper method retrieves list of vsphere networks
     */
    loadVSphereNetworks() {
        this.loadingNetworks = true;
        Broker.messenger.publish({
            type: TkgEventType.GET_VM_NETWORKS
        });
    }
    // Reset the relevant fields upon data center change
    resetFieldsUponDCChange() {
        const fieldsToReset = ['networkName'];
        fieldsToReset.forEach(f => this.formGroup.get(f) && this.formGroup.get(f).setValue('', { onlySelf: true }));
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
        // reset validations for httpProxyUrl and httpsProxyUrl when
        // the data is loaded from localstorage.
        this.toggleProxySetting(true);
        this.scrubPasswordField('httpProxyPassword');
        this.scrubPasswordField('httpsProxyPassword');
    }

    dynamicDescription(): string {
        const serviceCidr = this.getFieldValue('clusterServiceCidr', true);
        const podCidr = this.getFieldValue('clusterPodCidr', true);
        if (serviceCidr && podCidr) {
            return `Cluster service CIDR: ${serviceCidr} Cluster POD CIDR: ${podCidr}`;
        }
        if (podCidr) {
            return `Cluster Pod CIDR: ${podCidr}`;
        }
        return '';
    }
}
