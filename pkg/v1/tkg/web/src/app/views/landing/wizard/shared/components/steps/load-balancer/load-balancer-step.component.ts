/**
 * Angular Modules
 */
import { Component, OnInit } from '@angular/core';
import { FormControl, Validators } from '@angular/forms';

import { debounceTime, distinctUntilChanged, finalize, takeUntil } from 'rxjs/operators';
import { StepFormDirective } from "../../../step-form/step-form";
import { ValidationService } from "../../../validation/validation.service";
import { APIClient } from "../../../../../../../swagger";

import { AviCloud } from "src/app/swagger/models/avi-cloud.model";
import { AviServiceEngineGroup } from "src/app/swagger/models/avi-service-engine-group.model";
import { ClrLoadingState } from "@clr/angular";
import { VSphereWizardFormService } from 'src/app/shared/service/vsphere-wizard-form.service';
import { AviVipNetwork } from './../../../../../../../swagger/models/avi-vip-network.model';
import { TkgEventType } from 'src/app/shared/service/Messenger';
import Broker from 'src/app/shared/service/broker';
import { IpFamilyEnum } from 'src/app/shared/constants/app.constants';
import { FormUtils } from '../../../utils/form-utils';

export const KUBE_VIP = 'Kube-vip';
export const NSX_ADVANCED_LOAD_BALANCER = "NSX Advanced Load Balancer";

const SupervisedFields = ['controllerHost', 'username', 'password', 'controllerCert'];
const HA_REQUIRED_FIELDS = [
    'controllerHost',
    'username',
    'password',
    'controllerCert',
    'managementClusterNetworkName',
    'managementClusterNetworkCIDR'
]

@Component({
    selector: 'app-load-balancer-step',
    templateUrl: './load-balancer-step.component.html',
    styleUrls: ['./load-balancer-step.component.scss']
})
export class SharedLoadBalancerStepComponent extends StepFormDirective implements OnInit {

    loadingState: ClrLoadingState = ClrLoadingState.DEFAULT;
    connected: boolean = false;
    clouds: Array<AviCloud>;
    selectedCloud: AviCloud;
    selectedCloudName: string;
    serviceEngineGroups: Array<AviServiceEngineGroup>;
    serviceEngineGroupsFiltered: Array<AviServiceEngineGroup>;
    labels: Map<String, String> = new Map<String, String>();
    vipClusterNetworkNameLabel: string;
    vipClusterNetworkCidrLabel: string;
    vipNetworks: Array<AviVipNetwork> = [];
    selectedNetworkName: string;
    selectedManagementClusterNetworkName: string;
    currentControlPlaneEndpoingProvider: string;

    constructor(private validationService: ValidationService,
        private apiClient: APIClient, private wizardFormService: VSphereWizardFormService) {
        super();
    }

