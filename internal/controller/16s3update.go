package controller

import (
	"context"
	"encoding/json"
	"fmt"
	gateway "gateway/api/v1beta2"
	"gateway/internal/controller/ontap"
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const S3LifType = "" //magic word

func (r *StorageVirtualMachineReconciler) reconcileS3Update(ctx context.Context, svmCR *gateway.StorageVirtualMachine,
	uuid string, oc *ontap.Client, log logr.Logger) error {
	log.Info("STEP 16: Update S3 service")

	// S3 SERVICE

	createS3Service := false
	updateS3Service := false

	// Check to see if S3 configuration is provided in custom resource
	if svmCR.Spec.S3Config == nil {
		// If not, exit with no error
		log.Info("No S3 service defined - skipping STEP 16")
		return nil
	}

	S3Service, err := oc.GetS3ServiceBySvmUuid(uuid)
	if err != nil && errors.IsNotFound(err) {
		createS3Service = true
	} else if err != nil {
		//some other error
		log.Error(err, "Error retrieving S3 service for SVM by UUID - requeuing")
		return err
	}
	var upsertS3Service ontap.S3Service

	if createS3Service {
		log.Info("No S3 service defined for SVM: " + uuid + " - creating S3 service")

		upsertS3Service.Svm.Uuid = svmCR.Spec.SvmUuid
		upsertS3Service.Enabled = &svmCR.Spec.S3Config.Enabled
		upsertS3Service.IsHttpEnabled = &svmCR.Spec.S3Config.Http.Enabled
		upsertS3Service.IsHttpsEnabled = &svmCR.Spec.S3Config.Https.Enabled
		upsertS3Service.Name = &svmCR.Spec.SvmName

		jsonPayload, err := json.Marshal(upsertS3Service)
		if err != nil {
			//error creating the json body
			log.Error(err, "Error creating the json payload for S3 service creation - requeuing")
			_ = r.setConditionS3Service(ctx, svmCR, CONDITION_STATUS_FALSE)
			return err

		}

		if oc.Debug {
			log.Info("[DEBUG] S3 service creation payload: " + fmt.Sprintf("%#v\n", upsertS3Service))
		}

		err = oc.CreateS3Service(jsonPayload)
		if err != nil {
			log.Error(err, "Error creating the S3 service - requeuing")
			_ = r.setConditionS3Service(ctx, svmCR, CONDITION_STATUS_FALSE)
			r.Recorder.Event(svmCR, "Warning", "S3CreationFailed", "Error: "+err.Error())
			return err
		}
		_ = r.setConditionS3Service(ctx, svmCR, CONDITION_STATUS_TRUE)
		r.Recorder.Event(svmCR, "Normal", "S3CreationSucceeded", "Created S3 service successfully")
		log.Info("S3 service created successful")
	} else {
		// Compare enabled to custom resource enabled
		if *S3Service.Enabled != svmCR.Spec.S3Config.Enabled {
			updateS3Service = true
			upsertS3Service.Enabled = &svmCR.Spec.S3Config.Enabled
		}

		if *S3Service.IsHttpEnabled != svmCR.Spec.S3Config.Http.Enabled {
			updateS3Service = true
			upsertS3Service.IsHttpEnabled = &svmCR.Spec.S3Config.Http.Enabled
		}

		if *S3Service.IsHttpsEnabled != svmCR.Spec.S3Config.Https.Enabled {
			updateS3Service = true
			upsertS3Service.IsHttpsEnabled = &svmCR.Spec.S3Config.Https.Enabled
		}

		if oc.Debug && updateS3Service {
			log.Info("[DEBUG] S3 service update payload: " + fmt.Sprintf("%#v\n", upsertS3Service))
		}

		if updateS3Service {
			jsonPayload, err := json.Marshal(upsertS3Service)
			if err != nil {
				//error creating the json body
				log.Error(err, "Error creating the json payload for S3 service update - requeuing")
				_ = r.setConditionS3Service(ctx, svmCR, CONDITION_STATUS_FALSE)
				return err
			}

			//Patch S3 service
			log.Info("S3 service update attempt for SVM: " + uuid)
			err = oc.PatchS3Service(uuid, jsonPayload)
			if err != nil {
				log.Error(err, "Error updating the S3 service - requeuing")
				_ = r.setConditionS3Service(ctx, svmCR, CONDITION_STATUS_FALSE)
				r.Recorder.Event(svmCR, "Warning", "S3UpdateFailed", "Error: "+err.Error())
				return err
			}
			log.Info("S3 service updated successful")
			_ = r.setConditionS3Service(ctx, svmCR, CONDITION_STATUS_TRUE)
			r.Recorder.Event(svmCR, "Normal", "S3UpdateSucceeded", "Updated S3 service successfully")
		} else {
			log.Info("No S3 service changes detected - skip updating")
		}
	}

	// END S3 SERVICE

	// S3 LIFS

	// Check to see if S3 interfaces are defined in custom resource
	if svmCR.Spec.S3Config.Lifs == nil {
		// If not, exit with no error
		log.Info("No S3 LIFs defined - skipping updates")
	} else {

		createS3Lifs := false

		// Check to see if S3 interfaces defined and compare to custom resource's definitions
		lifs, err := oc.GetS3InterfacesBySvmUuid(uuid, S3LifType)
		if err != nil {
			//error creating the json body
			log.Error(err, "Error getting S3 service LIFs for SVM: "+uuid)
			_ = r.setConditionS3Lif(ctx, svmCR, CONDITION_STATUS_FALSE)
			return err
		}

		if lifs.NumRecords == 0 {
			// no data LIFs for the SVM provided in UUID
			// create new LIF(s)
			log.Info("No LIFs defined for SVM: " + uuid + " - creating S3 Lif(s)")
			createS3Lifs = true
		}

		for index, val := range svmCR.Spec.S3Config.Lifs {

			// Check to see need to create all LIFS or
			// if lifs.Records[index] is out of index - if so, need to create LIF
			if createS3Lifs || index > lifs.NumRecords-1 {
				// Need to create LIF for val
				err = CreateLif(val, S3LifType, uuid, oc, log)
				if err != nil {
					_ = r.setConditionS3Lif(ctx, svmCR, CONDITION_STATUS_FALSE)
					r.Recorder.Event(svmCR, "Warning", "S3CreationLifFailed", "Error: "+err.Error())
					return err
				}

			} else {
				//check to see if we need to update the LIF

				//do I need this? checking to see if I have valid LIF returned
				if reflect.ValueOf(lifs.Records[index]).IsZero() {
					break
				}

				err = UpdateLif(val, lifs.Records[index], S3LifType, oc, log)
				if err != nil {
					_ = r.setConditionS3Lif(ctx, svmCR, CONDITION_STATUS_FALSE)
					r.Recorder.Event(svmCR, "Warning", "S3UpdateLifFailed", "Error: "+err.Error())
					// e := err.(*apiError)
					// if e.errorCode == 1 {
					// // Json parsing error
					// 	return err
					// } else if e.errorCode == 2 {
					// // Patch error
					// 	return err
					// } else {
					// 	return err
					// }
					return err
				}
			}

			// Delete all SVM data LIFs that are not defined in the custom resource
			for i := len(svmCR.Spec.S3Config.Lifs); i < lifs.NumRecords; i++ {
				log.Info("S3 LIF delete attempt: " + lifs.Records[i].Name)
				oc.DeleteIpInterface(lifs.Records[i].Uuid)
				if err != nil {
					log.Error(err, "Error occurred when deleting S3 LIF: "+lifs.Records[i].Name)
					// don't requeue on failed delete request
					// no condition error
					// return err
				} else {
					log.Info("S3 LIF delete successful: " + lifs.Records[i].Name)
				}
			}

			_ = r.setConditionS3Lif(ctx, svmCR, CONDITION_STATUS_TRUE)
			r.Recorder.Event(svmCR, "Normal", "S3UpsertLifSucceeded", "Upserted S3 LIF(s) successfully")
		} // End looping through S3 LIF definitions in custom resource
	} // LIFS defined in custom resources

	// END S3 LIFS

	// S3 Users
	if svmCR.Spec.S3Config.Users == nil {
		// If none, exit with no error
		log.Info("No S3 users defined - skipping")
	} else {
		createS3Users := false

		// Check to see if S3 users defined and compare to custom resource
		usersRetrieved, err := oc.GetS3UsersBySvmUuid(uuid)
		if err != nil {
			//error creating the json body
			log.Error(err, "Error getting S3 users for SVM: "+uuid+" - requeuing")
			_ = r.setConditionS3User(ctx, svmCR, CONDITION_STATUS_FALSE)
			return err
		}

		if usersRetrieved.NumRecords == 0 {
			// No exports for the SVM provided in UUID
			// create new export(s)
			log.Info("No S3 users defined for SVM: " + uuid + " - creating S3 user(s)")
			createS3Users = true
		}

		if createS3Users {
			// creating users
			for _, val := range svmCR.Spec.S3Config.Users {
				user, err := CreateUser(val, uuid, oc, log)
				if err != nil {
					_ = r.setConditionS3User(ctx, svmCR, CONDITION_STATUS_FALSE)
					return err
				} else {
					// Create a secret with the access key and secret key
					secret := &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      user.Records[0].Name,
							Namespace: svmCR.Namespace,
						},
						StringData: map[string]string{
							"accessKeyID":     user.Records[0].AccessKey,
							"secretAccessKey": user.Records[0].SecretKey,
						},
						Type: corev1.SecretTypeOpaque,
					}
					err = r.Create(ctx, secret)
					if err != nil {
						_ = r.setConditionS3UserSecret(ctx, svmCR, CONDITION_STATUS_FALSE)
						log.Error(err, "Error creating S3 user secret for SVM: "+uuid+" and user: "+user.Records[0].Name+" with access key: "+user.Records[0].AccessKey+" and secret key: "+user.Records[0].SecretKey)
					}
					log.Info("S3 User and secret creation succesful: " + val.Name)
				}
			}
		} else {
			// delete all users and recreate to match custom resource
		}
	}

	return nil
}

