//
//A RESTful style web-services framework for the Go language. </br>
//Creating services in Go is straight forward, GoRest? takes this a step further by adding a layer that
//makes tedious tasks much more automated and avoids regular pitfalls. <br/>
//This gives you the opportunity to focus more on the task at hand... minor low-level http handling.<br/>
//
//
//Example usage below:
//
package main

import (
	"code.google.com/p/gorest"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	//"encoding/json"
	//"cloudfoundry.com/cf-utils"
	cfutils "cf_utils"
	scalerutils "scaler_utils"
)

/*
type AppDetail struct {
  TargetQ string `json:"queue"`
  AppName string `json:"app"`
  Org     string `json:"org"`
  Space   string `json:"space"`
}

type AppInstanceDetail struct {
  TargetQ   string
  AppName   string
  AppGuid   string
  Instances int
}
*/

var api_endpoint string

var unit_increment int
var listen_port string

var appDetailsMapMutex = &sync.RWMutex{}
var appInstanceMapMutex = &sync.RWMutex{}

var appDetailsMap map[string]scalerutils.AppDetail
var appInstanceMap map[string]scalerutils.AppInstanceDetail

func init() {
	listen_port = os.Getenv("PORT")
	if listen_port == "" {
		listen_port = "8080"
	}

	fmt.Printf("App Auto Scaler running on Listen Port: %s !!\n", listen_port)

	unit_increment = 1
}

func main() {

	//var appDetailsMapMutex = &sync.Mutex{}
	appDetailsMap = make(map[string]scalerutils.AppDetail)
	appInstanceMap = make(map[string]scalerutils.AppInstanceDetail)

	gorest.RegisterMarshaller("application/json", gorest.NewJSONMarshaller())
	gorest.RegisterService(new(AppAutoScalerService)) //Register our service
	http.Handle("/", gorest.Handle())
	http.ListenAndServe(":"+listen_port, nil)

}

//Service Definition
type AppAutoScalerService struct {
	gorest.RestService `root:"/"`

	load       gorest.EndPoint `method:"POST" path:"/load/" postdata:"[]AppDetail"`
	register   gorest.EndPoint `method:"POST" path:"/register/" postdata:"AppDetail" `
	deregister gorest.EndPoint `method:"DELETE" path:"/register/{destnName:string}"`

	details      gorest.EndPoint `method:"GET" path:"/details" output:"string"`
	queueDetails gorest.EndPoint `method:"GET" path:"/details/{destnName:string}" output:"string"`

	scaleUp   gorest.EndPoint `method:"PUT" path:"/scale/{destnName:string}/up/{increments:int}" postdata:"string" `
	scaleDown gorest.EndPoint `method:"PUT" path:"/scale/{destnName:string}/down/{increments:int}" postdata:"string" `
}

func (serv AppAutoScalerService) Register(appDetail scalerutils.AppDetail) {

	fmt.Printf("\nApp Details ... %v\n", appDetail)
	destnName := appDetail.TargetQ

	appDetailsMapMutex.Lock()
	appDetailsMap[destnName] = appDetail
	appDetailsMapMutex.Unlock()

	// Get the actual instance count...
	appGuidInstanceDetail := cfutils.FindApp(appDetail.Org, appDetail.Space, appDetail.AppName)

    var instances int
	appGuid := appGuidInstanceDetail["guid"]
	if appGuid != "" {
		instanceCount, err := strconv.Atoi(appGuidInstanceDetail["instances"])

		if err != nil {
			fmt.Printf("Could not convert into int: %v\n", err)
		} else {
			instances = instanceCount
		}
	} else {
		fmt.Printf("\nWarning!! App could not be located successfully!\n")
	}

	fmt.Printf("\nApp: %s under org: %s, space: %s, App Guid: %s, Instances: %d\n", appDetail.AppName, appDetail.Org, appDetail.Space, appGuid, instances)

	appInstanceDetails := scalerutils.AppInstanceDetail{destnName, appDetail.AppName, appGuid, instances}

	appInstanceMapMutex.Lock()
	appInstanceMap[destnName] = appInstanceDetails
	appInstanceMapMutex.Unlock()
}