    ngOnInit() {
        super.ngOnInit();

        this.vipClusterNetworkNameLabel = this.modeClusterStandalone ?
            'STANDALONE CLUSTER VIP NETWORK NAME' : 'MANAGEMENT VIP NETWORK NAME';
        this.vipClusterNetworkCidrLabel = this.modeClusterStandalone ?
            'STANDALONE CLUSTER VIP NETWORK CIDR' : 'MANAGEMENT VIP NETWORK CIDR';

        FormUtils.addControl(
            this.formGroup,
            'controllerHost',
            new FormControl('', [
                this.validationService.isValidIpOrFqdn()
            ])
        );
        FormUtils.addControl(
            this.formGroup,
            'username',
            new FormControl('', [])
        );
        FormUtils.addControl(
            this.formGroup,
            'password',
            new FormControl('', [])
        );

        FormUtils.addControl(
            this.formGroup,
            'cloudName',
            new FormControl('', [])
        );
        FormUtils.addControl(
            this.formGroup,
            'serviceEngineGroupName',
            new FormControl('', [])
        );

        FormUtils.addControl(
            this.formGroup,
            'networkName',
            new FormControl('', [])
        );
        FormUtils.addControl(
            this.formGroup,
            'networkCIDR',
            new FormControl('', [])
        );
        FormUtils.addControl(
            this.formGroup,
            'managementClusterNetworkName',
            new FormControl('', [])
        );
        FormUtils.addControl(
            this.formGroup,
            'managementClusterNetworkCIDR',
            new FormControl('', [this.validationService.isValidIpNetworkSegment()])
        );
        FormUtils.addControl(
            this.formGroup,
            'controllerCert',
            new FormControl('', [])
        );
        FormUtils.addControl(
            this.formGroup,
            'clusterLabels',
            new FormControl('', [])
        );
        FormUtils.addControl(
            this.formGroup,
            'newLabelKey',
            new FormControl('', [
                this.validationService.isValidLabelOrAnnotation()
            ])
        );
        FormUtils.addControl(
            this.formGroup,
            'newLabelValue',
            new FormControl('', [
                this.validationService.isValidLabelOrAnnotation()
            ])
        );

        SupervisedFields.forEach(field => {
            this.formGroup.get(field).valueChanges
                .pipe(
                    debounceTime(500),
                    distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
                    takeUntil(this.unsubscribe)
                )
                .subscribe(() => {
                    this.connected = false;
                    this.disarmField('cloudName', true);
                    this.clouds = [];
                    this.disarmField('serviceEngineGroupName', true);
                    this.serviceEngineGroups = [];
                    this.disarmField('networkCIDR', true);
                    this.disarmField('managementClusterNetworkCIDR', true);

                    // If connection cleared, toggle validators OFF
                    this.toggleValidators(false);
                });
        });

        this.formGroup.get('cloudName').valueChanges.pipe(
            distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
            takeUntil(this.unsubscribe)
        ).subscribe((cloud) => {
            this.selectedCloudName = cloud;
            this.onSelectCloud(this.selectedCloudName);
        });

        this.registerOnValueChange("networkName", this.onSelectVipNetwork.bind(this));
        this.registerOnValueChange("networkCIDR", this.onSelectVipCIDR.bind(this));
        this.registerOnValueChange("managementClusterNetworkName", this.onSelectManagementNetwork.bind(this));

        Broker.messenger.getSubject(TkgEventType.CONTROL_PLANE_ENDPOINT_PROVIDER_CHANGED)
            .subscribe(({ payload }) => {
                this.currentControlPlaneEndpoingProvider = payload;
                if (this.currentControlPlaneEndpoingProvider === NSX_ADVANCED_LOAD_BALANCER) {
                    HA_REQUIRED_FIELDS.forEach(fieldName => this.resurrectField(fieldName, [Validators.required]));
                } else {
                    HA_REQUIRED_FIELDS.forEach(fieldName => this.disarmField(fieldName, true));
                }
            });
        this.registerOnIpFamilyChange('networkCIDR', [], []);
        this.registerOnIpFamilyChange('managementClusterNetworkCIDR', [
            this.validationService.isValidIpNetworkSegment()], [
            this.validationService.isValidIpv6NetworkSegment()
        ]);

        this.initFormWithSavedData();
    }

    isFieldReadyForInitWithSavedValue(fieldName: string): boolean {
        if (fieldName === "cloudName") {
            return !this.isEmptyArray(this.serviceEngineGroups) && !this.isEmptyArray(this.vipNetworks);
        }
        return true;
    }

    initFormWithSavedData() {
        const fieldControllerHost = this.formGroup.get('controllerHost');
        if (fieldControllerHost) {
            fieldControllerHost.setValue(this.getSavedValue('controllerHost', ''));
        }
        const fieldUserName = this.formGroup.get('username');
        if (fieldUserName) {
            fieldUserName.setValue(this.getSavedValue('username', ''));
        }

        const savedLabelsString = this.getSavedValue('clusterLabels', '');
        if (savedLabelsString !== '') {
            const savedLabelsArray = savedLabelsString.split(', ')
            savedLabelsArray.map(label => {
                const labelArray = label.split(':');
                this.labels.set(labelArray[0], labelArray[1]);
            });
        }

        // clear password from saved data
        const fieldPassword = this.formGroup.get('password');
        if (fieldPassword) {
            fieldPassword.setValue('');
        }

        this.startProcessDelayedFieldInit();    // init those fields with saved value when they become ready
    }

