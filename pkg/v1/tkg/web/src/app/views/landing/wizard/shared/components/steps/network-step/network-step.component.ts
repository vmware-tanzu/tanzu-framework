// Angular imports
import { Component, Input, OnInit } from '@angular/core';
import { FormGroup, Validators } from '@angular/forms';
// Third party imports
import { distinctUntilChanged, takeUntil } from 'rxjs/operators';
// App imports
import AppServices from '../../../../../../../shared/service/appServices';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { FormMetaDataStore, FormMetaData } from '../../../FormMetaDataStore';
import { IAAS_DEFAULT_CIDRS, IpFamilyEnum } from '../../../../../../../shared/constants/app.constants';
import { managementClusterPlugin } from "../../../constants/wizard.constants";
import { NetworkField, NetworkIpv4StepMapping, NetworkIpv6StepMapping } from './network-step.fieldmapping';
import { StepFormDirective } from '../../../step-form/step-form';
import { StepMapping } from '../../../field-mapping/FieldMapping';
import { TanzuEventType } from 'src/app/shared/service/Messenger';
import { ValidationService } from '../../../validation/validation.service';
import { VSphereNetwork } from 'src/app/swagger/models/v-sphere-network.model';

declare var sortPaths: any;
@Component({
    selector: 'app-shared-network-step',
    templateUrl: './network-step.component.html',
    styleUrls: ['./network-step.component.scss']
})
export class SharedNetworkStepComponent extends StepFormDirective implements OnInit {
    static description  = 'Specify how TKG networking is provided and global network settings';
    enableNetworkName: boolean;

    form: FormGroup;
    cniType: string;
    vmNetworks: Array<VSphereNetwork> = [];
    additionalNoProxyInfo: string;
    fullNoProxy: string;
    infraServiceAddress: string = '';
    loadingNetworks: boolean = false;   // only used by vSphere
    hideNoProxyWarning: boolean = true; // only used by vSphere

    constructor(protected validationService: ValidationService) {
        super();
    }

    protected supplyEnablesNetworkName(): boolean {
        return false;
    }

    protected supplyEnablesNoProxyWarning(): boolean {
        return false;
    }

    protected supplyStepMapping(): StepMapping {
        return this.ipFamily === IpFamilyEnum.IPv4 ? NetworkIpv4StepMapping : NetworkIpv6StepMapping;
    }

    // This method may be overridden by subclasses that describe this step using different fields
    protected supplyFieldsAffectingStepDescription(): string[] {
        return [NetworkField.CLUSTER_SERVICE_CIDR, NetworkField.CLUSTER_POD_CIDR];
    }

    private customizeForm() {
        if (!this.enableNetworkName) {
            this.clearFieldSavedData(NetworkField.NETWORK_NAME);
            this.formGroup.removeControl(NetworkField.NETWORK_NAME);
        }

        const cidrs = [NetworkField.CLUSTER_SERVICE_CIDR, NetworkField.CLUSTER_POD_CIDR];
        cidrs.forEach(cidr => {
            this.registerOnIpFamilyChange(cidr, [
                this.validationService.isValidIpNetworkSegment()], [
                this.validationService.isValidIpv6NetworkSegment(),
                this.setCidrs
            ]);
        });
        this.registerStepDescriptionTriggers({fields: this.supplyFieldsAffectingStepDescription()});

        this.setValidators();
    }

    ngOnInit() {
        super.ngOnInit();
        AppServices.fieldMapUtilities.buildForm(this.formGroup, this.formName, this.supplyStepMapping());
        this.htmlFieldLabels = AppServices.fieldMapUtilities.getFieldLabelMap(this.supplyStepMapping());
        this.storeDefaultLabels(this.supplyStepMapping());

        this.customizeForm();
        this.listenToEvents();
        this.subscribeToServices();

        const cniTypeData = {
            label: 'CNI PROVIDER',
            displayValue: this.cniType,
        } as FormMetaData;
        FormMetaDataStore.saveMetaDataEntry(this.formName, NetworkField.CNI_TYPE, cniTypeData);
        // TODO: guessing we don't need this line (due to initFormWithSavedData() below)
        this.formGroup.get(NetworkField.CNI_TYPE).setValue(this.cniType, { onlySelf: true });
        this.initFormWithSavedData();
    }

    setValidators() {
        const configuredCni = AppServices.appDataService.getPluginFeature(managementClusterPlugin, 'cni');
        if (configuredCni && ['antrea', 'calico', 'none'].includes(configuredCni)) {
            this.cniType = configuredCni;
        } else {
            this.cniType = 'antrea';
        }

        if (this.cniType === 'none') {
            [NetworkField.CLUSTER_SERVICE_CIDR, NetworkField.CLUSTER_POD_CIDR].forEach( field => this.disarmField(field, false));
        } else {
            if (this.cniType === 'calico') {
                this.disarmField(NetworkField.CLUSTER_SERVICE_CIDR, false);
            }
            this.setCidrs();

            if (this.enableNetworkName) {
                this.resurrectField(NetworkField.NETWORK_NAME, [
                    Validators.required
                ], '', { onlySelf: true }); // only for current form control
            }
        }
    }
    setCidrs = () => {
        if (this.cniType === 'antrea') {
            this.resurrectField(NetworkField.CLUSTER_SERVICE_CIDR, [
                Validators.required,
                this.validationService.noWhitespaceOnEnds(),
                this.ipFamily === IpFamilyEnum.IPv4 ?
                    this.validationService.isValidIpNetworkSegment() : this.validationService.isValidIpv6NetworkSegment(),
                this.validationService.isIpUnique([this.formGroup.get(NetworkField.CLUSTER_POD_CIDR)])
            ], this.ipFamily === IpFamilyEnum.IPv4 ?
                IAAS_DEFAULT_CIDRS.CLUSTER_SVC_CIDR : IAAS_DEFAULT_CIDRS.CLUSTER_SVC_IPV6_CIDR, { onlySelf: true });
        }

        this.resurrectField(NetworkField.CLUSTER_POD_CIDR, [
            Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.ipFamily === IpFamilyEnum.IPv4 ?
                this.validationService.isValidIpNetworkSegment() : this.validationService.isValidIpv6NetworkSegment(),
            this.validationService.isIpUnique([this.formGroup.get(NetworkField.CLUSTER_SERVICE_CIDR)])
        ], this.ipFamily === IpFamilyEnum.IPv4 ?
            IAAS_DEFAULT_CIDRS.CLUSTER_POD_CIDR : IAAS_DEFAULT_CIDRS.CLUSTER_POD_IPV6_CIDR, { onlySelf: true });
    }