func (serv AppAutoScalerService) Load(appDetails []scalerutils.AppDetail) {
	for _, appDetail := range appDetails {
		serv.Register(appDetail)
	}
}

func (serv AppAutoScalerService) Details() string {
	output := "["
	appInstanceMapMutex.RLock()
	for key, value := range appInstanceMap {
		appInstance := value
		output += fmt.Sprintf("{ 'DestinationName' : '%s', 'AppName' : '%s', 'Instances': '%d'}\n", key, appInstance.AppName, appInstance.Instances)
	}
	appInstanceMapMutex.RUnlock()
	output = output + "]"
	return output
}

func findAppDetail(targetQ string) scalerutils.AppDetail {

	appDetail := appDetailsMap[targetQ]
	return appDetail
}

func deleteAppDetail(targetQ string) {

	delete(appDetailsMap, targetQ)
}

func findAppInstanceDetail(targetQ string) scalerutils.AppInstanceDetail {

	appInstance := appInstanceMap[targetQ]
	return appInstance
}

func deleteAppInstanceDetail(targetQ string) {

	delete(appInstanceMap, targetQ)
}

func (serv AppAutoScalerService) QueueDetails(targetQ string) string {
	appInstanceMapMutex.RLock()
	appInstance := findAppInstanceDetail(targetQ)
	appInstanceMapMutex.RUnlock()

	if appInstance.AppName == "" {
		fmt.Printf("Error!! App Instance not found for %s\n", targetQ)
		return ""
	}

	return fmt.Sprintf("{ 'DestinationName' : '%s', 'AppName' : '%s', 'Instances': '%d'}\n", targetQ, appInstance.AppName, appInstance.Instances)
}

func (serv AppAutoScalerService) Deregister(targetQ string) {
	appInstanceMapMutex.Lock()
	deleteAppInstanceDetail(targetQ)
	appInstanceMapMutex.Unlock()

	appDetailsMapMutex.Lock()
	deleteAppDetail(targetQ)
	appDetailsMapMutex.Unlock()
}

func (serv AppAutoScalerService) Scale(targetQ string, increments int, up bool) {
	appInstanceMapMutex.Lock()
	appInstanceDetails := appInstanceMap[targetQ]
	if appInstanceDetails.AppName == "" {
		fmt.Printf("Error!! App Instance not found for %s\n", targetQ)
		return
	}

	if increments == 0 {
		increments = unit_increment
	}

	increment := increments
	if !up {
		increment = -increment
	}

	appInstanceDetails.Instances += increment

	if appInstanceDetails.Instances < 0 {
		appInstanceDetails.Instances = 0
	}

	appInstanceMap[targetQ] = appInstanceDetails
	appInstanceMapMutex.Unlock()

	direction := "Up"
	if !up {
		direction = "Down"
	}

	fmt.Printf("Scaling[%s] App : { 'DestinationName' : '%s', 'AppName' : '%s', 'Instances': '%d'}\n", direction, targetQ, appInstanceDetails.AppName, appInstanceDetails.Instances)
	fmt.Printf("{ 'name' : '%s', 'instance' : '%d'-'%d' }", targetQ, appInstanceMap[targetQ].Instances, appInstanceDetails.Instances)

	/*
	   command := "curl /v2/apps/" + appInstanceDetails.AppGuid
	   payload := fmt.Sprintf("{ \"instances\": %d }", appInstanceDetails.Instances)
	*/
}

func (serv AppAutoScalerService) ScaleUp(postdata string, targetQ string, increments int) {
	serv.Scale(targetQ, increments, true)
}

func (serv AppAutoScalerService) ScaleDown(postdata string, targetQ string, increments int) {
	serv.Scale(targetQ, increments, false)
}
