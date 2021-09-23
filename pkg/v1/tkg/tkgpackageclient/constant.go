package tkgpackageclient

const (
	msgRunPackageInstalledDelete = "\n\nPlease consider using 'tanzu package installed delete' to delete the already created associated resources\n"
	msgRunPackageInstalledUpdate = "\n\nPlease consider using 'tanzu package installed update' to update the installed package with correct settings\n"

	defaultImageTag = "latest"
	defaultImageTagConstraint = ">0.0.0"
	kindCRDFullName = "CustomResourceDefinition"
	packageRepositoryCRDName = "packagerepositories.packaging.carvel.dev"
	packageRepositoryTagSelectionJSONPath = ".spec.versions[0].schema.openAPIV3Schema.properties.spec.properties.fetch.properties.imgpkgBundle.properties.tagSelection"
)