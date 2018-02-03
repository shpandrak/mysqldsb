package impl

import (
	"ocopea/mysqldsb/models"
	"ocopea/mysqldsb/restapi/operations/dsb_web"
	"github.com/go-openapi/runtime/middleware"
	"log"
	k8sClient "ocopea/kubernetes/client"
)

func DeleteMongoInstance(
k8s *k8sClient.Client,
params dsb_web.DeleteServiceInstanceParams) middleware.Responder {

	// First verifying if service exists
	exist, _ := k8s.CheckServiceExists(getServiceNameFromInstanceId(params.InstanceID))
	if (!exist) {
		return dsb_web.NewDeleteServiceInstanceDefault(404)
	}

	err := deleteMongoService(k8s, params.InstanceID)
	if (err != nil) {
		log.Printf("delete instance resulted in error %s\n", err.Error())
		return getError(dsb_web.NewDeleteServiceInstanceDefault(500), err, 500)
	}

	return dsb_web.NewDeleteServiceInstanceOK().WithPayload(
		&models.ServiceInstance{
			InstanceID: params.InstanceID,
		})
}

func deleteMongoService(k8s *k8sClient.Client, instanceId string) error {

	// We need to delete both replication controller and service
	k8sServiceName := getServiceNameFromInstanceId(instanceId)
	err := k8s.DeleteReplicationController(k8sServiceName)
	if (err != nil) {
		return err
	}
	err = k8s.DeleteService(k8sServiceName)
	if (err != nil) {
		return err
	}

	return nil

}



