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
        private titleService: Title,
        private apiClient: APIClient
    ) {
        super(router, el, formMetaDataService);
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
            })
            // identityForm: this.formBuilder.group({
            // })
        });
    }

    getStepDescription(stepName: string): string {
        if (stepName === 'network') {
            if (this.getFieldValue('networkForm', 'clusterPodCidr')) {
                return 'Cluster Pod CIDR: ' + this.getFieldValue('networkForm', 'clusterPodCidr');
            } else {
                return 'Specify the cluster Pod CIDR';
            }
        }

        // else if (stepName === 'identity') {
        //     if (this.getFieldValue('identityForm', 'identityType') === 'oidc' &&
        //         this.getFieldValue('identityForm', 'issuerURL')) {
        //         return 'OIDC configured: ' + this.getFieldValue('identityForm', 'issuerURL')
        //     } else if (this.getFieldValue('identityForm', 'identityType') === 'ldap' &&
        //         this.getFieldValue('identityForm', 'endpointIp')) {
        //         return 'LDAP configured: ' + this.getFieldValue('identityForm', 'endpointIp') + ':' +
        //         this.getFieldValue('identityForm', 'endpointPort');
        //     } else {
        //         return 'Specify identity management'
        //     }
        // }
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

        // let ldap_url = '';
        // if (this.getFieldValue('identityForm', 'endpointIp')) {
        //     ldap_url = this.getFieldValue('identityForm', 'endpointIp') +
        //         ':' + this.getFieldValue('identityForm', 'endpointPort');
        // }

        payload.identityManagement = {
            // 'idm_type': this.getFieldValue('identityForm', 'identityType') || 'none'
            'idm_type': 'none'
        }

        // if (this.getFieldValue('identityForm', 'identityType') === 'oidc') {
        //     payload.identityManagement = Object.assign({
        //             'oidc_provider_name': '',
        //             'oidc_provider_url': this.getFieldValue('identityForm', 'issuerURL'),
        //             'oidc_client_id': this.getFieldValue('identityForm', 'clientId'),
        //             'oidc_client_secret': this.getFieldValue('identityForm', 'clientSecret'),
        //             'oidc_scope': this.getFieldValue('identityForm', 'scopes'),
        //             'oidc_claim_mappings': {
        //                 'username': this.getFieldValue('identityForm', 'oidcUsernameClaim'),
        //                 'groups': this.getFieldValue('identityForm', 'oidcGroupsClaim')
        //             }
        //
        //         }
        //         , payload.identityManagement);
        // } else if (this.getFieldValue('identityForm', 'identityType') === 'ldap') {
        //     payload.identityManagement = Object.assign({
        //             'ldap_url': ldap_url,
        //             'ldap_bind_dn': this.getFieldValue('identityForm', 'bindDN'),
        //             'ldap_bind_password': this.getFieldValue('identityForm', 'bindPW'),
        //             'ldap_user_search_base_dn': this.getFieldValue('identityForm', 'userSearchBaseDN'),
        //             'ldap_user_search_filter': this.getFieldValue('identityForm', 'userSearchFilter'),
        //             'ldap_user_search_username': this.getFieldValue('identityForm', 'userSearchUsername'),
        //             'ldap_user_search_name_attr': this.getFieldValue('identityForm', 'userSearchUsername'),
        //             'ldap_group_search_base_dn': this.getFieldValue('identityForm', 'groupSearchBaseDN'),
        //             'ldap_group_search_filter': this.getFieldValue('identityForm', 'groupSearchFilter'),
        //             'ldap_group_search_user_attr': this.getFieldValue('identityForm', 'groupSearchUserAttr'),
        //             'ldap_group_search_group_attr': this.getFieldValue('identityForm', 'groupSearchGroupAttr'),
        //             'ldap_group_search_name_attr': this.getFieldValue('identityForm', 'groupSearchNameAttr'),
        //             'ldap_root_ca': this.getFieldValue('identityForm', 'ldapRootCAData')
        //         }
        //         , payload.identityManagement);
        // }

        return payload;
    }

    applyTkgConfig(): Observable<ConfigFileInfo> {
        return this.apiClient.applyTKGConfigForDocker({ params: this.getPayload() });
    }

    getCli(configPath: string): string {
        const cliG = new CliGenerator();
        const cliParams: CliFields = {
            configPath: configPath,
            clusterType: this.clusterType
        };
        return cliG.getCli(cliParams);
    }

    createRegionalCluster(payload: any): Observable<any> {
        return this.apiClient.createDockerRegionalCluster(payload);
    }
}
