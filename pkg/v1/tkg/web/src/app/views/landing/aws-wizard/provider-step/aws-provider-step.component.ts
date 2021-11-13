/**
 * Angular Modules
 */
import { Component, OnInit } from '@angular/core';
import {
    FormControl,
    Validators
} from '@angular/forms';

import { takeUntil, debounceTime, distinctUntilChanged } from 'rxjs/operators';
import { forkJoin, of } from 'rxjs';
import { catchError } from 'rxjs/operators';

import { APIClient } from '../../../../swagger/api-client.service';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { TkgEventType } from '../../../../shared/service/Messenger';
import Broker from 'src/app/shared/service/broker';

export const AWSAccountParamsKeys = ['profileName', 'sessionToken', 'region', 'accessKeyID', 'secretAccessKey'];

@Component({
    selector: 'app-aws-provider-step',
    templateUrl: './aws-provider-step.component.html',
    styleUrls: ['./aws-provider-step.component.scss']
})
export class AwsProviderStepComponent extends StepFormDirective implements OnInit {

    loading = false;
    authTypeValue: string = 'oneTimeCredentials';

    regions = [];
    profileNames: Array<string> = [];
    validCredentials: boolean = false;
    isProfileChoosen: boolean = false;

    constructor(private apiClient: APIClient) {
        super();

        console.log('cluster type from stepform directive: ' + this.clusterTypeDescriptor);
    }

    /**
     * Create the initial form
     */
    private buildForm() {
        this.formGroup.addControl('authType', new FormControl('oneTimeCredentials', []));

        AWSAccountParamsKeys.forEach(key => this.formGroup.addControl(
            key,
            new FormControl('')
        ));

        this.formGroup.get('region').setValidators([
            Validators.required
        ]);

        this.formGroup['canMoveToNext'] = () => {
            return this.formGroup.valid && this.validCredentials;
        }
    }

    /**
     * Set the hidden form field to proper value based on form validity
     * @param valid whether we want the form to be valid
     */
    setValidCredentials(valid) {
        this.validCredentials = valid;

        // call to get existing VPCs of region
        if (valid === true) {
            Broker.messenger.publish({
                type: TkgEventType.AWS_GET_EXISTING_VPCS
            });

            Broker.messenger.publish({
                type: TkgEventType.AWS_GET_AVAILABILITY_ZONES
            });

            Broker.messenger.publish({
                type: TkgEventType.AWS_GET_NODE_TYPES
            });
        }
    }

    /**
     * Initialize the form with data from the backend
     * @param credentials AWS credentials
     * @param regions AWS regions
     */
    private initForm() {
        this.initAwsCredentials();
    }

    ngOnInit() {
        super.ngOnInit();

        this.loading = true;
        this.buildForm();
        this.initForm();

        this.formGroup.valueChanges
        .pipe(
            debounceTime(500),
            distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
            takeUntil(this.unsubscribe)
        )
        .subscribe(
            (data) => {
                this.trimCreds(data);
            }
        );

        this.formGroup.get('authType').valueChanges.pipe(
            distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
            takeUntil(this.unsubscribe)
        ).subscribe(data => {
            this.authTypeValue = data;

            if (this.authTypeValue === 'oneTimeCredentials') {
                this.oneTimeCredentialsSelectedHandler();
            } else if (this.authTypeValue === 'credentialProfile') {
                this.credentialProfileSelectedHandler();
            } else {
                this.disarmField('authType', true);
            }
        });
        this.authTypeValue = this.getSavedValue('authType', 'credentialProfile');

        this.formGroup.get('authType').setValue(this.authTypeValue);
    }

    trimCreds(data) {
        const trimmedId = data.accessKeyID && data.accessKeyID.replace('\t', '');
        const trimmedKey = data.secretAccessKey && data.secretAccessKey.replace('\t', '');

        if (trimmedId !== data.accessKeyID) {
            this.formGroup.get('accessKeyID').setValue(trimmedId);
        }

        if (trimmedKey !== data.secretAccessKey) {
            this.formGroup.get('secretAccessKey').setValue(trimmedKey);
        }

        this.validCredentials = false
    }

    initAwsCredentials() {
        const getRegionObs$ =  this.apiClient.getAWSRegions().pipe(
            catchError(err => of([])),
        );
        const getProfilesObs$ = this.apiClient.getAWSCredentialProfiles().pipe(
            catchError(err => of([])),
        );
        forkJoin([getRegionObs$, getProfilesObs$])
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(
                (next) => {
                    this.regions = next[0].sort();
                    this.profileNames = next[1];
                    if (this.profileNames.length === 1) {
                        this.formGroup.get('profileName').setValue(this.profileNames[0]);
                    }
                    if (this.regions.length === 1) {
                        this.formGroup.get('region').setValue(this.regions[0]);
                    }
                },
                () => this.loading = false
            );
    }

    oneTimeCredentialsSelectedHandler() {
        this.disarmField('profileName', true);
    }

    credentialProfileSelectedHandler() {
        const resetFields = ['accessKeyID', 'secretAccessKey', 'sessionToken'];
        resetFields.forEach(field => this.disarmField(field, true));
    }

    setAWSCredentialsValuesFromAPI(credentials) {
        // init form values for AWS credentials
        for (const key of AWSAccountParamsKeys) {
            this.formGroup.get(key).setValue(credentials[key]);
        }
    }

    setSavedDataAfterLoad() {
        // disabled saved data
        // super.setSavedDataAfterLoad();
        // don't fill password fields with ****
        this.formGroup.get('accessKeyID').setValue('');
        this.formGroup.get('secretAccessKey').setValue('');
    }

    /**
     * @method verifyCredentails
     * helper method to verify AWS connection credentials
     */
    verifyCredentails() {
        this.loading = true;
        this.errorNotification = '';
        const params = {};
        for (const key of AWSAccountParamsKeys) {
            params[key] = this.formGroup.get(key).value;
        }
        this.apiClient.setAWSEndpoint({
            accountParams: params
        })
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(
                (() => {
                    this.errorNotification = '';

                    // Notify the universe that region has changed.
                    Broker.messenger.publish({
                        type: TkgEventType.AWS_REGION_CHANGED,
                        payload: this.formGroup.get('region').value
                    });

                    Broker.messenger.publish({
                        type: TkgEventType.AWS_GET_OS_IMAGES,
                        payload: {
                            region: this.formGroup.get('region').value
                        }
                    });

                    this.setValidCredentials(true);
                }),
                ((err) => {
                    const error = err.error.message || err.message || JSON.stringify(err);
                    this.errorNotification = `Invalid AWS credentials: all credentials and region must be valid. ${error}`;
                    this.setValidCredentials(false)
                }),
                (() => {
                    this.loading = false;
                })
            );
    }

    /**
     * Whether to disable "Connect" button
     */
    isConnectDisabled() {
        return !AWSAccountParamsKeys.reduce((accu, key) => this.formGroup.get(key).valid && accu, true);
    }
}
