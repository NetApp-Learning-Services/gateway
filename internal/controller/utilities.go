//From:  https://github.com/nheidloff/operator-sample-go/blob/1be830ea555c6401d85c3677ddcfd2ecf83fd601/operator-application/utilities/conditions.go

package controller

import (
	"context"
	"fmt"
	"time"

	gatewayv1alpha2 "gateway/api/v1alpha2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ConditionsAware interface {
	GetConditions() []metav1.Condition
	SetConditions(conditions []metav1.Condition)
}

func (reconciler *StorageVirtualMachineReconciler) containsCondition(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, reason string) bool {

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
			errMessage := "custom resource status update failed"
			log.Error(err, errMessage)
			return fmt.Errorf(errMessage)
		}

	} else {
		errMessage := "status cannot be set, custom resource doesn't support conditions"
		log.Info(errMessage)
		return fmt.Errorf(errMessage)
	}
	return nil
}

func (reconciler *StorageVirtualMachineReconciler) deleteCondition(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine,
	typeName string, reason string) error {

	log := log.FromContext(ctx)
	var newConditions = make([]metav1.Condition, 0)
	for _, condition := range svmCR.Status.Conditions {
		if condition.Type != typeName && condition.Reason != reason {
			newConditions = append(newConditions, condition)
		}
	}
	svmCR.Status.Conditions = newConditions

	err := reconciler.Client.Status().Update(ctx, svmCR)
	if err != nil {
		log.Error(err, "Deleting the condition in the custom resource failed")
	}
	return nil
}
