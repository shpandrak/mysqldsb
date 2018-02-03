package impl

import (
	"ocopea/mysqldsb/models"
	"ocopea/mysqldsb/restapi/operations/dsb_web"
	"github.com/go-openapi/runtime/middleware"
	k8sClient "ocopea/kubernetes/client"
	"fmt"
	"ocopea/kubernetes/client/v1"
	"os"
	"bytes"
	"os/exec"
	"log"
	"net/http"
)

type MongoBindingInfo struct {
	Server string `json:"server"`
	Port   string `json:"port"`
}

func DsbGetServiceInstancesResponse(
k8s *k8sClient.Client,
params dsb_web.GetServiceInstanceParams) middleware.Responder {

	exist, _ := k8s.CheckServiceExists(getServiceNameFromInstanceId(params.InstanceID))

	if (!exist) {
		return dsb_web.NewGetServiceInstanceDefault(http.StatusNotFound)
	}

	bindings, err := getBindingInfoForInstance(k8s, params.InstanceID)
	if (err != nil) {
		return getError(dsb_web.NewGetServiceInstanceDefault(500), err, 500)
	}

	return dsb_web.NewGetServiceInstanceOK().WithPayload(bindings)


}

func getBindingInfoForInstance(k8s *k8sClient.Client, instanceId string) (*models.ServiceInstanceDetails, error) {
	serviceName := getServiceNameFromInstanceId(instanceId)

	isReady, svc, err := k8s.TestService(serviceName)
	if (err != nil) {
		return nil, err
	}

	// Still creating
	if (!isReady) {
		return &models.ServiceInstanceDetails{
			InstanceID: instanceId,
			State: "CREATING",
		}, nil

	}

	var port int32 = 3306
	var server string
	bindingInfo := make(map[string]string)
	bindingInfo["port"] = fmt.Sprintf("%d", port)
	if (svc.Spec.Type == v1.ServiceTypeClusterIP) {
		server = svc.Spec.ClusterIP
		bindingInfo["server"] = server
	} else {
		return nil, fmt.Errorf("Unsupported k8s service type %s for service %s", svc.Spec.Type, serviceName);
	}

	// Verify mysql is alive
	mysql := fmt.Sprintf("%s:%d", server , port)
	fmt.Printf("TESTING for mongo to start accepting connections %s", mysql)
	err = verifyMysqlLive(mysql)
	if (err != nil) {
		log.Printf("mongo not live yet %s on %s - %s", serviceName, mysql, err.Error())
		// Assuming still creating
		return &models.ServiceInstanceDetails{
			InstanceID: instanceId,
			State: "CREATING",
		}, nil
	}

	p := "tcp"
	return &models.ServiceInstanceDetails{
		InstanceID: instanceId,
		State: "RUNNING",
		Binding: bindingInfo,
		BindingPorts:[]*models.BindingPort{
			{
				Protocol:&p,
				Destination:&server,
				Port: &port},
		},
		Size:500,
		StorageType: "Kubernetes Temp Volume",
	}, nil

}

func verifyMysqlLive(mongoDBAddress string) (error) {
	log.Println("Verifying mysql live")

	cmdName := "mysql"
	cmdArgs := generateMysqlCommandLineArgs()
		[]string{"status"}

	cmd := exec.Command(cmdName, cmdArgs...)
	fmt.Println(cmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	return cmd.Run()
}