// STEP 16
// S3 update
// Note: Status of S3_SERVICE can only be true or false
const CONDITION_TYPE_S3_SERVICE = "16S3service"
const CONDITION_REASON_S3_SERVICE = "S3service"
const CONDITION_MESSAGE_S3_SERVICE_TRUE = "S3 service configuration succeeded"
const CONDITION_MESSAGE_S3_SERVICE_FALSE = "S3 service configuration failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionS3Service(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_S3_SERVICE) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_S3_SERVICE, CONDITION_REASON_S3_SERVICE)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_S3_SERVICE, status,
			CONDITION_REASON_S3_SERVICE, CONDITION_MESSAGE_S3_SERVICE_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_S3_SERVICE, status,
			CONDITION_REASON_S3_SERVICE, CONDITION_MESSAGE_S3_SERVICE_FALSE)
	}
	return nil
}

const CONDITION_REASON_S3_LIF = "S3lif"
const CONDITION_MESSAGE_S3_LIF_TRUE = "S3 LIF configuration succeeded"
const CONDITION_MESSAGE_S3_LIF_FALSE = "S3 LIF configuration failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionS3Lif(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_S3_LIF) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_S3_SERVICE, CONDITION_REASON_S3_LIF)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_S3_SERVICE, status,
			CONDITION_REASON_S3_LIF, CONDITION_MESSAGE_S3_LIF_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_S3_SERVICE, status,
			CONDITION_REASON_S3_LIF, CONDITION_MESSAGE_S3_LIF_FALSE)
	}
	return nil
}

