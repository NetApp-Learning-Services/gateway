// From:  https://github.com/nheidloff/operator-sample-go/blob/main/operator-application/controllers/application/deployment.go

package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	gatewayv1alpha1 "gateway/api/v1alpha1"
	"gateway/ontap"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *StorageVirtualMachineReconciler) reconcileSecurityAccount(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine, oc *ontap.Client, credentials *corev1.Secret) (ctrl.Result, error) {

	log := log.FromContext(ctx)
	log.Info("reconcileSecurityAccount started")

	var payload ontap.SecurityAccountPayload
	payload.Name = string(credentials.Data["username"])
	payload.Owner.Uuid = svmCR.Spec.SvmUuid

	ap := ontap.Application{
		AppType:          ontap.Ssh,
		SecondAuthMethod: "none", //special word
	}
	ap.AuthMethods = append(ap.AuthMethods, ontap.Password)
	payload.Applications = append(payload.Applications, ap)
	payload.Role = ontap.Vsadmin
	payload.Password = string(credentials.Data["password"])

	log.Info("Security account payload: " + fmt.Sprintf("%#v\n", payload))

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		//error creating the json body
		log.Error(err, "Error creating the json payload for security account")
		return ctrl.Result{}, err
	}

	log.Info("Security account creation attempt")
	err = oc.CreateSecurityAccount(jsonPayload)
	if err != nil {
		log.Error(err, "Error occurred when creating security account")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
