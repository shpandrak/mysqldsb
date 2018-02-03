package impl

import (
	"ocopea/mysqldsb/models"
	"ocopea/mysqldsb/restapi/operations/dsb_web"
	"github.com/go-openapi/runtime/middleware"
	"log"
	"ocopea/kubernetes/client/v1"
	"ocopea/kubernetes/client/types"
	k8sClient "ocopea/kubernetes/client"
	"strings"
	"fmt"
	"errors"
	"ocopea/mysqldsb/mongo"
	"net/http"
	"io"
	"bufio"
	"time"
)

func CreateInstanceResponse(
k8s *k8sClient.Client,
params dsb_web.CreateServiceInstanceParams) middleware.Responder {

	err := createMongoService(k8s, params.ServiceSettings)
	if (err != nil) {
		log.Printf("create instance resulted in error %s\n", err.Error())
		return getError(dsb_web.NewCreateServiceInstanceDefault(500), err, 500)
	}

	return dsb_web.NewCreateServiceInstanceOK().WithPayload(
		&models.ServiceInstance{
			InstanceID: *params.ServiceSettings.InstanceID,
		})
}

func getError(failedResponse ErrorResponse, err error, errCode int) middleware.Responder {
	modelErrorInt := int32(errCode)
	errStr := err.Error()
	modelsError := models.Error{Code: &modelErrorInt, Message: &errStr}
	failedResponse.SetPayload(&modelsError)
	return failedResponse

}

//todo:better
func getServiceNameFromInstanceId(instanceId string) string {
	l := len(instanceId)
	return "m-" + instanceId[l - 22:l]
}

func createBlankMysqlService(k8s *k8sClient.Client, instanceId string) error {
	log.Printf("Creating mongo service with id %s", instanceId);
	appUniqueName := getServiceNameFromInstanceId(instanceId)

	var replicas int = 1;
	// Building rc spec

	spec := v1.ReplicationControllerSpec{}
	spec.Replicas = &replicas
	spec.Selector = make(map[string]string)
	spec.Selector["app"] = appUniqueName
	spec.Template = &v1.PodTemplateSpec{}
	spec.Template.ObjectMeta = v1.ObjectMeta{}
	spec.Template.ObjectMeta.Labels = make(map[string]string)
	spec.Template.ObjectMeta.Labels["app"] = appUniqueName
	spec.Template.ObjectMeta.Labels["nazKind"] = "dsbInstance"

	containerSpec := v1.Container{}
	containerSpec.Name = appUniqueName

	containerSpec.Image = "mysql:5.6";
	//containerSpec.Env = envVar;
	//todo: 1) all ports
	//todo: 2) allow no http port at all
	//if (appManifest.HttpPort != nil){
	containerSpec.Ports = []v1.ContainerPort{{ContainerPort:3306}}
	//}

	containers := []v1.Container{containerSpec}

	spec.Template.Spec = v1.PodSpec{}
	spec.Template.Spec.Containers = containers



	// Create a replicationController object for running the app
	rc := &v1.ReplicationController{}
	rc.Name = appUniqueName;
	rc.Labels = make(map[string]string)
	rc.Labels["nazKind"] = "dsbInstance"
	rc.Spec = spec

	_, err := k8s.CreateReplicationController(rc, false)
	if err != nil {
		return err
	}

	svc := &v1.Service{}
	svc.Name = appUniqueName
	svc.Spec.Type = v1.ServiceTypeClusterIP;
	svc.Labels = map[string]string{
		"nazIdentifier":"mysql-dsb-instance",
		"nazServiceInstanceId":instanceId,
	}
	svc.Spec.Ports = []v1.ServicePort{{
		Port:27017,
		TargetPort: types.NewIntOrStringFromInt(3306),
		Protocol:v1.ProtocolTCP,
		Name:"tcp",
	}}
	svc.Spec.Selector = map[string]string{"app": appUniqueName}
	svc, err = k8s.CreateService(svc, false)

	return err

}
func createMongoService(k8s *k8sClient.Client, createServiceSettings *models.CreateServiceInstance) error {

	// First creating a blank mongo service
	err := createBlankMysqlService(k8s, *createServiceSettings.InstanceID)
	if (err != nil) {
		return err
	}

	// Checking if need to restore from copy
	if (createServiceSettings.RestoreInfo != nil) {
		err = restoreMysqlToExistingInstance(
			k8s,
			*createServiceSettings.InstanceID,
			createServiceSettings.RestoreInfo)

		if (err != nil) {
			return err
		}
	}
	return nil

}

