package impl

import (
	"ocopea/mysqldsb/models"
	"ocopea/mysqldsb/restapi/operations/dsb_web"
	"github.com/go-openapi/runtime/middleware"
	"log"
	k8sClient "ocopea/kubernetes/client"
	"fmt"
	"errors"
	"net/http"
	"io/ioutil"
	"ocopea/mongodsb/mongo"
	"os"
)

func CopyServiceInstance(
k8s *k8sClient.Client,
params dsb_web.CopyServiceInstanceParams) middleware.Responder {

	err, response := createMysqlCopy(k8s, params.InstanceID, params.CopyDetails)
	if (err != nil) {
		log.Printf("Failed creating copy for instance id %s - %s\n", params.InstanceID, err.Error())
		return getError(dsb_web.NewCopyServiceInstanceDefault(http.StatusInternalServerError), err, http.StatusInternalServerError)
	}

	log.Printf("copy created successfully for %s with %d", response.CopyID, response.Status)
	return dsb_web.NewCopyServiceInstanceOK().WithPayload(&models.CopyServiceInstanceResponse{
		CopyID: *params.CopyDetails.CopyID,
		Status: 0,
		StatusMessage:"yey",
	})
}

func createMysqlCopy(
k8sClient *k8sClient.Client,
instanceId string,
copyDetails *models.CopyServiceInstance) (error, *models.CopyServiceInstanceResponse) {

	log.Printf("Creating copy for %s\n", instanceId);

	mysqlBindings, err := getBindingInfoForInstance(k8sClient, instanceId)
	if (err != nil) {
		return fmt.Errorf("Failed getting bind details for %s: %s", instanceId, err.Error()), nil
	}



	cmd := fmt.Sprintf("mysqldump --all-databases %s", generateMysqlCommandLineArgs())
	dumpReader, err := mongo.MongoDump(*mysqlBindings.BindingPorts[0].Destination + ":" + fmt.Sprint(*mysqlBindings.BindingPorts[0].Port))
	if (err != nil) {
		return fmt.Errorf("Failed getting mongo dump reader for %s: %s", instanceId, err.Error()), nil
	}

	//req, err := http.NewRequest("POST", gCopyRepoLocation + "/internal/data/" + copyRequest.CopyId, bytes.NewReader([]byte("shit!!!")));


	copyRepoUrl, err := getCopyRepUrl(copyDetails)
	if (err != nil) {
		return err, nil
	}

	postCopyUrl := fmt.Sprintf("%s/copies/%s/data", copyRepoUrl, *copyDetails.CopyID)
	log.Printf("Posting copy to %s\n", postCopyUrl)
	req, err := http.NewRequest("POST", postCopyUrl, dumpReader);
	if err != nil {
		return err, nil
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("dsb", DSB_NAME)
	req.Header.Set("copyTimestamp", fmt.Sprintf("%d", copyDetails.CopyTime))
	req.Header.Set("facility", "mongodump")

	httpClient := http.Client{}
	resp, err := httpClient.Do(req)
	log.Printf("After Do for copy %s to CRB URL %s, status is %s(%d)\n", *copyDetails.CopyID, postCopyUrl, resp.Status, resp.StatusCode);
	if (err != nil) {
		return fmt.Errorf(
		"Failed execute request for posting copy %s to CRB with url %s: %s", *copyDetails.CopyID, copyRepoUrl, err.Error()), nil
	}

	// Validating http return code
	defer resp.Body.Close()
	if (resp.StatusCode != http.StatusOK) {
		return fmt.Errorf(
		"Failed execute request for posting copy %s to CRB with url %s: - received status %s", *copyDetails.CopyID, copyRepoUrl, resp.Status), nil
	}

	all, err := ioutil.ReadAll(resp.Body)
	if (err != nil) {
		return fmt.Errorf("Failed reading post copy response from crb url %s for copyId %s - %s", copyRepoUrl, copyDetails.CopyID, err.Error()), nil
	}
	log.Printf("response %s", string(all))

	return nil, &models.CopyServiceInstanceResponse{
		CopyID: *copyDetails.CopyID,
		Status: 0,
		StatusMessage: "yey",
	}

}
func getCopyRepUrl(copyDetails *models.CopyServiceInstance) (string, error) {
	shpanRestCopyUrl := copyDetails.CopyRepoCredentials["url"]
	if (shpanRestCopyUrl == "") {
		return "", errors.New("Invalid copy repo credentails format. missing \"url\"")
	} else {
		return shpanRestCopyUrl, nil
	}

}

func generateMysqlCommandLineArgs(service string, username string, password string) []string {

	// safer to use var than prompt
	//os.Setenv("MYSQL_PWD", password)
	//return fmt.Sprintf("-h %s -P %s -u %s", service, "3306", username)

	return []string{
		fmt.Sprintf("-h %s", service),
		fmt.Sprintf("-u %s", username),
		fmt.Sprintf("-p %s", password),
		"-P 3306",
	}
}



