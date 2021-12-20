/**
 * Angular Modules
 */
import { Component, OnInit, ViewChild } from '@angular/core';
import { Router } from '@angular/router';
import { FormControl, Validators } from '@angular/forms';
import { ClrLoadingState } from '@clr/angular';
import { debounceTime, distinctUntilChanged, finalize, takeUntil } from 'rxjs/operators';
import * as _ from 'lodash';

/**
 * App imports
 */
import { APP_ROUTES, Routes } from 'src/app/shared/constants/routes.constants';
import { APIClient } from 'src/app/swagger/api-client.service';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { VSphereDatacenter } from 'src/app/swagger/models/v-sphere-datacenter.model';
import { ValidatorEnum } from '../../wizard/shared/constants/validation.constants';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { TkgEvent, TkgEventType } from 'src/app/shared/service/Messenger';
import { SSLThumbprintModalComponent } from '../../wizard/shared/components/modals/ssl-thumbprint-modal/ssl-thumbprint-modal.component';
import { FormMetaDataStore } from '../../wizard/shared/FormMetaDataStore';
import Broker from 'src/app/shared/service/broker';
import { EditionData } from 'src/app/shared/service/branding.service';
import { AppEdition } from 'src/app/shared/constants/branding.constants';
import { managementClusterPlugin } from "../../wizard/shared/constants/wizard.constants";
import { VsphereField } from "../vsphere-wizard.constants";
import { IpFamilyEnum } from "../../../../shared/constants/app.constants";
import { NotificationTypes } from "../../../../shared/components/alert-notification/alert-notification.component";
import { FormUtils } from '../../wizard/shared/utils/form-utils';

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
    vsphereVersion: string;
    vsphereHost: string;
    hasPacific: string;

    vSphereModalTitle: string;
    vSphereModalBody: string;
    thumbprint: string;

    edition: AppEdition = AppEdition.TCE;
    enableIpv6: boolean = false;

    constructor(private validationService: ValidationService,
        private apiClient: APIClient,
        private router: Router) {
        super();

        this.fileReader = new FileReader();
    }

    ngOnInit() {
        super.ngOnInit();
        this.enableIpv6 = Broker.appDataService.isPluginFeatureActivated(managementClusterPlugin, 'vsphereIPv6');
        FormUtils.addControl(
            this.formGroup,
            VsphereField.PROVIDER_IP_FAMILY,
            new FormControl( IpFamilyEnum.IPv4, [])
        );
        FormUtils.addControl(
            this.formGroup,
            VsphereField.PROVIDER_VCENTER_ADDRESS,
            new FormControl('', [
                Validators.required,
                this.validationService.isValidIpOrFqdn()
            ])
        );
        FormUtils.addControl(
            this.formGroup,
            VsphereField.PROVIDER_USER_NAME,
            new FormControl('', [
                Validators.required
            ])
        );
        FormUtils.addControl(
            this.formGroup,
            VsphereField.PROVIDER_USER_PASSWORD,
            new FormControl('', [
                Validators.required
            ])
        );
        FormUtils.addControl(
            this.formGroup,
            VsphereField.PROVIDER_CONNECTION_INSECURE,
            new FormControl(false, [])
        );
        FormUtils.addControl(
            this.formGroup,
            VsphereField.PROVIDER_DATA_CENTER,
            new FormControl('', [
                Validators.required
            ])
        );
        FormUtils.addControl(
            this.formGroup,
            VsphereField.PROVIDER_SSH_KEY,
            new FormControl('', [
                Validators.required
            ])
        );
        FormUtils.addControl(
            this.formGroup,
            VsphereField.PROVIDER_SSH_KEY_FILE,
            new FormControl('', [])
        );

        FormUtils.addControl(
            this.formGroup,
            VsphereField.PROVIDER_THUMBPRINT,
            new FormControl('', [])
        );

        this.formGroup.setValidators((data: any) => {
            if (data.controls.datacenter.value) {
                return null;
            } else {
                return { [ValidatorEnum.REQUIRED]: true };
            }
        });
        this.formGroup.get(VsphereField.PROVIDER_DATA_CENTER).disable({ emitEvent: false});
        this.datacenters = [];
        this.formGroup.get(VsphereField.PROVIDER_SSH_KEY).disable({ emitEvent: false});

        SupervisedField.forEach(field => {
            this.formGroup.get(field).valueChanges
                .pipe(
                    debounceTime(500),
                    distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
                    takeUntil(this.unsubscribe)
                )
                .subscribe(() => {
                    this.disconnect();
                });
        });

        this.formGroup.get(VsphereField.PROVIDER_DATA_CENTER).valueChanges.subscribe(data => {
            this.dcOnChange(data)
        });

        this.formGroup.get(VsphereField.PROVIDER_IP_FAMILY).valueChanges.subscribe(data => {
            Broker.messenger.publish({
                type: TkgEventType.IP_FAMILY_CHANGE,
                payload: data
            });
            this.disconnect();
        });

        Broker.messenger.getSubject(TkgEventType.BRANDING_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data: TkgEvent) => {
                const content: EditionData = data.payload;
                this.edition = content.edition;
            });

        Broker.messenger.getSubject(TkgEventType.CONFIG_FILE_IMPORTED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data: TkgEvent) => {
                this.configFileNotification = {
                    notificationType: NotificationTypes.SUCCESS,
                    message: data.payload
                };
                // The file import saves the data to local storage, so we reinitialize this step's form from there
                this.savedMetadata = FormMetaDataStore.getMetaData(this.formName);
                this.initFormWithSavedData();

                // Clear event so that listeners in other provider workflows do not receive false notifications
                Broker.messenger.clearEvent(TkgEventType.CONFIG_FILE_IMPORTED);
            });

        this.fileReader.onload = (event) => {
            try {
                this.formGroup.get(VsphereField.PROVIDER_SSH_KEY).setValue(event.target.result);
                console.log(event.target.result);
            } catch (error) {
                console.log(error.message);
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
        this.initFormWithSavedData();
    }
    disconnect() {
        this.connected = false;
        this.loadingState = ClrLoadingState.DEFAULT;
        this.formGroup.markAsPending(); // a temperary fix to ignore this.formGroup.statusChanges detection
        this.formGroup.get(VsphereField.PROVIDER_DATA_CENTER).setValue('');
        this.datacenters = [];
        this.formGroup.get(VsphereField.PROVIDER_DATA_CENTER).disable();
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
        this.vsphereHost = this.formGroup.controls[VsphereField.PROVIDER_VCENTER_ADDRESS].value;

        if (this.formGroup.controls[VsphereField.PROVIDER_CONNECTION_INSECURE].value) {
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
            (res) => {
                const vsphereVerInfo: VsphereVersioninfo = this.getParsedVsphereVer(res.version);
                const isCompatible: boolean = this.checkVsphereCompatible(vsphereVerInfo);

                this.errorNotification = '';
                this.connected = true;
                this.vsphereVersion = vsphereVerInfo.version;
                this.hasPacific = res.hasPacific;
                Broker.appDataService.setVsphereVersion(vsphereVerInfo.version);

                if (isCompatible && !(_.startsWith(this.vsphereVersion, '6'))
                    && this.edition !== AppEdition.TCE) {
                    // for 7 and newer and other potential anomolies, show modal suggesting upgrade
                    this.showVSphereWithK8Modal();
                } else if (!isCompatible) {
                    // route to vsphere not compatible
                    this.router.navigate([this.APP_ROUTES.INCOMPATIBLE]);
                }
                this.retrieveDatacenters();

                Broker.messenger.publish({
                    type: TkgEventType.VC_AUTHENTICATED,
                    payload: this.formGroup.controls[VsphereField.PROVIDER_VCENTER_ADDRESS].value
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
            },
            (err) => {
                const error = err.error.message || err.message || JSON.stringify(err);
                this.errorNotification =
                    `Failed to connect to the specified vCenter Server. ${error}`;
            }
        );
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
    dcOnChange(datacenter: string) {
        const dcMoid = this.datacenters.find(dc => dc.name === datacenter)
            && this.datacenters.find(dc => dc.name === datacenter).moid || "";
        Broker.messenger.publish({
            type: TkgEventType.DATACENTER_CHANGED,
            payload: dcMoid
        });
    }

    /**
     * @method getDisabled
     * helper method to get if connect btn should be disabled
     */
    getDisabled(): boolean {
        return !(this.formGroup.get(VsphereField.PROVIDER_VCENTER_ADDRESS).valid &&
            this.formGroup.get(VsphereField.PROVIDER_USER_NAME).valid &&
            this.formGroup.get(VsphereField.PROVIDER_USER_PASSWORD).valid);
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
}
