// From:  https://github.com/nheidloff/operator-sample-go/blob/main/operator-application/controllers/application/deployment.go

package controllers

import (
	"context"

	gatewayv1alpha1 "gateway/api/v1alpha1"
)

func (r *StorageVirtualMachineReconciler) reconcileSvmUpdate(ctx context.Context, svmCR *gatewayv1alpha1.StorageVirtualMachine) {

	// If SVM exists, then check to see if svmName needs to be updated
	//if svm.Name != svmCR.Spec.SvmName {
	//update name in ONTAP cluster
	//}

	// If SVM exists, then check to see if managment LIF needs to be created or updated

	// If SVM exists, then check to see if vsadmin needs to be created/updated

}
