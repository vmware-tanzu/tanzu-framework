package predicates

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/vmware-tanzu-private/core/addons/constants"
	addonsv1alpha1 "github.com/vmware-tanzu-private/core/apis/addons/v1alpha1"
	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// processApp returns true if app should be processed.
// App can be processed if it is has tanzu addon annotation
func processApp(o client.Object, log logr.Logger) bool {

	if !validateAppObject(o) {
		log.Info("Expected object type of App. Got object type", "actualType", fmt.Sprintf("%T", o))
		return true
	}

	app := o.(*kappctrl.App)

	if isATanzuAddon(app) {
		return true
	}

	log.V(7).Info("App is not a TKG addon", constants.AddonNamespaceLogKey, app.Namespace, constants.AddonNamespaceLogKey, app.Name)

	return false
}

// processAppUpdate returns true if app update should be processed
// TODO: Update this once we figure out how to retrieve app status
func processAppUpdate(old client.Object, new client.Object, log logr.Logger) bool {
	if !validateAppObject(old) {
		log.Info("Expected old object type of App. Got object type", "actualType", fmt.Sprintf("%T", old))
		return true
	}

	if !validateAppObject(new) {
		log.Info("Expected new object type of App. Got object type", "actualType", fmt.Sprintf("%T", new))
		return true
	}

	newApp := old.(*kappctrl.App)

	if isATanzuAddon(newApp) && isSuccessOrFailedReconcile(newApp) {
		return true
	}

	log.V(7).Info("App is not a TKG addon or status hasnt changed", constants.AddonNamespaceLogKey, newApp.Namespace, constants.AddonNamespaceLogKey, newApp.Name)

	return false
}

func validateAppObject(o client.Object) bool {
	switch o.(type) {
	case *kappctrl.App:
		return true
	default:
		return false
	}
}

// isATanzuAddon returns true if app has tanzu addon annotation
func isATanzuAddon(app *kappctrl.App) bool {
	if _, ok := app.Annotations[addonsv1alpha1.AddonTypeAnnotation]; ok {
		return true
	}
	return false
}

// isSuccessOrFailedReconcile returns true if app reconciliation succeed or failed
func isSuccessOrFailedReconcile(app *kappctrl.App) bool {
	var toBeProcessed bool
	for _, condition := range app.Status.Conditions {
		if condition.Type == kappctrl.ReconcileFailed ||
			condition.Type == kappctrl.ReconcileSucceeded {
			toBeProcessed = true
			break
		}
	}
	return toBeProcessed
}

// App returns a predicate.Predicate that filters app resource
func App(log logr.Logger) predicate.Funcs {
	return predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return false },
		UpdateFunc:  func(e event.UpdateEvent) bool { return false },
		DeleteFunc:  func(e event.DeleteEvent) bool { return processApp(e.Object, log) },
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}
}
