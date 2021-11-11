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
import { managementClusterPlugin } from "../../../constants/wizard.constants";

declare var sortPaths: any;
@Component({
    selector: 'app-shared-network-step',
    templateUrl: './network-step.component.html',
    styleUrls: ['./network-step.component.scss']
})
export class SharedNetworkStepComponent extends StepFormDirective implements OnInit {
    @Input() enableNetworkName: boolean;

    form: FormGroup;
    cniType: string;
    loadingNetworks: boolean = false;
    vmNetworks: Array<VSphereNetwork>;
    additionalNoProxyInfo: string;
    fullNoProxy: string;

    constructor(private validationService: ValidationService,
        private wizardFormService: VSphereWizardFormService) {
        super();
    }

    ngOnInit() {
        super.ngOnInit();
        this.buildForm();
        this.listenToEvents();

        const cniTypeData = {
            label: 'CNI PROVIDER',
            displayValue: this.cniType,
        } as FormMetaData;
        FormMetaDataStore.saveMetaDataEntry(this.formName, 'cniType', cniTypeData);
        this.formGroup.get('cniType').setValue(this.cniType);
    }
    buildForm() {
        const fieldsMapping = [
            ['cniType', 'antrea'],
            ['clusterServiceCidr', this.ipFamily === IpFamilyEnum.IPv4 ?
                IAAS_DEFAULT_CIDRS.CLUSTER_SVC_CIDR : IAAS_DEFAULT_CIDRS.CLUSTER_SVC_IPV6_CIDR],
            ['clusterPodCidr', this.ipFamily === IpFamilyEnum.IPv4 ?
                IAAS_DEFAULT_CIDRS.CLUSTER_POD_CIDR : IAAS_DEFAULT_CIDRS.CLUSTER_POD_IPV6_CIDR],
            ['httpProxyUrl', ''],
            ['httpProxyUsername', ''],
            ['httpProxyPassword', ''],
            ['httpsProxyUrl', ''],
            ['httpsProxyUsername', ''],
            ['httpsProxyPassword', ''],
            ['noProxy', '']
        ];

        if (this.enableNetworkName) {
            fieldsMapping.push(['networkName', '']);
        } else {
            this.clearFieldSavedData('networkName');
        }
        fieldsMapping.forEach(field => {
            this.formGroup.addControl(field[0], new FormControl(field[1], []));
        });

        const cidrs = ['clusterServiceCidr', 'clusterPodCidr'];
        cidrs.forEach(cidr => {
            this.registerOnIpFamilyChange(cidr, [
                this.validationService.isValidIpNetworkSegment()], [
                this.validationService.isValidIpv6NetworkSegment(),
                this.setCidrs
            ]);
        });

        this.formGroup.addControl('proxySettings', new FormControl(false));
        this.formGroup.addControl('isSameAsHttp', new FormControl(true));
        this.setValidators();
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
                ]);
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
            ], this.ipFamily === IpFamilyEnum.IPv4 ? IAAS_DEFAULT_CIDRS.CLUSTER_SVC_CIDR : IAAS_DEFAULT_CIDRS.CLUSTER_SVC_IPV6_CIDR);
        }

        this.resurrectField('clusterPodCidr', [
            Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.ipFamily === IpFamilyEnum.IPv4 ?
                this.validationService.isValidIpNetworkSegment() : this.validationService.isValidIpv6NetworkSegment(),
            this.validationService.isIpUnique([this.formGroup.get('clusterServiceCidr')])
        ], this.ipFamily === IpFamilyEnum.IPv4 ? IAAS_DEFAULT_CIDRS.CLUSTER_POD_CIDR : IAAS_DEFAULT_CIDRS.CLUSTER_POD_IPV6_CIDR);
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
                    [Validators.required, this.validationService.isValidNameInList(
                        this.vmNetworks.map(vmNetwork => vmNetwork.displayName))], networks.length === 1 ? networks[0].name : '');
            });

        const noProxyFieldChangeMap = ['noProxy', 'clusterServiceCidr', 'clusterPodCidr'];

        noProxyFieldChangeMap.forEach((field) => {
            this.formGroup.get(field).valueChanges.pipe(
                distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
                takeUntil(this.unsubscribe)
            ).subscribe(() => {
                this.generateFullNoProxy();
            });
        });

        Broker.messenger.getSubject(TkgEventType.AWS_GET_NO_PROXY_INFO)
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

    toggleProxySetting() {
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
        if (this.formGroup.value['proxySettings']) {
            this.resurrectField('httpProxyUrl', [
                Validators.required,
                this.validationService.isHttpOrHttps()
            ], this.formGroup.value['httpProxyUrl']);
            if (!this.formGroup.value['isSameAsHttp']) {
                this.resurrectField('httpsProxyUrl', [
                    Validators.required,
                    this.validationService.isHttpOrHttps()
                ], this.formGroup.value['httpsProxyUrl']);
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
    // Reset the relevent fields upon data center change
    resetFieldsUponDCChange() {
        const fieldsToReset = ['networkName'];
        fieldsToReset.forEach(f => this.formGroup.get(f) && this.formGroup.get(f).setValue(''));
    }

    setSavedDataAfterLoad() {
        super.setSavedDataAfterLoad();
        if (this.formGroup.get('networkName')) {
            this.formGroup.get('networkName').setValue(this.vmNetworks.length === 1 ? this.vmNetworks[0].name : '');
        }
        // reset validations for httpProxyUrl and httpsProxyUrl when
        // the data is loaded from localstorage.
        this.toggleProxySetting();
        // don't fill password field with ****
        this.formGroup.get('httpProxyPassword').setValue('');
        this.formGroup.get('httpsProxyPassword').setValue('');
    }
}
