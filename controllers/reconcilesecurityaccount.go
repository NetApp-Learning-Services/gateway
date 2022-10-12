// From:  https://github.com/nheidloff/operator-sample-go/blob/main/operator-application/controllers/application/deployment.go

package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	gatewayv1alpha1 "gateway/api/v1alpha1"
	"gateway/ontap"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *StorageVirtualMachineReconciler) reconcileSecurityAccount(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine, oc *ontap.Client, credentials *corev1.Secret) (ctrl.Result, error) {

	log := log.FromContext(ctx)
	log.Info("reconcileSecurityAccount started")

	if string(credentials.Data["username"]) == "vsadmin" {
		log.Info("vsadmin credentials - need to patch")
		var payload ontap.SecurityAccountPayload
		payload.Name = string(credentials.Data["username"])

		ssh := ontap.Application{
			AppType:          ontap.Ssh,
			SecondAuthMethod: "none", //special word
		}
		ssh.AuthMethods = append(ssh.AuthMethods, ontap.Password)
		payload.Applications = append(payload.Applications, ssh)
		ontapi := ontap.Application{
			AppType:          ontap.Ssh,
			SecondAuthMethod: "none", //special word
		}
		ontapi.AuthMethods = append(ontapi.AuthMethods, ontap.Password)
		payload.Applications = append(payload.Applications, ontapi)

		http := ontap.Application{
			AppType:          ontap.Ssh,
			SecondAuthMethod: "none", //special word
		}
		http.AuthMethods = append(http.AuthMethods, ontap.Password)
		payload.Applications = append(payload.Applications, http)

		payload.Role = ontap.Vsadmin
		payload.Password = string(credentials.Data["password"])
		payload.Locked = false

		log.Info("Security account payload: " + fmt.Sprintf("%#v\n", payload))

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			//error creating the json body
			log.Error(err, "Error creating the json payload for security account patch")
			return ctrl.Result{}, err
		}

		log.Info("Security account patch attempt")
		if svmCR.Spec.SvmUuid == "" {
			return ctrl.Result{}, errors.NewBadRequest("No SVM uuid during security account patch")
		}
		err = oc.PatchSecurityAccount(jsonPayload, svmCR.Spec.SvmUuid, payload.Name)
		if err != nil {
			log.Error(err, "Error occurred when patching security account")
			return ctrl.Result{}, err
		}

	} else {
		log.Info("not vsadmin credentials - try to create")
		var payload ontap.SecurityAccountPayload
		payload.Name = string(credentials.Data["username"])
		payload.Owner.Uuid = svmCR.Spec.SvmUuid

		ssh := ontap.Application{
			AppType:          ontap.Ssh,
			SecondAuthMethod: "none", //special word
		}
		ssh.AuthMethods = append(ssh.AuthMethods, ontap.Password)
		payload.Applications = append(payload.Applications, ssh)
		ontapi := ontap.Application{
			AppType:          ontap.Ssh,
			SecondAuthMethod: "none", //special word
		}
		ontapi.AuthMethods = append(ontapi.AuthMethods, ontap.Password)
		payload.Applications = append(payload.Applications, ontapi)

		http := ontap.Application{
			AppType:          ontap.Ssh,
			SecondAuthMethod: "none", //special word
		}
		http.AuthMethods = append(http.AuthMethods, ontap.Password)
		payload.Applications = append(payload.Applications, http)

		payload.Role = ontap.Vsadmin
		payload.Password = string(credentials.Data["password"])

		log.Info("Security account payload: " + fmt.Sprintf("%#v\n", payload))

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			//error creating the json body
			log.Error(err, "Error creating the json payload for security account creation")
			return ctrl.Result{}, err
		}

		log.Info("Security account creation attempt")
		err = oc.CreateSecurityAccount(jsonPayload)
		if err != nil {
			log.Error(err, "Error occurred when creating security account")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
