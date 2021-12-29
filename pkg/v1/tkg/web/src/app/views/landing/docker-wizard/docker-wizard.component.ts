import { Component, ElementRef, OnInit } from '@angular/core';
import { FormBuilder } from '@angular/forms';
import { Router } from '@angular/router';
import { Title } from '@angular/platform-browser';

import { Observable } from 'rxjs';

import { FormMetaDataService } from 'src/app/shared/service/form-meta-data.service';
import { APIClient } from 'src/app/swagger';
import { ConfigFileInfo, DockerRegionalClusterParams } from 'src/app/swagger/models';
import { CliFields, CliGenerator } from '../wizard/shared/utils/cli-generator';
import { WizardBaseDirective } from '../wizard/shared/wizard-base/wizard-base';
import { ImportParams, ImportService } from "../../../shared/service/import.service";
import { FormDataForHTML, FormUtility } from '../wizard/shared/components/steps/form-utility';
import { NodeSettingStepComponent } from './node-setting-step/node-setting-step.component';
import { DaemonValidationStepComponent } from './daemon-validation-step/daemon-validation-step.component';
import { DockerNetworkStepComponent } from './network-step/docker-network-step.component';

@Component({
    selector: 'app-docker-wizard',
    templateUrl: './docker-wizard.component.html',
    styleUrls: ['./docker-wizard.component.scss']
})
export class DockerWizardComponent extends WizardBaseDirective implements OnInit {

    constructor(
        router: Router,
        el: ElementRef,
        formMetaDataService: FormMetaDataService,
        private importService: ImportService,
        formBuilder: FormBuilder,
        titleService: Title,
        private apiClient: APIClient
    ) {
        super(router, el, formMetaDataService, titleService, formBuilder);
    }

    protected supplyWizardName(): string {
        return 'Docker Wizard';
    }

    protected supplyStepData(): FormDataForHTML[] {
        return [
            this.DockerDaemonForm,
            this.DockerNetworkForm,
            this.DockerNodeSettingForm
        ];
    }

    ngOnInit(): void {
        super.ngOnInit();

        // To avoid re-open issue for the first step.
        this.form.markAsDirty();

        this.titleService.setTitle(this.title + ' Docker');
    }

    setFromPayload(payload: DockerRegionalClusterParams) {
        this.setFieldValue('networkForm', 'networkName', payload.networking.networkName);
        this.setFieldValue('networkForm', 'clusterServiceCidr',  payload.networking.clusterServiceCIDR);
        this.setFieldValue('networkForm', 'clusterPodCidr',  payload.networking.clusterPodCIDR);
        this.setFieldValue('networkForm', 'cniType',  payload.networking.cniType);

        this.setFieldValue('dockerNodeSettingForm', 'clusterName', payload.clusterName);

        this.saveProxyFieldsFromPayload(payload);
    }

    getPayload() {
        const payload: DockerRegionalClusterParams = {}

        payload.networking = {
            networkName: this.getFieldValue('networkForm', 'networkName'),
            clusterDNSName: '',
            clusterNodeCIDR: '',
            clusterServiceCIDR: this.getFieldValue('networkForm', 'clusterServiceCidr'),
            clusterPodCIDR: this.getFieldValue('networkForm', 'clusterPodCidr'),
            cniType: this.getFieldValue('networkForm', 'cniType')
        };

        payload.clusterName = this.getFieldValue('dockerNodeSettingForm', 'clusterName');

        if (this.getFieldValue('networkForm', 'proxySettings')) {
            let proxySettingsMap = null;
            proxySettingsMap = [
                ['HTTPProxyURL', 'networkForm', 'httpProxyUrl'],
                ['HTTPProxyUsername', 'networkForm', 'httpProxyUsername'],
                ['HTTPProxyPassword', 'networkForm', 'httpProxyPassword'],
                ['noProxy', 'networkForm', 'noProxy']
            ];
            if (this.getFieldValue('networkForm', 'isSameAsHttp')) {
                proxySettingsMap = [
                    ...proxySettingsMap,
                    ['HTTPSProxyURL', 'networkForm', 'httpProxyUrl'],
                    ['HTTPSProxyUsername', 'networkForm', 'httpProxyUsername'],
                    ['HTTPSProxyPassword', 'networkForm', 'httpProxyPassword']
                ];
            } else {
                proxySettingsMap = [
                    ...proxySettingsMap,
                    ['HTTPSProxyURL', 'networkForm', 'httpsProxyUrl'],
                    ['HTTPSProxyUsername', 'networkForm', 'httpsProxyUsername'],
                    ['HTTPSProxyPassword', 'networkForm', 'httpsProxyPassword']
                ];
            }
            payload.networking.httpProxyConfiguration = {
                enabled: true
            };
            proxySettingsMap.forEach(attr => {
                let val = this.getFieldValue(attr[1], attr[2]);
                if (attr[0] === 'noProxy') {
                    val = val.replace(/\s/g, ''); // remove all spaces
                }
                payload.networking.httpProxyConfiguration[attr[0]] = val;
            });
        }

        payload.identityManagement = {
            'idm_type': 'none'
        }

        return payload;
    }

