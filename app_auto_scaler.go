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
	"time"
	cfutils "github.com/cf-platform-eng/autoscaler/cf_utils"
	scalerutils "github.com/cf-platform-eng/autoscaler/scaler_utils"
)

var (
	api_endpoint        string
	unit_increment      = 1
	listen_port         string
	appDetailsMapMutex  = &sync.RWMutex{}
	appInstanceMapMutex = &sync.RWMutex{}
	appDetailsMap       map[string]scalerutils.AppDetail
	appInstanceMap      map[string]scalerutils.AppInstanceDetail
)

// On Cache Expiration of 1 min, reload the app instance data
const cache_expiration = 60

func init() {
	if listen_port = os.Getenv("PORT"); listen_port == "" {
		listen_port = "8080"
	}

	fmt.Printf("App Auto Scaler Properties\n")
	fmt.Printf("  Listen Port: %s !!\n", listen_port)
	fmt.Printf("  Instance default increment set to: %d !!\n\n", unit_increment)
	fmt.Printf("  Instance Cache Expiration set to: %d seconds !!\n", cache_expiration)
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

	fmt.Printf("\nRegistering App ... %#v", appDetail)
	destnName := appDetail.Target

	//fmt.Printf("Getting AD Write lock")
	appDetailsMapMutex.Lock()
	appDetailsMap[destnName] = appDetail
	appDetailsMapMutex.Unlock()
	
	if (scalerutils.DBEnabled()) {
		scalerutils.Insert(appDetail)
	}
	//fmt.Printf("Released AD Write lock")

	// Get the actual instance count...
	appGuidInstanceDetail := cfutils.FindApp(appDetail.Org, appDetail.Space, appDetail.AppName)

	var instances int
	var appGuid string
	if appGuid = appGuidInstanceDetail["guid"]; appGuid != "" {
		instanceCount, err := strconv.Atoi(appGuidInstanceDetail["instances"])

		if err != nil {
			fmt.Printf("Could not convert into int: %v\n", err)
		} else {
			instances = instanceCount
		}
	} else {
		fmt.Printf("\nWarning!! Failed to locate App : %s in Org: %s, Space: %s\n", appDetail.AppName, appDetail.Org, appDetail.Space)
		return
	}

	fmt.Printf("\nApp: %s under org: %s, space: %s, app guid: %s, instances: %d", appDetail.AppName, appDetail.Org, appDetail.Space, appGuid, instances)

	appInstanceDetails := scalerutils.AppInstanceDetail{destnName, appDetail.AppName, appGuid, instances, time.Now().Unix()}

	//fmt.Printf("Getting AI Write lock")
	appInstanceMapMutex.Lock()
	appInstanceMap[destnName] = appInstanceDetails
	appInstanceMapMutex.Unlock()
	
	
	//fmt.Printf("Released AI Write lock")
}

func (serv AppAutoScalerService) Load(appDetails []scalerutils.AppDetail) {
	for _, appDetail := range appDetails {
		serv.Register(appDetail)
	}
}

func getDetail(target string) string {

	var output string
	curTime := time.Now().Unix()
	appInstance := appInstanceMap[target]
	
	if (curTime - appInstance.LastFetchTime) < cache_expiration {

		//fmt.Printf("Getting AI Read lock")
		appInstanceMapMutex.RLock()
		output = fmt.Sprintf("{ 'queue' : '%s', 'app' : '%s', 'instances': '%d'}\n", target, appInstance.AppName, appInstance.Instances)
		appInstanceMapMutex.RUnlock()
		//fmt.Printf("Released AI Read lock")

		return output
	}

	appDetail := appDetailsMap[target]
	appGuidInstanceDetail := cfutils.FindApp(appDetail.Org, appDetail.Space, appDetail.AppName)

	// Acquire Lock before updating...
	//fmt.Printf("Getting AI Write lock")
	appInstanceMapMutex.Lock()
	//fmt.Printf("Got AI Write lock")

	if appGuid := appGuidInstanceDetail["guid"]; appGuid != "" {
		instanceCount, err := strconv.Atoi(appGuidInstanceDetail["instances"])

		if err != nil {
			fmt.Printf("Could not convert into int: %v\n", err)
			output = fmt.Sprintf("Error with Target: %s, could not covert instances count to int: %s", target, err)
		} else {

			appInstance.Instances = instanceCount
			output = fmt.Sprintf("{ 'queue' : '%s', 'app' : '%s', 'instances': '%d'}\n", target, appInstance.AppName, appInstance.Instances)
		}

	} else {
		appInstance.Instances = 0

		fmt.Printf("\nError!! Failed to locate App : %s in Org: %s, Space: %s\n", appDetail.AppName, appDetail.Org, appDetail.Space)

		output = fmt.Sprintf("\nError!! Failed to locate App : %s in Org: %s, Space: %s\n", appDetail.AppName, appDetail.Org, appDetail.Space)
	}

	// Update Last fetch time
	appInstance.LastFetchTime = time.Now().Unix()
	appInstanceMap[target] = appInstance

	// Release lock
	appInstanceMapMutex.Unlock()
	//fmt.Printf("Released AI Write lock")

	return output
}

