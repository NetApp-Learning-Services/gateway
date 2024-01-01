// From:  https://github.com/nheidloff/operator-sample-go/blob/main/operator-application/controllers/application/deployment.go

package controller

import (
	"context"
	"encoding/json"
	"fmt"

	gatewayv1alpha2 "gateway/api/v1alpha2"
	"gateway/internal/controller/ontap"
	defaultLog "log"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

const secondAuthMethod = "none" // magic word

func (r *StorageVirtualMachineReconciler) reconcileSecurityAccount(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, oc *ontap.Client, credentials *corev1.Secret, log logr.Logger) error {

	log.Info("STEP 9: Verify SVM management account is update to date")

	userNameToModify := string(credentials.Data["username"])

	// Check to see if we have a uuid
	if svmCR.Spec.SvmUuid == "" {
		err := errors.NewBadRequest("No SVM uuid during security account update")
		log.Error(err, "Error while updating SVM management credentials - requeuing")
		return err
	}

	// Check to see if username exists
	user, err := oc.GetSecurityAccount(svmCR.Spec.SvmUuid, userNameToModify)
	if err != nil {
		log.Error(err, "Error checking to see if username exists - requeuing")
	}

	//log.Info("User: " + user.Name + " locked: " + fmt.Sprintf("%v", user.Locked))

	if user.Name != "" && user.Locked {
		// User already created - need to patch
		log.Info("Credentials " + userNameToModify + " - need to patch")
		var payload ontap.SecurityAccountPatchPayload

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

		if oc.Debug {
			defaultLog.Printf("[DEBUG] Security account payload: " + fmt.Sprintf("%#v\n", payload))
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			//error creating the json body
			log.Error(err, "Error creating the json payload for security account patch - requeuing")
			return err
		}

		log.Info("Security account patch attempt")
		err = oc.PatchSecurityAccount(jsonPayload, svmCR.Spec.SvmUuid, userNameToModify)
		if err != nil {
			log.Error(err, "Error occurred when patching security account - requeuing")
			_ = r.setConditionVsadminSecretUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
			r.Recorder.Event(svmCR, "Warning", "VsadminUpdateFailed", "Error: "+err.Error())
			return err
		} else {
			log.Info("SVM managment credentials updated in ONTAP")
			_ = r.setConditionVsadminSecretUpdate(ctx, svmCR, CONDITION_STATUS_TRUE)
			r.Recorder.Event(svmCR, "Normal", "VsadminUpdateSuccessed", "Updated SVM admin")
		}

	} else {
		log.Info("Nothing to do - skipping STEP 9")
		return nil //do nothing
	}

	if user.Name == "" {
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

		if oc.Debug {
			defaultLog.Printf("[DEBUG] Security account payload: " + fmt.Sprintf("%#v\n", payload))
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			//error creating the json body
			log.Error(err, "Error creating the json payload for security account creation - requeuing")
			return err
		}

		log.Info("Security account creation attempt")
		err = oc.CreateSecurityAccount(jsonPayload)
		if err != nil {
			log.Error(err, "Error occurred when creating security account - requeuing")
			_ = r.setConditionVsadminSecretUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
			r.Recorder.Event(svmCR, "Warningg", "VsadminCreationFailed", "Error: "+err.Error())
			return err
		} else {
			log.Info("SVM managment credentials created in ONTAP")
			_ = r.setConditionVsadminSecretUpdate(ctx, svmCR, CONDITION_STATUS_TRUE)
			r.Recorder.Event(svmCR, "Normal", "VsadminCreationSuccessed", "Created SVM admin")
		}
	}

	return nil
}