    applyTkgConfig(): Observable<ConfigFileInfo> {
        return this.apiClient.applyTKGConfigForDocker({ params: this.getPayload() });
    }

    /**
     * Return management/standalone cluster name
     */
    getMCName() {
        return this.getFieldValue('dockerNodeSettingForm', 'clusterName');
    }

    /**
     * Retrieve the config file from the backend and return as a string
     */
    retrieveExportFile() {
        return this.apiClient.exportTKGConfigForDocker({ params: this.getPayload() });
    }

    getCli(configPath: string): string {
        const cliG = new CliGenerator();
        const cliParams: CliFields = {
            configPath: configPath,
            clusterType: this.getClusterType(),
            clusterName: this.getMCName(),
            extendCliCmds: []
        };
        return cliG.getCli(cliParams);
    }

    createRegionalCluster(payload: any): Observable<any> {
        return this.apiClient.createDockerRegionalCluster(payload);
    }

    // FormDataForHTML methods
    //
    get DockerNetworkForm(): FormDataForHTML {
        return FormUtility.formWithOverrides(super.NetworkForm,
            { description: DockerNetworkStepComponent.description, clazz: DockerNetworkStepComponent });
    }
    get DockerNodeSettingForm(): FormDataForHTML {
        const title = FormUtility.titleCase(this.clusterTypeDescriptor) + ' Cluster Settings';
        return { name: 'dockerNodeSettings', title: title, description: 'Optional: Specify the management cluster name',
            i18n: { title: 'node setting step name', description: 'node setting step description' },
        clazz: NodeSettingStepComponent};
    }
    get DockerDaemonForm(): FormDataForHTML {
        return { name: 'dockerDaemonForm', title: 'Docker Prerequisites',
            description: 'Validate the local Docker daemon, allocated CPUs and Total Memory',
            i18n: {title: 'docker prerequisite step name', description: 'Docker prerequisite step description'},
        clazz: DaemonValidationStepComponent};
    }

    // returns TRUE if the file contents appear to be a valid config file for Docker
    // returns FALSE if the file is empty or does not appear to be valid. Note that in the FALSE
    // case we also alert the user.
    importFileValidate(nameFile: string, fileContents: string): boolean {
        if (fileContents.includes('INFRASTRUCTURE_PROVIDER: docker')) {
            return true;
        }
        alert(nameFile + ' is not a valid docker configuration file!');
        return false;
    }

    importFileRetrieveClusterParams(fileContents: string): Observable<DockerRegionalClusterParams> {
        return this.apiClient.importTKGConfigForVsphere( { params: { filecontents: fileContents } } );
    }

    importFileProcessClusterParams(nameFile: string, dockerClusterParams: DockerRegionalClusterParams) {
        this.setFromPayload(dockerClusterParams);
        this.resetToFirstStep();
        this.importService.publishImportSuccess(nameFile);
    }

    // returns TRUE if user (a) will not lose data on import, or (b) confirms it's OK
    onImportButtonClick() {
        let result = true;
        if (!this.isOnFirstStep()) {
            result = confirm('Importing will overwrite any data you have entered. Proceed with import?');
        }
        return result;
    }

    onImportFileSelected(event) {
        const params: ImportParams<DockerRegionalClusterParams> = {
            file: event.target.files[0],
            validator: this.importFileValidate,
            backend: this.importFileRetrieveClusterParams.bind(this),
            onSuccess: this.importFileProcessClusterParams.bind(this),
            onFailure: this.importService.publishImportFailure
        }
        this.importService.import(params);

        // clear file reader target so user can re-select same file if needed
        event.target.value = '';
    }
}
