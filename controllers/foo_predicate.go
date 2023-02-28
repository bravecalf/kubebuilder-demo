package controllers

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	PublicLabelKey   = "myapp.my.domain/use-for"
	PublicLabelValue = "kubebuilder-demo"

	// PrivateLabelKey 用于指定用户id,默认必须包含
	PrivateLabelKey = "myapp.my.domain/manager-by"
)

// NewFooGlobalPredicate add foo global labels selector predicate
func NewFooGlobalPredicate() predicate.Predicate {
	labelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			PublicLabelKey: PublicLabelValue,
		},
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      PrivateLabelKey,
				Operator: metav1.LabelSelectorOpExists,
			},
		},
	}
	labelSelectorPredicate, err := predicate.LabelSelectorPredicate(labelSelector)
	if err != nil {
		log.Fatalf("Failed to create label selector predictate, error: %v \n", err)
		return nil
	}
	return labelSelectorPredicate
}

// NewFooOption add foo private predicate
// Add/Update operation to do Reconcile
// Delete operation to be ignored
func NewFooOption() builder.Predicates {
	pred := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return !reflect.DeepEqual(e.ObjectNew, e.ObjectOld)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
	return builder.WithPredicates(pred)
}
