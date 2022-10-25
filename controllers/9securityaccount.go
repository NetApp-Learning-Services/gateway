// From:  https://github.com/nheidloff/operator-sample-go/blob/main/operator-application/controllers/application/deployment.go

package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	gatewayv1alpha1 "gateway/api/v1alpha1"
	"gateway/ontap"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	secondAuthMethod = "none" // Special word
)

func (r *StorageVirtualMachineReconciler) reconcileSecurityAccount(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine, oc *ontap.Client, credentials *corev1.Secret, log logr.Logger) (ctrl.Result, error) {

	log.Info("Step 9: Verify SVM management account is update to date")

	userNameToModify := string(credentials.Data["username"])

	// Check to see if we have a uuid
	if svmCR.Spec.SvmUuid == "" {
		return ctrl.Result{}, errors.NewBadRequest("No SVM uuid during security account update")
	}

	// Check to see if username exists
	user, err := oc.GetSecurityAccount(svmCR.Spec.SvmUuid, userNameToModify)
	if err != nil {
		log.Error(err, "Error checking to see if username exists")
	}

	if user.Name != "" {
		// User already created - need to patch
		log.Info("credentials " + userNameToModify + " - need to patch")
		var payload ontap.SecurityAccountPayload

		ssh := ontap.Application{
			AppType:          ontap.Ssh,
			SecondAuthMethod: secondAuthMethod,
		}
		ssh.AuthMethods = append(ssh.AuthMethods, ontap.Password)
		payload.Applications = append(payload.Applications, ssh)
		ontapi := ontap.Application{
			AppType:          ontap.Ontapi,
			SecondAuthMethod: secondAuthMethod,
		}
		ontapi.AuthMethods = append(ontapi.AuthMethods, ontap.Password)
		payload.Applications = append(payload.Applications, ontapi)

		http := ontap.Application{
			AppType:          ontap.Http,
			SecondAuthMethod: secondAuthMethod,
		}
		http.AuthMethods = append(http.AuthMethods, ontap.Password)
		payload.Applications = append(payload.Applications, http)

		//payload.Role = ontap.Vsadmin
		payload.Password = string(credentials.Data["password"])
		var a bool = false
		payload.Locked = &a // always unlock

		log.Info("Security account payload: " + fmt.Sprintf("%#v\n", payload))

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			//error creating the json body
			log.Error(err, "Error creating the json payload for security account patch")
			return ctrl.Result{}, err
		}

		log.Info("Security account patch attempt")
		err = oc.PatchSecurityAccount(jsonPayload, svmCR.Spec.SvmUuid, userNameToModify)
		if err != nil {
			log.Error(err, "Error occurred when patching security account")
			_ = r.setConditionVsadminSecretUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
			return ctrl.Result{}, nil //TODO: CHANGE THIS
		} else {
			_ = r.setConditionVsadminSecretUpdate(ctx, svmCR, CONDITION_STATUS_TRUE)
		}
	} else {
		log.Info("User not found - try to create")
		var payload ontap.SecurityAccountPayload
		payload.Name = userNameToModify
		payload.Owner.Uuid = svmCR.Spec.SvmUuid

		ssh := ontap.Application{
			AppType:          ontap.Ssh,
			SecondAuthMethod: secondAuthMethod,
		}
		ssh.AuthMethods = append(ssh.AuthMethods, ontap.Password)
		payload.Applications = append(payload.Applications, ssh)
		ontapi := ontap.Application{
			AppType:          ontap.Ontapi,
			SecondAuthMethod: secondAuthMethod,
		}
		ontapi.AuthMethods = append(ontapi.AuthMethods, ontap.Password)
		payload.Applications = append(payload.Applications, ontapi)

		http := ontap.Application{
			AppType:          ontap.Http,
			SecondAuthMethod: secondAuthMethod,
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
			_ = r.setConditionVsadminSecretUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
			return ctrl.Result{}, err
		} else {
			_ = r.setConditionVsadminSecretUpdate(ctx, svmCR, CONDITION_STATUS_TRUE)
		}
	}

	return ctrl.Result{}, nil
}
