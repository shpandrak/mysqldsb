package impl

import (
	"ocopea/mysqldsb/models"
	"ocopea/mysqldsb/restapi/operations/dsb_web"
	"github.com/go-openapi/runtime/middleware"
	k8sClient "ocopea/kubernetes/client"
)

func DsbInstancesResponse(k8s *k8sClient.Client) middleware.Responder {

	services, err := k8s.ListServiceInfo(map[string]string {
		"nazIdentifier":"mongo-dsb-instance",
	})

	if err != nil {
		return getError(dsb_web.NewGetServiceInstancesDefault(500), err, 500)
	}

	retVal := make([]*models.ServiceInstance, len(services))
	for i, s := range services {
		retVal[i] = &models.ServiceInstance{
			InstanceID:  s.Labels["nazServiceInstanceId"],
		}
	}
	return dsb_web.NewGetServiceInstancesOK().WithPayload(retVal)
}

