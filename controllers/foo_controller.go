/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"kubebuilder-demo/controllers/utils"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	myappv1 "kubebuilder-demo/api/v1"
)

// FooReconciler reconciles a Foo object
type FooReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logr.Logger
}

//+kubebuilder:rbac:groups=myapp.my.domain,resources=foos,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=myapp.my.domain,resources=foos/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=myapp.my.domain,resources=foos/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Foo object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *FooReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = log.FromContext(ctx)

	// TODO(user): your logic here
	foo := new(myappv1.Foo)
	err := r.Get(ctx, req.NamespacedName, foo)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// sync deployment
	err = r.handleDeployment(ctx, foo)
	if err != nil {
		r.logger.Error(err, "Failed to handle deployment")
		return ctrl.Result{}, err
	}

	//sync service
	err = r.handleService(ctx, foo)
	if err != nil {
		r.logger.Error(err, "Failed to handle service")
		return ctrl.Result{}, err
	}

	// sync ingress
	err = r.handleIngress(ctx, foo)
	if err != nil {
		r.logger.Error(err, "Failed to handle ingress")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *FooReconciler) handleDeployment(ctx context.Context, foo *myappv1.Foo) error {
	// process deployment
	curDeployment := new(appsv1.Deployment)
	err := r.Get(ctx, types.NamespacedName{Name: foo.Name, Namespace: foo.Namespace}, curDeployment)
	if err != nil && !errors.IsNotFound(err) {
		r.logger.Error(err, "Failed to get deployment from cache Reader.")
		return err
	}

	expectDeployment := utils.ConstructDeployment(foo)
	// add owner reference
	errOwnerRef := controllerutil.SetControllerReference(foo, expectDeployment, r.Scheme)
	if errOwnerRef != nil {
		r.logger.Error(errOwnerRef, "Failed to set controller reference.", "owner", foo, "controlled", expectDeployment)
		return errOwnerRef
	}

	if errors.IsNotFound(err) {
		// create deployment
		errCreate := r.Create(ctx, expectDeployment)
		if errCreate != nil && !errors.IsAlreadyExists(errCreate) {
			r.logger.Error(errCreate, "Failed to create deployment.")
			return errCreate
		}
	} else {
		if *curDeployment.Spec.Replicas != foo.Spec.Replicas {
			// update deployment replicas
			errUpdate := r.Update(ctx, expectDeployment)
			if errUpdate != nil {
				r.logger.Error(errUpdate, "Failed to update deployment.")
				return errUpdate
			}
		}
		if curDeployment.Status.AvailableReplicas != foo.Status.AvailableReplicas {
			// update availableReplicas of Foo Status
			fooCopy := foo.DeepCopy()
			fooCopy.Status.AvailableReplicas = curDeployment.Status.AvailableReplicas
			errUpdate := r.Status().Update(ctx, fooCopy)
			if errUpdate != nil {
				r.logger.Error(errUpdate, "Failed to update foo.")
				return errUpdate
			}
		}
	}
	return nil
}

func (r *FooReconciler) handleService(ctx context.Context, foo *myappv1.Foo) error {
	//process service
	curService := new(corev1.Service)
	err := r.Get(ctx, types.NamespacedName{Name: foo.Name, Namespace: foo.Namespace}, curService)
	if err != nil && !errors.IsNotFound(err) {
		r.logger.Error(err, "Failed to get service from cache Reader.")
		return err
	}

	expectService := utils.ConstructService(foo)
	errControllerRef := controllerutil.SetControllerReference(foo, expectService, r.Scheme)
	if errControllerRef != nil {
		r.logger.Error(errControllerRef, "Failed to set controller reference to service.")
		return errControllerRef
	}

	if errors.IsNotFound(err) {
		if foo.Spec.EnableService {
			// create service
			err := r.Create(ctx, expectService)
			if err != nil && !errors.IsAlreadyExists(err) {
				r.logger.Error(err, "Failed to create service.")
				return err
			}
		}
		return nil
	}

	if !foo.Spec.EnableService {
		err := r.Delete(ctx, curService)
		if err != nil {
			r.logger.Error(err, "Failed to delete service.")
			return err
		}
	} else {
		if curService.Spec.Ports[0].TargetPort.IntVal != expectService.Spec.Ports[0].TargetPort.IntVal {
			err := r.Update(ctx, expectService)
			if err != nil {
				r.logger.Error(err, "Failed to update service.")
				return err
			}
		}
	}

	return nil
}

func (r *FooReconciler) handleIngress(ctx context.Context, foo *myappv1.Foo) error {
	//process ingress
	curIngress := new(networkingv1.Ingress)
	err := r.Get(ctx, types.NamespacedName{Name: foo.Name, Namespace: foo.Namespace}, curIngress)
	if err != nil && !errors.IsNotFound(err) {
		r.logger.Error(err, "Failed to get ingress from cache Reader.")
		return err
	}

	expectIngress := utils.ConstructIngress(foo)
	errControllerRef := controllerutil.SetControllerReference(foo, expectIngress, r.Scheme)
	if errControllerRef != nil {
		r.logger.Error(errControllerRef, "Failed to set controller reference to ingress.")
		return errControllerRef
	}

	if errors.IsNotFound(err) {
		if foo.Spec.EnableIngress {
			// create ingress
			err := r.Create(ctx, expectIngress)
			if err != nil && !errors.IsAlreadyExists(err) {
				r.logger.Error(err, "Failed to create ingress.")
				return err
			}
		}
		return nil
	}

	if !foo.Spec.EnableIngress {
		err := r.Delete(ctx, curIngress)
		if err != nil {
			r.logger.Error(err, "Failed to delete ingress.")
			return err
		}
	} else {
		if !reflect.DeepEqual(curIngress.Spec.Rules, expectIngress.Spec.Rules) {
			err := r.Update(ctx, expectIngress)
			if err != nil {
				r.logger.Error(err, "Failed to update ingress.")
				return err
			}
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FooReconciler) SetupWithManager(mgr ctrl.Manager) error {

	newBuilder := ctrl.NewControllerManagedBy(mgr)

	// config global label predicate
	newBuilder.WithEventFilter(NewFooGlobalPredicate())

	// config object
	newBuilder.For(&myappv1.Foo{}, NewFooOption())
	newBuilder.Owns(&appsv1.Deployment{})
	newBuilder.Owns(&corev1.Service{})
	newBuilder.Owns(&networkingv1.Ingress{})

	return newBuilder.Complete(r)
}