    /**
     * @method connectLB
     * helper method to make connection to AVI LB controller, call getClouds and
     * getSvcEngineGroups methods if AVI LB controller connection successful
     */
    connectLB() {
        this.loadingState = ClrLoadingState.LOADING;

        this.apiClient.verifyAccount({
            credentials: {
                username: this.formGroup.controls['username'].value,
                password: this.formGroup.controls['password'].value,
                host: this.formGroup.controls['controllerHost'].value,
                tenant: this.formGroup.controls['username'].value,
                CAData: this.formGroup.controls['controllerCert'].value
            }
        })
            .pipe(
                finalize(() => this.loadingState = ClrLoadingState.DEFAULT),
                takeUntil(this.unsubscribe))
            .subscribe(
                ((res) => {
                    this.errorNotification = '';
                    this.connected = true;

                    this.getClouds();
                    this.getServiceEngineGroups();
                    this.getVipNetworks();

                    // If connection successful, toggle validators ON
                    this.toggleValidators(true);
                }),
                ((err) => {
                    const error = err.error.message || err.message || JSON.stringify(err);
                    if (error.indexOf('Invalid credentials') >= 0) {
                        this.errorNotification = `Invalid credentials: check your username and password`;
                    } else if (error.indexOf('Rest request error, returning to caller') >= 0) {
                        this.errorNotification = `Invalid Controller Certificate Authority: check the validity of the certificate`;
                    } else {
                        this.errorNotification = `Failed to connect to the specified Avi load balancer controller. ${error}`;
                    }
                })
            );
    }

    /**
     * @method getClouds
     * helper method calls API to get list of clouds
     */
    getClouds() {
        this.apiClient.getAviClouds()
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(
                ((res) => {
                    this.errorNotification = '';
                    this.clouds = res;
                }),
                ((err) => {
                    const error = err.error.message || err.message || JSON.stringify(err);
                    this.errorNotification =
                        `Failed to retrieve Avi load balancer clouds list. ${error}`;
                })
            );
    }

    /**
     * @method getServiceEngineGroups
     * helper method calls API to get list of service engine groups
     */
    getServiceEngineGroups() {
        this.apiClient.getAviServiceEngineGroups()
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(
                ((res) => {
                    this.errorNotification = '';
                    this.serviceEngineGroups = res;
                    // refreshes service engine group list which may not have been loaded at time of cloud selection
                    this.onSelectCloud(this.selectedCloudName);
                }),
                ((err) => {
                    const error = err.error.message || err.message || JSON.stringify(err);
                    this.errorNotification =
                        `Failed to retrieve Avi load balancer service engine groups list. ${error}`;
                })
            );
    }

    /**
     * @method getServiceEngineGroups
     * helper method calls API to get list of service engine groups
     */
    getVipNetworks() {
        this.apiClient.getAviVipNetworks()
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(
                ((res) => {
                    this.errorNotification = '';
                    this.vipNetworks = res;
                }),
                ((err) => {
                    const error = err.error.message || err.message || JSON.stringify(err);
                    this.errorNotification =
                        `Failed to retrieve Avi VIP networks. ${error}`;
                })
            );
    }

    /**
     * @method onSelectCloud
     * @param cloudName - name of selected cloud
     * helper method sets selected cloud object based on cloud 'name' returned from dropdown; then
     * filters list of service engine groups where the cloud uuid matches in 'location'
     */
    onSelectCloud(cloudName: string) {
        this.serviceEngineGroupsFiltered = [];

        if (cloudName) {
            this.selectedCloud = this.clouds.find((cloud: AviCloud) => {
                return cloud.name === cloudName;
            });

            if (this.selectedCloud) {
                this.serviceEngineGroupsFiltered = this.serviceEngineGroups.filter((group: AviServiceEngineGroup) => {
                    return group.location.includes(this.selectedCloud.uuid);
                });
            }
        }
    }

    /**
     * Return all the configured subnets for the selected vip network.
     */
    onSelectVipNetwork(networkName: string): void {
        this.selectedNetworkName = networkName;
        if (!this.formGroup.get("managementClusterNetworkName").value) { }
        this.formGroup.get("managementClusterNetworkName").setValue(networkName)
    }

