package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/inspect"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/tanzu-auth-controller-manager/pkg/pinnipedinfo"
)

type Controller struct {
	client.Client

	serviceNamespace, serviceName string

	k8sClientset kubernetes.Interface
}

func NewController(
	c client.Client,
	serviceNamespace, serviceName string,
	k8sClientset kubernetes.Interface,
) *Controller {
	return &Controller{
		Client: c,

		serviceNamespace: serviceNamespace,
		serviceName:      serviceName,

		k8sClientset: k8sClientset,
	}
}

func (c *Controller) SetupWithManager(manager ctrl.Manager) error {
	emptyReqMapFunc := handler.EnqueueRequestsFromMapFunc(func(_ client.Object) []ctrl.Request { return []ctrl.Request{{}} })
	return ctrl.
		NewControllerManagedBy(manager).
		Named("pinniped-status").
		For(
			&corev1.Service{},
			withNamespacedName(types.NamespacedName{Namespace: c.serviceNamespace, Name: c.serviceName}),
		).
		Watches(
			&source.Kind{Type: &corev1.ConfigMap{}},
			emptyReqMapFunc, // Since this is a singleton controller, we don't need to specify a req
			withNamespacedName(types.NamespacedName{
				Namespace: pinnipedinfo.ConfigMapNamespace,
				Name:      pinnipedinfo.ConfigMapName,
			}),
		).
		Complete(c)
}

func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// This is a singleton controller - set the req to the actual resource we are reconciling
	req.NamespacedName = types.NamespacedName{Namespace: c.serviceNamespace, Name: c.serviceName}

	// Get the service that triggered this reconcile.
	var s corev1.Service
	err := c.Get(ctx, req.NamespacedName, &s)
	notFound := k8serrors.IsNotFound(err)
	if err != nil && !notFound {
		return ctrl.Result{}, fmt.Errorf("get service: %w", err)
	}
	log.V(1).Info("got service", "exists", !notFound, "type", s.Spec.Type)

	var serviceEndpoint string
	if !notFound {
		// Get the service address to use for the supervisor.
		inspector := inspect.Inspector{K8sClientset: c.k8sClientset, Context: ctx}
		serviceEndpoint, err = inspector.GetServiceEndpoint(c.serviceNamespace, c.serviceName)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("get service endpiont: %w", err)
		}
		log.V(1).Info("got service endpoint", "endpoint", serviceEndpoint)
	}

	// Update the pinniped-info ConfigMap
	pinnipedInfoConfigMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: pinnipedinfo.ConfigMapNamespace,
			Name:      pinnipedinfo.ConfigMapName,
		},
	}
	if _, err := controllerutil.CreateOrPatch(
		ctx,
		c.Client,
		&pinnipedInfoConfigMap,
		func() error {
			if pinnipedInfoConfigMap.Data == nil {
				pinnipedInfoConfigMap.Data = make(map[string]string)
			}

			// If the service does not exist, then remove the issuer field in the CM to communicate this
			// to the reader; otherwise, update the CM's "issuer" data
			if len(serviceEndpoint) == 0 {
				delete(pinnipedInfoConfigMap.Data, pinnipedinfo.IssuerKey)
			} else {
				pinnipedInfoConfigMap.Data[pinnipedinfo.IssuerKey] = serviceEndpoint
			}

			return nil
		},
	); err != nil {
		return ctrl.Result{}, fmt.Errorf("create or patch pinniped-info ConfigMap: %w", err)
	}
	log.V(1).Info("pinniped-info updated", "data", pinnipedInfoConfigMap.Data)

	return ctrl.Result{}, nil
}

func withNamespacedName(namespacedName types.NamespacedName) builder.Predicates {
	isNamespacedName := func(o client.Object) bool {
		return o.GetNamespace() == namespacedName.Namespace && o.GetName() == namespacedName.Name
	}
	return builder.WithPredicates(
		predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool { return isNamespacedName(e.Object) },
			UpdateFunc: func(e event.UpdateEvent) bool {
				return isNamespacedName(e.ObjectOld) || isNamespacedName(e.ObjectNew)
			},
			DeleteFunc:  func(e event.DeleteEvent) bool { return isNamespacedName(e.Object) },
			GenericFunc: func(e event.GenericEvent) bool { return isNamespacedName(e.Object) },
		},
	)
}