func restoreMysqlToExistingInstance(k8s *k8sClient.Client, instanceId string, restoreInfo *models.DsbRestoreCopyInfo) error {

	// Testing for supported copy protocol
	switch strings.ToLower(restoreInfo.CopyRepoProtocol) {
	case "shpanrest":
		return restoreMysqlUsingShpanRestCopyProtocol(k8s, instanceId, restoreInfo)
	default:
		return fmt.Errorf("Unsupported resotre copy protocol %s", restoreInfo.CopyRepoProtocol)
	}
}

func restoreMysqlUsingShpanRestCopyProtocol(k8s *k8sClient.Client, instanceId string, restoreInfo *models.DsbRestoreCopyInfo) error {

	// Extracting url from credentials
	copyRepoUrl := restoreInfo.CopyRepoCredentials["url"]
	if (copyRepoUrl == "") {
		return errors.New("Missing url in ShpanRest copy repo credentials. unable to restore")
	}

	mongoBindings, err := getBindingInfoForInstance(k8s, instanceId)
	if (err != nil) {
		return fmt.Errorf("Failed getting bind details for %s: %s", instanceId, err.Error())
	}


	// In case service is still creating wait for it for a while
	maxRetries := 50
	for mongoBindings.State == "CREATING" && maxRetries > 0 {
		time.Sleep(5 * time.Second)
		mongoBindings, err = getBindingInfoForInstance(k8s, instanceId)
		if (err != nil) {
			return fmt.Errorf("Failed getting bind details for %s: %s", instanceId, err.Error())
		}
		maxRetries--
	}

	if mongoBindings.State != "RUNNING" {
		return fmt.Errorf("Could not restore for instance %s, since it is in state %s", instanceId, mongoBindings.State)
	}

	// Reading copy data from dsb
	copyUrl := fmt.Sprintf("%s/copies/%s/data", copyRepoUrl, restoreInfo.CopyID)
	log.Printf("Reading copy for instance %s via ShpanRest protocol from url %s", instanceId, copyUrl)
	req, err := http.NewRequest("GET", copyUrl, nil);
	if err != nil {
		return err
	}

	//req.Header.Set("Content-Type", "application/octet-stream")
	httpClient := http.Client{}
	resp, err := httpClient.Do(req)
	if (err != nil) {
		return fmt.Errorf("Failed reading copy body for instance id %s using copy url %s", instanceId, copyUrl)
	}
	defer resp.Body.Close();

	// Validating http return code
	defer resp.Body.Close()
	if (resp.StatusCode != http.StatusOK) {
		return fmt.Errorf(
			"Failed execute request for getting copy from url %s - received status %s", copyRepoUrl, resp.Status)
	}
	log.Printf("After getting url %s to, status is %s\n", copyUrl, resp.Status);

	restoreWriter, restoreCmd, err := mongo.MongoRestore(fmt.Sprintf("%s:%d", *mongoBindings.BindingPorts[0].Destination, *mongoBindings.BindingPorts[0].Port))
	if (err != nil) {
		return fmt.Errorf("Failed getting mongo restore writer for %s - %s", instanceId, err.Error())
	}
	reader := bufio.NewReader(resp.Body)
	written, err := io.Copy(restoreWriter, reader)
	if (err != nil) {
		return fmt.Errorf("Failed copying data from request body to mongorestore for %s via url %s - %s", instanceId, copyUrl, err.Error())
	}
	fmt.Printf("restored copy with %d bytes", written)

	err = restoreWriter.Close()
	if (err != nil) {
		return fmt.Errorf("Failed closing input stream of mongorestore for %s via url %s - %s", instanceId, copyUrl, err.Error())
	}

	err = restoreCmd.Wait()
	if (err != nil) {
		return fmt.Errorf("Failed executng restore command for %s via url %s - %s", instanceId, copyUrl, err.Error())
	}
	fmt.Printf("Restoring %s complete\n", instanceId)
	//	io.WriteString(w, "{\"status\":0,\"statusMessage\":\"Mongo db restored, you foolish english kniggits!\"}")
	return nil

}