    onSelectVipCIDR(cidr: string): void {
        if (!this.formGroup.get("managementClusterNetworkCIDR").value) {
                this.formGroup.get("managementClusterNetworkCIDR").setValue(cidr);
        }
    }

    onSelectManagementNetwork(networkName: string): void {
        this.selectedManagementClusterNetworkName = networkName;
    }

    /**
     * @method toggleValidators
     * @param validate - boolean if true activates all fields and sets their validators; if false
     * disarms fields and clears validators
     */
    toggleValidators(validate: boolean) {
        if (validate === true) {
            this.resurrectField('cloudName', [Validators.required],
                this.getSavedValue('cloudName', ''));
            this.resurrectField('serviceEngineGroupName', [Validators.required],
                this.getSavedValue('serviceEngineGroupName', ''));
            if (!this.modeClusterStandalone) {
                this.resurrectField('networkName', [Validators.required],
                    this.getSavedValue('networkName', ''));
                this.resurrectField('networkCIDR', [
                    Validators.required,
                    this.validationService.noWhitespaceOnEnds(),
                    this.ipFamily === IpFamilyEnum.IPv4 ?
                        this.validationService.isValidIpNetworkSegment() : this.validationService.isValidIpv6NetworkSegment()
                ], this.getSavedValue('networkCIDR', ''));
            }
        } else {
            this.disarmField('cloudName', true);
            this.disarmField('serviceEngineGroupName', true);
            if (!this.modeClusterStandalone) {
                this.disarmField('networkName', true);
                this.disarmField('networkCIDR', true);
            }
        }
    }

    /**
     * @method getDisabled
     * helper method to get if connect btn should be disabled
     */
    getDisabled(): boolean {
        return (
            SupervisedFields.some(f => !this.formGroup.get(f).value) ||
                SupervisedFields.some(f => !this.formGroup.get(f).valid)
        )
    }

    /**
     * Add workload cluster label'
     */
    addLabel(key: string, value: string) {
        if (key === '' || value === '') {
            this.errorNotification = `Key and value for Labels are required.`;
        } else if (!this.labels.has(key)) {
            this.labels.set(key, value);
            this.formGroup.get('clusterLabels').setValue(this.labels);
            this.formGroup.controls['newLabelKey'].setValue('');
            this.formGroup.controls['newLabelValue'].setValue('');
        } else {
            this.errorNotification = `A Label with the same key already exists.`;
        }
    }

    /**
     * Delete workload cluster label'
     */
    deleteLabel(key: string) {
        this.labels.delete(key);
        this.formGroup.get('clusterLabels').setValue(this.labels);
    }

    /**
     * Get the current value of 'clusterLabels'
     */
    get clusterLabelsValue() {
        let labelsStr: string = '';
        this.labels.forEach((value: string, key: string) => {
            labelsStr += key + ':' + value + ', '
        });
        return labelsStr.slice(0, -2);
    }

    /**
     * @method getLabelDisabled
     * helper method to get if label add btn should be disabled
     */
    getLabelDisabled(): boolean {
        return !(this.formGroup.get('newLabelKey').valid &&
            this.formGroup.get('newLabelValue').valid);
    }

    /**
     * This is to make sense that the list returned is always up to date.
     */
    get vipNetworksPerCloud() {
        if (this.vipNetworks && this.vipNetworks.length > 0 && this.selectedCloud) {
            return this.vipNetworks.filter(net => net.cloud === this.selectedCloud.uuid);
        }
        return [];
    }

    getSubnets(networkName: string): any[] {
        if (!this.isEmptyArray(this.vipNetworksPerCloud) && networkName) {
            const temp = this.vipNetworksPerCloud
                .find(net => net.name === networkName);

            if (temp && !this.isEmptyArray(temp.configedSubnets)) {
                return temp.configedSubnets
                    .filter(subnet => subnet.family === "V4");      // Only V4 are supported in Calgary 1.
            }
        }
        return [];
    }
    /**
     * This is to make sense that the list returned is always up to date.
     */
    get subnetsPerNetwork() {
        return this.getSubnets(this.selectedNetworkName);
    }

    /**
     * This is to make sense that the list returned is always up to date.
     */
    get subnetsPerManagementNetwork() {
        return this.getSubnets(this.selectedManagementClusterNetworkName);
    }

}
