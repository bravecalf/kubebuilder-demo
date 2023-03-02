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

package v1

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var foolog = logf.Log.WithName("foo-resource")

func (r *Foo) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-myapp-my-domain-v1-foo,mutating=true,failurePolicy=fail,sideEffects=None,groups=myapp.my.domain,resources=foos,verbs=create;update,versions=v1,name=mfoo.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Foo{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Foo) Default() {
	foolog.Info("default", "name", r.Name)
	// TODO(user): fill in your defaulting logic.
	// 未设置foo的副本数，默认副本数为1
	if r.Spec.Replicas == 0 {
		r.Spec.Replicas = 1
	}

	// 初始化状态需要将foo的可用副本数置为0
	if r.Status.AvailableReplicas > 0 {
		r.Status.AvailableReplicas = 0
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-myapp-my-domain-v1-foo,mutating=false,failurePolicy=fail,sideEffects=None,groups=myapp.my.domain,resources=foos,verbs=create;update,versions=v1,name=vfoo.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Foo{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Foo) ValidateCreate() error {
	foolog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return r.validateFooSpec()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Foo) ValidateUpdate(old runtime.Object) error {
	foolog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return r.validateFooSpec()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Foo) ValidateDelete() error {
	foolog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *Foo) validateFooSpec() error {
	// The field helpers from the kubernetes API machinery help us return nicely
	// structured validation errors.
	if !r.Spec.EnableService && r.Spec.EnableIngress {
		return errors.NewInvalid(GroupVersion.WithKind("Foo").GroupKind(), r.Name, field.ErrorList{
			field.Invalid(field.NewPath("enable_service"), r.Spec.EnableService, "enable_service should be true when enable_ingress is true"),
		})
	}
	return nil
}