func (serv AppAutoScalerService) Details() string {
	output := "["
	for key := range appInstanceMap {
		output += getDetail(key) + "\n"
	}

	output = output + "]"
	return output
}

func (serv AppAutoScalerService) QueueDetails(target string) string {

	return getDetail(target)
}

func findAppDetail(target string) scalerutils.AppDetail {

	appDetail := appDetailsMap[target]
	return appDetail
}

func deleteAppDetail(target string) {

	delete(appDetailsMap, target)
}

func findAppInstanceDetail(target string) scalerutils.AppInstanceDetail {

	appInstance := appInstanceMap[target]
	return appInstance
}

func deleteAppInstanceDetail(target string) {

	delete(appInstanceMap, target)
}

func isAppInstanceNil(appInstance scalerutils.AppInstanceDetail) bool {
	return appInstance.AppName == ""
}

func isAppNil(app scalerutils.AppDetail) bool {
	return app.AppName == ""
}

func (serv AppAutoScalerService) Deregister(target string) {
	//fmt.Printf("Getting AI Write lock")
	appInstanceMapMutex.Lock()
	deleteAppInstanceDetail(target)
	appInstanceMapMutex.Unlock()
	//fmt.Printf("Releasing AI Write lock")

	//fmt.Printf("Getting AD Write lock")
	appDetail := appDetailsMap[target]
	if (scalerutils.DBEnabled()) {
		scalerutils.Delete(appDetail)
	}
	
	appDetailsMapMutex.Lock()
	deleteAppDetail(target)
	appDetailsMapMutex.Unlock()
	//fmt.Printf("Releasing AD Write lock")
	
}

func (serv AppAutoScalerService) Scale(target string, increments int, up bool) {

	//fmt.Printf("Getting AI Write lock")
	appInstanceMapMutex.Lock()
	appInstanceDetails := appInstanceMap[target]
	if isAppInstanceNil(appInstanceDetails) {
		fmt.Printf("Error!! App Instance not found for %s\n", target)
		return
	}

	if increments == 0 {
		increments = unit_increment
	}

	direction := "Up"
	increment := increments
	if !up {
		increment = -increment
		direction = "Down"
	}

	appInstanceDetails.Instances += increment

	if appInstanceDetails.Instances < 0 {
		appInstanceDetails.Instances = 0
	}

	appInstanceMap[target] = appInstanceDetails
	appInstanceMapMutex.Unlock()
	//fmt.Printf("Releasing AI Write lock")

	fmt.Printf("\nScaling[%s] App : { 'queue' : '%s', 'app' : '%s', 'instances': '%d'}\n", direction, target, appInstanceDetails.AppName, appInstanceDetails.Instances)

	cfutils.ScaleApp(appInstanceDetails.AppName, appInstanceDetails.Instances)
}

func (serv AppAutoScalerService) ScaleUp(postdata string, target string, increments int) {
	serv.Scale(target, increments, true)
}

func (serv AppAutoScalerService) ScaleDown(postdata string, target string, increments int) {
	serv.Scale(target, increments, false)
}
