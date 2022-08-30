export interface Routes {
    [key: string]: string;
}

export interface Fragment {
    [key: string]: string;
}

export const APP_ROUTES: Routes = {
    LANDING: '/ui',
    VSPHERE_WITH_KUBERNETES: '/ui/vsphere-with-kubernetes',
    INCOMPATIBLE: '/ui/incompatible',
    WIZARD_MGMT_CLUSTER: '/ui/wizard',
    AWS_WIZARD: '/ui/aws/wizard',
    AZURE_WIZARD: '/ui/azure/wizard',
    DOCKER_WIZARD: '/ui/docker/wizard',
    WIZARD_PROGRESS: '/ui/deploy-progress'
};
