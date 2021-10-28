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
        private formBuilder: FormBuilder,
        titleService: Title,
        private apiClient: APIClient
    ) {
        super(router, el, formMetaDataService, titleService);
        this.buildForm();
    }

    ngOnInit(): void {
        super.ngOnInit();

        // To avoid re-open issue for the first step.
        this.form.markAsDirty();

        this.titleService.setTitle(this.title + ' Docker');
    }

    buildForm() {
        this.form = this.formBuilder.group({
            dockerDaemonForm: this.formBuilder.group({
            }),
            networkForm: this.formBuilder.group({
            }),
            dockerNodeSettingForm: this.formBuilder.group({
            })
        });
    }

    getStepDescription(stepName: string): string {
        if (stepName === 'network') {
            if (this.getFieldValue('networkForm', 'clusterPodCidr')) {
                return 'Cluster Pod CIDR: ' + this.getFieldValue('networkForm', 'clusterPodCidr');
            } else {
                return 'Specify the cluster Pod CIDR';
            }
        } else if (stepName === 'nodeSetting') {
            return 'Optional: Specify the management cluster name'
        }
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
}
