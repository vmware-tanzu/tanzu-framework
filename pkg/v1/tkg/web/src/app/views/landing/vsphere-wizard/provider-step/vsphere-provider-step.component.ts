// Angular imports
import { Component, OnInit, ViewChild } from '@angular/core';
import { Router } from '@angular/router';
import { Validators } from '@angular/forms';
// Third party imports
import { debounceTime, distinctUntilChanged, finalize, takeUntil } from 'rxjs/operators';
import * as _ from 'lodash';
import { ClrLoadingState } from '@clr/angular';
// App imports
import { AppEdition } from 'src/app/shared/constants/branding.constants';
import { APIClient } from 'src/app/swagger/api-client.service';
import { APP_ROUTES, Routes } from 'src/app/shared/constants/routes.constants';
import AppServices from 'src/app/shared/service/appServices';
import { EditionData } from 'src/app/shared/service/branding.service';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { FormMetaDataStore } from '../../wizard/shared/FormMetaDataStore';
import { managementClusterPlugin } from "../../wizard/shared/constants/wizard.constants";
import { NotificationTypes } from "../../../../shared/components/alert-notification/alert-notification.component";
import { SSLThumbprintModalComponent } from '../../wizard/shared/components/modals/ssl-thumbprint-modal/ssl-thumbprint-modal.component';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { TanzuEvent, TanzuEventType } from 'src/app/shared/service/Messenger';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { VSphereDatacenter } from 'src/app/swagger/models/v-sphere-datacenter.model';
import { VsphereField } from "../vsphere-wizard.constants";
import { VsphereProviderStepFieldMapping } from './vsphere-provider-step.fieldmapping';

declare var sortPaths: any;

const SupervisedField = [VsphereField.PROVIDER_VCENTER_ADDRESS, VsphereField.PROVIDER_USER_NAME, VsphereField.PROVIDER_USER_PASSWORD];

/**
 * vSphere Version Info definition
 */
export interface VsphereVersioninfo {
    version: string,
    build: string;
}

@Component({
    selector: 'app-vsphere-provider-step',
    templateUrl: './vsphere-provider-step.component.html',
    styleUrls: ['./vsphere-provider-step.component.scss']
})

export class VSphereProviderStepComponent extends StepFormDirective implements OnInit {
    @ViewChild(SSLThumbprintModalComponent) sslThumbprintModal: SSLThumbprintModalComponent;
    fileReader: FileReader;

    APP_ROUTES: Routes = APP_ROUTES;

    loading: boolean = false;
    connected: boolean = false;
    loadingState: ClrLoadingState = ClrLoadingState.DEFAULT;
    vSphereWithK8ModalOpen: boolean = false;
    datacenters: Array<VSphereDatacenter>;
    vsphereVersion: string = '';
    vsphereHost: string;
    hasPacific: string;

    vSphereModalTitle: string;
    vSphereModalBody: string;
    thumbprint: string;

    edition: AppEdition = AppEdition.TCE;
    enableIpv6: boolean = false;

    // As a hack to avoid onChange events where the value really hasn't changed, we track the field values ourselves
    supervisedFieldValues = new Map<string, string>();

    constructor(private validationService: ValidationService,
                private apiClient: APIClient,
                private router: Router) {
        super();

        this.fileReader = new FileReader();
    }