    listenToEvents() {
        this.listenToCidrEvents();
        this.listenToNoProxyEvents();
    }

    private listenToCidrEvents() {
        const cidrFields = [NetworkField.CLUSTER_SERVICE_CIDR, NetworkField.CLUSTER_POD_CIDR];
        cidrFields.forEach((field) => {
            this.formGroup.get(field).valueChanges.pipe(
                distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
                takeUntil(this.unsubscribe)
            ).subscribe((value) => {
                this.generateFullNoProxy();
                this.triggerStepDescriptionChange();
            });
        });
    }

    private listenToNoProxyEvents() {
        this.formGroup.get('noProxy').valueChanges.pipe(
            distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
            takeUntil(this.unsubscribe)
        ).subscribe((value) => {
            this.onNoProxyChange(value);
            this.triggerStepDescriptionChange();
        });

        AppServices.messenger.getSubject(TanzuEventType.NETWORK_STEP_GET_NO_PROXY_INFO)
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
        const noProxy = this.formGroup.get(NetworkField.NO_PROXY);
        if (noProxy && !noProxy.value) {
            this.fullNoProxy = '';
            return;
        }
        const clusterServiceCidr = this.formGroup.get(NetworkField.CLUSTER_SERVICE_CIDR);
        const clusterPodCidr = this.formGroup.get(NetworkField.CLUSTER_POD_CIDR);

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
            NetworkField.HTTP_PROXY_URL,
            NetworkField.HTTP_PROXY_USERNAME,
            NetworkField.HTTP_PROXY_PASSWORD,
            NetworkField.HTTPS_IS_SAME_AS_HTTP,
            NetworkField.HTTPS_PROXY_URL,
            NetworkField.HTTPS_PROXY_USERNAME,
            NetworkField.HTTPS_PROXY_PASSWORD,
            NetworkField.NO_PROXY
        ];

        if (!fromSavedData) {
            this.formGroup.markAsPending();
        }

        if (this.formGroup.value[NetworkField.PROXY_SETTINGS]) {
            this.resurrectField(NetworkField.HTTP_PROXY_URL, [
                Validators.required,
                this.validationService.isHttpOrHttps()
            ], this.formGroup.value[NetworkField.HTTP_PROXY_URL],
                { onlySelf: true }
            );
            this.resurrectField(NetworkField.NO_PROXY, [],
                this.formGroup.value[NetworkField.NO_PROXY] || this.infraServiceAddress,
                { onlySelf: true }
            );
            if (!this.formGroup.value[NetworkField.HTTPS_IS_SAME_AS_HTTP]) {
                this.resurrectField(NetworkField.HTTPS_PROXY_URL, [
                    Validators.required,
                    this.validationService.isHttpOrHttps()
                ], this.formGroup.value[NetworkField.HTTPS_PROXY_URL],
                    { onlySelf: true }
                );
            } else {
                const httpsFields = [
                    NetworkField.HTTPS_PROXY_URL,
                    NetworkField.HTTPS_PROXY_USERNAME,
                    NetworkField.HTTPS_PROXY_PASSWORD,
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
        this.scrubPasswordField(NetworkField.HTTP_PROXY_PASSWORD);
        this.scrubPasswordField(NetworkField.HTTPS_PROXY_PASSWORD);
    }

    dynamicDescription(): string {
        const serviceCidr = this.getFieldValue(NetworkField.CLUSTER_SERVICE_CIDR, true);
        const podCidr = this.getFieldValue(NetworkField.CLUSTER_POD_CIDR, true);
        if (serviceCidr && podCidr) {
            return `Cluster Service CIDR: ${serviceCidr} Cluster Pod CIDR: ${podCidr}`;
        }
        if (podCidr) {
            return `Cluster Pod CIDR: ${podCidr}`;
        }
        return SharedNetworkStepComponent.description;
    }

    // allows subclasses to subscribe to services during ngOnInit by overriding this method
    protected subscribeToServices() {
    }

    protected storeUserData() {
        this.storeUserDataFromMapping(this.supplyStepMapping());
        this.storeDefaultDisplayOrder(this.supplyStepMapping());
    }

    get enableNoProxyWarning(): boolean {
        return this.supplyEnablesNoProxyWarning();
    }
    get enableNetworkName(): boolean {
        return this.supplyEnablesNetworkName();
    }
}
