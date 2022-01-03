// Angular imports
import { Component, Input, OnInit } from '@angular/core';
import { FormGroup, Validators } from '@angular/forms';
// Third party imports
import { distinctUntilChanged, takeUntil } from 'rxjs/operators';
// App imports
import Broker from 'src/app/shared/service/broker';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { FormMetaDataStore, FormMetaData } from '../../../FormMetaDataStore';
import { IAAS_DEFAULT_CIDRS, IpFamilyEnum } from '../../../../../../../shared/constants/app.constants';
import { managementClusterPlugin } from "../../../constants/wizard.constants";
import { NetworkIpv4StepMapping, NetworkIpv6StepMapping } from './network-step.fieldmapping';
import ServiceBroker from '../../../../../../../shared/service/service-broker';
import { StepFormDirective } from '../../../step-form/step-form';
import { StepMapping } from '../../../field-mapping/FieldMapping';
import { TkgEventType } from 'src/app/shared/service/Messenger';
import { ValidationService } from '../../../validation/validation.service';
import { VSphereNetwork } from 'src/app/swagger/models/v-sphere-network.model';

declare var sortPaths: any;
@Component({
    selector: 'app-shared-network-step',
    templateUrl: './network-step.component.html',
    styleUrls: ['./network-step.component.scss']
})
export class SharedNetworkStepComponent extends StepFormDirective implements OnInit {
    enableNetworkName: boolean;

    form: FormGroup;
    cniType: string;
    vmNetworks: Array<VSphereNetwork>;
    additionalNoProxyInfo: string;
    fullNoProxy: string;
    infraServiceAddress: string = '';
    loadingNetworks: boolean = false;   // only used by vSphere
    hideNoProxyWarning: boolean = true; // only used by vSphere

    constructor(protected validationService: ValidationService,
                protected fieldMapUtilities: FieldMapUtilities,
                protected serviceBroker: ServiceBroker
                ) {
        super();
    }

    private supplyStepMapping(): StepMapping {
        return this.ipFamily === IpFamilyEnum.IPv4 ? NetworkIpv4StepMapping : NetworkIpv6StepMapping;
    }

    private customizeForm() {
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
        this.fieldMapUtilities.buildForm(this.formGroup, this.formName, this.supplyStepMapping());
        this.customizeForm();
        this.listenToEvents();
        this.subscribeToServices();

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
        this.listenToCidrEvents();
        this.listenToNoProxyEvents();
    }

    private listenToCidrEvents() {
        const cidrFields = ['clusterServiceCidr', 'clusterPodCidr'];
        cidrFields.forEach((field) => {
            this.formGroup.get(field).valueChanges.pipe(
                distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
                takeUntil(this.unsubscribe)
            ).subscribe((value) => {
                this.generateFullNoProxy();
            });
        });
    }

    private listenToNoProxyEvents() {
        this.formGroup.get('noProxy').valueChanges.pipe(
            distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
            takeUntil(this.unsubscribe)
        ).subscribe((value) => {
            this.onNoProxyChange(value);
        });

        Broker.messenger.getSubject(TkgEventType.NETWORK_STEP_GET_NO_PROXY_INFO)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                this.additionalNoProxyInfo = event.payload.info;
                this.generateFullNoProxy();
            });
    }

    // onNoProxyChange() is protected to allow subclasses to override
    protected onNoProxyChange(value: string) {
        this.generateFullNoProxy();
    }

    // This is a method only implemented by the vSphere child class (which overrides this method);
    // we need a method in this class because the general HTML references it;
    // however it should only be called when enableNetworkName is true (which only the vSphere subclass sets)
    loadNetworks() {
        console.error('loadNetworks() was called, but no implementation is available. (enableNetworkName= ' + this.enableNetworkName + ')');
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

    initFormWithSavedData() {
        super.initFormWithSavedData();
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

    // allows subclasses to subscribe to services during ngOnInit by overriding this method
    protected subscribeToServices() {
    }
}