    private customizeForm() {
        SupervisedField.forEach(field => {
            this.formGroup.get(field).valueChanges
                .pipe(
                    debounceTime(500),
                    distinctUntilChanged(),
                    takeUntil(this.unsubscribe)
                )
                .subscribe((data) => {
                    const oldValue = this.supervisedFieldValues.get(field);
                    const same = data === oldValue;
                    if (same) {
                        console.log('IGNORING change event from field ' + field + ' because value is unchanged: ' +
                        oldValue + '-->' + data);
                    } else {
                        this.supervisedFieldValues.set(field, data);
                        const msg = 'disconnecting due to change event from field ' + field + ' value changed: ' +
                            oldValue + '-->' + data;
                        this.disconnect(msg);
                    }
                });
        });

        this.formGroup.get(VsphereField.PROVIDER_DATA_CENTER).valueChanges.subscribe(data => {
            this.dcOnChange(data)
        });

        if (this.enableIpv6) {
            this.formGroup.get(VsphereField.PROVIDER_IP_FAMILY).valueChanges
                .pipe(
                    distinctUntilChanged((prev, curr) => {
                        const same = prev === curr;
                        console.log('field PROVIDER_IP_FAMILY detects ' + !same + ' change from ' + prev + ' to ' + curr);
                        return same;
                    }),
                    takeUntil(this.unsubscribe)
                )
                .subscribe(data => {
                    // In theory, we should only receive this event if the ipFamily actually changed. In practice, we double-check.
                    const same = data === this.ipFamily;
                    if (!same) {
                        AppServices.messenger.publish({
                            type: TanzuEventType.VSPHERE_IP_FAMILY_CHANGE,
                            payload: data
                        });
                        this.disconnect('disconnecting because field PROVIDER_IP_FAMILY changed value to ' + data);
                    }
                }
            );
        }
        AppServices.messenger.subscribe<EditionData>(TanzuEventType.BRANDING_CHANGED, data => { this.edition = data.payload.edition; });

        this.registerDefaultFileImportedHandler(TanzuEventType.VSPHERE_CONFIG_FILE_IMPORTED, VsphereProviderStepFieldMapping);
        this.registerDefaultFileImportErrorHandler(TanzuEventType.VSPHERE_CONFIG_FILE_IMPORT_ERROR);

        this.fileReader.onload = (event) => {
            try {
                const sshKey = event.target['result'];
                this.formGroup.get(VsphereField.PROVIDER_SSH_KEY).setValue(sshKey);
            } catch (error) {
                console.error('Error reading SSH key file: ' + error.message);
                return;
            }
            this.formGroup.get(VsphereField.PROVIDER_SSH_KEY_FILE).setValue('');
        };
        this.registerOnIpFamilyChange(VsphereField.PROVIDER_VCENTER_ADDRESS, [
            Validators.required,
            this.validationService.isValidIpOrFqdn()
        ], [
            Validators.required,
            this.validationService.isValidIpv6OrFqdn()
        ]);
    }

    ngOnInit() {
        super.ngOnInit();
        this.enableIpv6 = AppServices.appDataService.isPluginFeatureActivated(managementClusterPlugin, 'vsphereIPv6');
        AppServices.fieldMapUtilities.buildForm(this.formGroup, this.wizardName, this.formName, VsphereProviderStepFieldMapping);
        this.htmlFieldLabels = AppServices.fieldMapUtilities.getFieldLabelMap(VsphereProviderStepFieldMapping);
        this.storeDefaultLabels(VsphereProviderStepFieldMapping);

        this.formGroup.get(VsphereField.PROVIDER_DATA_CENTER).disable({ emitEvent: false});
        this.datacenters = [];
        this.formGroup.get(VsphereField.PROVIDER_SSH_KEY).disable({ emitEvent: false});
        this.customizeForm();

        this.initFormWithSavedData();
    }

    private disconnect(consoleMsg?: string) {
        if (this.connected) {
            if (consoleMsg) {
                console.log(consoleMsg);
            }
            this.connected = false;
            this.loadingState = ClrLoadingState.DEFAULT;
            this.formGroup.markAsPending(); // a temporary fix to ignore this.formGroup.statusChanges detection
            this.clearControlValue(VsphereField.PROVIDER_DATA_CENTER);
            this.datacenters = [];
            this.formGroup.get(VsphereField.PROVIDER_DATA_CENTER).disable();
            this.triggerStepDescriptionChange();
        } else {
            console.log('already disconnected so ignoring disconnect call (msg:' + consoleMsg + ')');
        }
    }
    initFormWithSavedData() {
        super.initFormWithSavedData();
        this.scrubPasswordField(VsphereField.PROVIDER_USER_PASSWORD);
    }

    /**
     * @method getParsedVsphereVer
     * @param res - the vsphere version string
     * helper method to parse vsphere version/build string into object
     * @returns {{version: (string|any), build: (string|any|number)}}
     */
    getParsedVsphereVer(res: string) {
        const versionInfo = _.split(res, ':');

        return {
            version: versionInfo[0],
            build: versionInfo[1] || '0'
        }
    }

    /**
     * @method checkVsphereCompatible
     * @param versionInfo
     * @returns {boolean}
     */
    checkVsphereCompatible(versionInfo: VsphereVersioninfo) {
        const compatibleVersion = [6, 7, 0];
        const compatibleBuild = 14367737;
        const version: Array<string> = _.split(versionInfo.version, '.');
        let returnVal = true;

        // corner case to check if build is u3 or newer for 6.7.0
        if (versionInfo.version === '6.7.0') {
            return (_.toNumber(versionInfo.build) >= compatibleBuild);
        } else if (_.toNumber(version[0]) >= 7) {
            return returnVal;
        } else {
            // for all others check major, minor, and patch value
            version.forEach((item, index) => {
                if (_.toNumber(item) < compatibleVersion[index]) {
                    returnVal = false;
                }
            });

            return returnVal;
        }
    }

