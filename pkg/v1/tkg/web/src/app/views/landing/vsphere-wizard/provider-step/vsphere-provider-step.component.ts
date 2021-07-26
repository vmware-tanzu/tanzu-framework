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

declare var sortPaths: any;

const SupervisedField = ['vcenterAddress', 'username', 'password'];

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

    constructor(private validationService: ValidationService,
        private apiClient: APIClient,
        private router: Router) {
        super();
    }

    ngOnInit() {
        super.ngOnInit();
        this.formGroup.addControl(
            'vcenterAddress',
            new FormControl('', [
                Validators.required,
                this.validationService.isValidIpOrFqdn()
            ])
        );
        this.formGroup.addControl(
            'username',
            new FormControl('', [
                Validators.required
            ])
        );
        this.formGroup.addControl(
            'password',
            new FormControl('', [
                Validators.required
            ])
        );
        this.formGroup.addControl(
            'datacenter',
            new FormControl('', [
                Validators.required
            ])
        );
        this.formGroup.addControl(
            'ssh_key',
            new FormControl('', [
                Validators.required
            ])
        );

        this.formGroup.addControl(
            'thumbprint',
            new FormControl('', [])
        );

        this.formGroup.setValidators((data: any) => {
            if (data.controls.datacenter.value) {
                return null;
            } else {
                return { [ValidatorEnum.REQUIRED]: true };
            }
        });
        this.formGroup.get('datacenter').disable();
        this.datacenters = [];
        this.formGroup.get('ssh_key').disable();

        SupervisedField.forEach(field => {
            this.formGroup.get(field).valueChanges
                .pipe(
                    debounceTime(500),
                    distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
                    takeUntil(this.unsubscribe)
                )
                .subscribe(() => {
                    this.connected = false;
                    this.loadingState = ClrLoadingState.DEFAULT;
                    this.formGroup.get('datacenter').setValue('');
                    this.datacenters = [];
                    this.formGroup.get('datacenter').disable();
                });
        });

        this.formGroup.get('datacenter').valueChanges.subscribe(data => {
            this.dcOnChange(data)
        });

        Broker.messenger.getSubject(TkgEventType.BRANDING_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data: TkgEvent) => {
                const content: EditionData = data.payload;
                this.edition = content.edition;
            });
    }

    setSavedDataAfterLoad() {
        super.setSavedDataAfterLoad();
        // don't fill password field with ****
        this.formGroup.get('password').setValue('');
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
        this.vsphereHost = this.formGroup.controls['vcenterAddress'].value;

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
                } else {
                    this.thumbprint = thumbprint;
                    this.formGroup.controls['thumbprint'].setValue(thumbprint);
                    FormMetaDataStore.saveMetaDataEntry(this.formName, 'thumbprint', {
                        label: 'SSL THUMBPRINT',
                        displayValue: thumbprint
                    });
                    this.sslThumbprintModal.open();
                }
            },
            (err) => {
                const error = err.error.message || err.message || JSON.stringify(err);
                this.errorNotification =
                    `Failed to connect to the specified vCenter Server. ${error}`;
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
                username: this.formGroup.controls['username'].value,
                password: this.formGroup.controls['password'].value,
                host: this.formGroup.controls['vcenterAddress'].value,
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
                if (isCompatible && !(_.startsWith(this.vsphereVersion, '6'))
                    && this.edition !== AppEdition.TCE
                    && this.edition !== AppEdition.TCE_STANDALONE) {
                    // for 7 and newer and other potential anomolies, show modal suggesting upgrade
                    this.showVSphereWithK8Modal();
                } else if (!isCompatible) {
                    // route to vsphere not compatible
                    this.router.navigate([this.APP_ROUTES.INCOMPATIBLE]);
                }
                this.retrieveDatacenters();

                Broker.messenger.publish({
                    type: TkgEventType.VC_AUTHENTICATED,
                    payload: this.formGroup.controls['vcenterAddress'].value
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
                this.formGroup.get('datacenter').enable();
                this.formGroup.get('ssh_key').enable();
                this.formGroup.get('datacenter').setValue(this.getSavedValue('datacenter', ''));
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
        return !(this.formGroup.get('vcenterAddress').valid &&
            this.formGroup.get('username').valid &&
            this.formGroup.get('password').valid);
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
}
