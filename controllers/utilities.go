package controllers

import (
	"context"
	"fmt"
	"time"

	gatewayv1alpha1 "gateway/api/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ConditionsAware interface {
	GetConditions() []metav1.Condition
	SetConditions(conditions []metav1.Condition)
}

func (reconciler *StorageVirtualMachineReconciler) containsCondition(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine, reason string) bool {

	output := false
	for _, condition := range svmCR.Status.Conditions {
		if condition.Reason == reason {
			output = true
		}
	}
	return output
}

func appendCondition(ctx context.Context, reconcilerClient client.Client, object client.Object,
	typeName string, status metav1.ConditionStatus, reason string, message string) error {
	log := log.FromContext(ctx)
	conditionsAware, conversionSuccessful := (object).(ConditionsAware)
	if conversionSuccessful {
		time := metav1.Time{Time: time.Now()}
		condition := metav1.Condition{Type: typeName, Status: status, Reason: reason, Message: message, LastTransitionTime: time}
		conditionsAware.SetConditions(append(conditionsAware.GetConditions(), condition))
		err := reconcilerClient.Status().Update(ctx, object)
		if err != nil {
			errMessage := "Custom resource status update failed"
			log.Info(errMessage)
			return fmt.Errorf(errMessage)
		}

	} else {
		errMessage := "Status cannot be set, resource doesn't support conditions"
		log.Info(errMessage)
		return fmt.Errorf(errMessage)
	}
	return nil
}