    /**
     * @method connectVC
     * helper method to make connection to VC environment, call retrieveDatacenters
     * method if VC connection successful
     */
    connectVC() {
        this.loadingState = ClrLoadingState.LOADING;
        this.vsphereHost = this.getFieldValue(VsphereField.PROVIDER_VCENTER_ADDRESS);

        // we're recording these values to avoid onChange events where the value hasn't actually changed
        SupervisedField.forEach(field => {
            this.supervisedFieldValues.set(field, this.getFieldValue(field));
        })

        const insecure = this.getFieldValue(VsphereField.PROVIDER_CONNECTION_INSECURE);
        if (insecure) {
            this.login();
        } else {
            this.verifyThumbprint();
        }
    }

    verifyThumbprint() {
        this.apiClient.getVsphereThumbprint({
            host: this.vsphereHost
        })
            .pipe(
                finalize(() => this.loadingState = ClrLoadingState.DEFAULT),
                takeUntil(this.unsubscribe))
            .subscribe(
                ({thumbprint, insecure}) => {
                    if (insecure) {
                        this.login();
                        console.log('vSphere Insecure set true via VSPHERE_INSECURE environment variable. Bypassing thumbprint verification modal.')
                    } else {
                        this.thumbprint = thumbprint;
                        this.formGroup.controls[VsphereField.PROVIDER_THUMBPRINT].setValue(thumbprint);
                        FormMetaDataStore.saveMetaDataEntry(this.formName, VsphereField.PROVIDER_THUMBPRINT, {
                            label: 'SSL THUMBPRINT',
                            displayValue: thumbprint
                        });
                        this.sslThumbprintModal.open();
                    }
                },
                (err) => {
                    const error = err.error.message || err.message || JSON.stringify(err);
                    this.errorNotification =
                        `Failed to connect to the specified vCenter Server. ${ error }`;
                }
            );
    }

    thumbprintModalResponse(validThumbprint: boolean) {
        if (validThumbprint) {
            this.login();
        } else {
            this.errorNotification = "Connection failed. Certificate thumbprint was not validated.";
        }
    }

    login() {
        this.loadingState = ClrLoadingState.LOADING;
        this.apiClient.setVSphereEndpoint({
            credentials: {
                username: this.formGroup.controls[VsphereField.PROVIDER_USER_NAME].value,
                password: this.formGroup.controls[VsphereField.PROVIDER_USER_PASSWORD].value,
                host: this.formGroup.controls[VsphereField.PROVIDER_VCENTER_ADDRESS].value,
                insecure: this.formGroup.controls[VsphereField.PROVIDER_CONNECTION_INSECURE].value,
                thumbprint: this.thumbprint
            }
        })
        .pipe(
            finalize(() => this.loadingState = ClrLoadingState.DEFAULT),
            takeUntil(this.unsubscribe))
        .subscribe(
            this.onLoginSuccess.bind(this),
            (err) => {
                const errMsg = err.error ? err.error.message : null;
                const error = errMsg || err.message || JSON.stringify(err);
                this.errorNotification = `Failed to connect to the specified vCenter Server. ${error}`;
            }
        );
    }

    // TODO: this method is public so that tests can trigger it, but an extending class would be better
    onLoginSuccess(res) {
        const vsphereVerInfo: VsphereVersioninfo = this.getParsedVsphereVer(res.version);
        const isCompatible: boolean = this.checkVsphereCompatible(vsphereVerInfo);

        this.errorNotification = '';
        this.connected = true;
        this.vsphereVersion = vsphereVerInfo.version;
        this.hasPacific = res.hasPacific;
        AppServices.appDataService.setVsphereVersion(vsphereVerInfo.version);
        this.triggerStepDescriptionChange();

        if (isCompatible && !(_.startsWith(this.vsphereVersion, '6'))
        && this.edition !== AppEdition.TCE) {
            // for 7 and newer and other potential anomalies, show modal suggesting upgrade
            this.showVSphereWithK8Modal();
        } else if (!isCompatible) {
            // route to vsphere not compatible
            this.router.navigate([this.APP_ROUTES.INCOMPATIBLE]);
        }
        this.retrieveDatacenters();

        AppServices.messenger.publish({
            type: TanzuEventType.VSPHERE_VC_AUTHENTICATED,
            payload: this.getFieldValue(VsphereField.PROVIDER_VCENTER_ADDRESS)
        });

        if (this.hasPacific === 'yes') {
            this.vSphereModalTitle = `vSphere ${this.vsphereVersion} with Tanzu Detected`;
            this.vSphereModalBody = `You have connected to a vSphere ${this.vsphereVersion} with Tanzu
                                environment that includes an integrated Tanzu Kubernetes Grid Service which turns a
                                vSphere cluster into a platform for running Kubernetes workloads in dedicated resource
                                pools. Configuring Tanzu Kubernetes Grid Service is done through the vSphere HTML5 Client.`;
        } else {
            this.vSphereModalTitle = `vSphere ${this.vsphereVersion} Environment Detected`;
            this.vSphereModalBody = `You have connected to a vSphere ${this.vsphereVersion} environment
                                which does not have vSphere with Tanzu enabled. vSphere with Tanzu includes an
                                integrated Tanzu Kubernetes Grid Service which turns a vSphere cluster into a platform
                                for running Kubernetes workloads in dedicated resource pools. Configuring Tanzu Kubernetes
                                Grid Service is done through the vSphere HTML5 Client.`;
        }
    }

