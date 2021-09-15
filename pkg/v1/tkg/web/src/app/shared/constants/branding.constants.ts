export enum AppEdition {
    TKG = 'tkg',
    TCE = 'tce',
    TCE_STANDALONE = 'tce-standalone'
}

export const brandingDefault = {
    edition: AppEdition.TKG,
    clusterType: "management",
    branding: {
        title: "Tanzu Kubernetes Grid",
        landingPage: {
            logoClass: "tanzu-logo",
            title: "Welcome to the VMware Tanzu Kubernetes Grid Installer",
            intro: "VMware Tanzu Kubernetes Grid delivers the services that IT teams need to effectively support development " +
                "teams that develop and configure Kubernetes-based applications in a complex world. It balances the needs of " +
                "development teams to access resources and services, with the needs of centralized IT organizations to " +
                "maintain and control the development environments.<br\><br\>To begin using Tanzu Kubernetes Grid, you " +
                "first deploy a management cluster to your chosen infrastructure. The management cluster provides the entry " +
                "point for Tanzu Kubernetes Grid integration with your platform, and allows you to deploy multiple workload clusters." +
                "<br><br>Product documentation can be found <a href='https://docs.vmware.com/en/VMware-Tanzu-Kubernetes-Grid/index.html' " +
                "target='_blank'>here</a>."
        }
    }
}

export const brandingTce = {
    edition: AppEdition.TCE,
    clusterType: "management",
    branding: {
        title: "Tanzu Community Edition",
        landingPage: {
            logoClass: "tce-logo",
            title: "Welcome to the Tanzu Community Edition Installer",
            intro: "Tanzu Community Edition (TCE) is VMware's Open Source Kubernetes distribution. The installer will " +
                "deploy a temporary cluster on your local machine to bootstrap a management cluster on your desired target. " +
                "This management cluster can then be used to deploy and manage workload clusters.<br/><br/>For more details " +
                "see the <a href='http://tanzucommunityedition.io/docs' target='_blank'>getting started guide</a>."
        }
    }
}

export const brandingTceStandalone = {
    edition: AppEdition.TCE_STANDALONE,
    clusterType: "standalone",
    branding: {
        title: "Tanzu Community Edition",
        landingPage: {
            logoClass: "tce-logo",
            title: "Welcome to the Tanzu Community Edition Installer",
            intro: "Tanzu Community Edition (TCE) is VMware's Open Source Kubernetes distribution. The installer will " +
                "deploy a temporary cluster on your local machine to bootstrap a standalone cluster on your desired target. " +
                "<br/><br/>For more details see the <a href='http://tanzucommunityedition.io/docs' target='_blank'>getting started guide</a>."
        }
    }
}