const CONDITION_REASON_S3_USER = "S3user"
const CONDITION_MESSAGE_S3_USER_TRUE = "S3 user configuration succeeded"
const CONDITION_MESSAGE_S3_USER_FALSE = "S3 user configuration failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionS3User(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_S3_LIF) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_S3_SERVICE, CONDITION_REASON_S3_LIF)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_S3_SERVICE, status,
			CONDITION_REASON_S3_USER, CONDITION_MESSAGE_S3_USER_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_S3_SERVICE, status,
			CONDITION_REASON_S3_USER, CONDITION_MESSAGE_S3_USER_FALSE)
	}
	return nil
}

const CONDITION_REASON_S3_USERSECRET = "S3usersecret"
const CONDITION_MESSAGE_S3_USERSECRET_TRUE = "S3 user secret configuration succeeded"
const CONDITION_MESSAGE_S3_USERSECRET_FALSE = "S3 user secret configuration failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionS3UserSecret(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_S3_LIF) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_S3_SERVICE, CONDITION_REASON_S3_LIF)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_S3_SERVICE, status,
			CONDITION_REASON_S3_USER, CONDITION_MESSAGE_S3_USERSECRET_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_S3_SERVICE, status,
			CONDITION_REASON_S3_USER, CONDITION_MESSAGE_S3_USERSECRET_FALSE)
	}
	return nil
}