    /**
     * @method retrieveDatacenters
     * helper method to retrieve list of available datacenters from connected VC environment
     */
    retrieveDatacenters() {
        this.apiClient.getVSphereDatacenters()
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((res) => {
                this.datacenters = sortPaths(res, function (item) { return item.name; }, '/');
                this.formGroup.get(VsphereField.PROVIDER_DATA_CENTER).enable();
                this.formGroup.get(VsphereField.PROVIDER_SSH_KEY).enable();
                this.formGroup.get(VsphereField.PROVIDER_DATA_CENTER).setValue(this.getSavedValue(VsphereField.PROVIDER_DATA_CENTER, ''));
                if (this.datacenters.length === 1) {
                    this.formGroup.get(VsphereField.PROVIDER_DATA_CENTER).setValue(this.datacenters[0].name);
                }
            },
            (err) => {
                const error = err.error.message || err.message || JSON.stringify(err);
                this.errorNotification =
                    `Failed to retrieve list of datacenters from the specified vCenter Server. ${error}`;
            }
        );
    }

    /**
     * @method dcOnChange
     * helper method to emit datacenter selection value when changed
     * triggers all subsequent vsphere API discovery calls dependant on datacenter moid
     * @param $event {string} datacenter moid value emitted from select change
     */
    dcOnChange(nameDatacenter: string) {
        if (this.datacenters) {
            let dcMoid = '';
            const datacenter = this.datacenters.find(dc => dc.name === nameDatacenter);
            if (datacenter && datacenter.moid) {
                dcMoid = datacenter.moid;
            }
            AppServices.messenger.publish({
                type: TanzuEventType.VSPHERE_DATACENTER_CHANGED,
                payload: dcMoid
            });
        }
    }

    /**
     * @method getDisabled
     * helper method to get if connect btn should be disabled
     */
    getDisabled(): boolean {
        return !this.controlIsValid(VsphereField.PROVIDER_USER_NAME) ||
            !this.controlIsValid(VsphereField.PROVIDER_USER_PASSWORD) ||
            !this.controlIsValid(VsphereField.PROVIDER_VCENTER_ADDRESS);
    }

    private controlIsValid(controlName: VsphereField): boolean {
        if (!this.formGroup) { return false; }
        const control = this.formGroup.get(controlName);
        return control && control.valid;
    }

    /**
     * @method launchVsphereWcp
     * @desc helper method to launch vSphere wcp enablement workflow in new window
     */
    launchVsphereWcp() {
        window.open(`https://${this.vsphereHost}/ui/app/workload-platform/`, '_blank');
    }

    /**
     * @method showVSphereWithK8Modal
     * @desc helper method to open vSphere with K8's modal
     */
    showVSphereWithK8Modal() {
        this.vSphereWithK8ModalOpen = true;
    }

    /**
     * @method onFileChanged
     * `change` event handler for the file input.
     * @param event
     */
    onFileChanged(event) {
        if (event.target.files.length) {
            this.fileReader.readAsText(event.target.files[0]);
            // clear file reader target so user can re-select same file if needed
            event.target.value = '';
        }
    }

    dynamicDescription(): string {
        const vcenterIP = this.getFieldValue(VsphereField.PROVIDER_VCENTER_ADDRESS, true);
        const datacenter = this.getFieldValue(VsphereField.PROVIDER_DATA_CENTER, true);
        if (vcenterIP && datacenter) {
            return 'vCenter ' + vcenterIP + ' connected';
        }
        const version = this.vsphereVersion ? this.vsphereVersion + ' ' : '';
        return 'Validate the vSphere ' + version + 'provider account for Tanzu';
    }

    protected storeUserData() {
        this.storeUserDataFromMapping(VsphereProviderStepFieldMapping);
        this.storeDefaultDisplayOrder(VsphereProviderStepFieldMapping);
    }
}
