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
	DEBUG_LOCK          = false
	listen_port         string
	appDetailsMapMutex  = &sync.RWMutex{}
	appInstanceMapMutex = &sync.RWMutex{}
	appDetailsMap       map[string]scalerutils.AppDetail
	appInstanceMap      map[string]scalerutils.AppInstanceDetail
)

// Save app details with expiration of 60 seconds
// reload the app instance data if request comes for app details after the duration 
const cache_expiration = 60

func init() {
	if listen_port = os.Getenv("PORT"); listen_port == "" {
		listen_port = "8080"
	}
	
	if debug_lock := os.Getenv("DEBUG_LOCK"); debug_lock != "" {
		DEBUG_LOCK = true
	}

	fmt.Printf("App Auto Scaler Properties\n")
	fmt.Printf("  Listen Port: %s !!\n", listen_port)
	fmt.Printf("  Instance default increment set to: %d !!\n\n", unit_increment)
	fmt.Printf("  Instance Cache Expiration set to: %d seconds !!\n", cache_expiration)
	
	appDetailsMap = make(map[string]scalerutils.AppDetail)
	appInstanceMap = make(map[string]scalerutils.AppInstanceDetail)

	// Load persisted information into memory
	// We are only persisting and reloading app data (org/space/app name and not instances)
	// The app instance information can be stale and its better to get that from CF
	if (scalerutils.IsDBEnabled()) {
		appDetails := scalerutils.Load()
		for _, appDetail := range appDetails {
			register(appDetail, false)
		}
	}
}

func main() {

	//var appDetailsMapMutex = &sync.Mutex{}
	gorest.RegisterMarshaller("application/json", gorest.NewJSONMarshaller())
	gorest.RegisterService(new(AppAutoScalerService)) //Register our service
	http.Handle("/", gorest.Handle())
	http.ListenAndServe(":"+listen_port, nil)
	
	defer close()	
}

func close() {
	if (scalerutils.IsDBEnabled()) {
		scalerutils.CloseDB()
	}	
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

func lockDebug(lockType string, activity string, step string) {
	if (DEBUG_LOCK) {
		fmt.Printf("%s (%s)-Lock from %s\n", activity, lockType, step )
	}
}

func (serv AppAutoScalerService) Register(appDetail scalerutils.AppDetail) {
	register(appDetail, true)
}
	
// Private method to handle both new incoming registrations 
// as well as loading up saved app details 
// isNew is false for reloading previously persisted data
func register(appDetail scalerutils.AppDetail, isNew bool) {

	fmt.Printf("\nRegistering App ... %#v", appDetail)
		
	destnName := appDetail.Target

	lockDebug("AD Write", "register", "Locking")
	appDetailsMapMutex.Lock()
	appDetailsMap[destnName] = appDetail
	appDetailsMapMutex.Unlock()
	lockDebug("AD Write", "register", "Released")
	
	if (scalerutils.IsDBEnabled() && isNew) {
		scalerutils.Insert(appDetail)
	}
	

	// Get the actual instance count... from CF
	// A map of Guid and Instance count is returned by cfutils.FindApp, 
	// using 'guid' and 'instances' as keys
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

	fmt.Printf("\nApp: %s under org: %s, space: %s, app guid: %s, instances: %d\n", appDetail.AppName, appDetail.Org, appDetail.Space, appGuid, instances)

	appInstanceDetails := scalerutils.AppInstanceDetail{destnName, appDetail.AppName, appGuid, instances, time.Now().Unix()}

	lockDebug("AI Write", "register", "Locking")
	appInstanceMapMutex.Lock()
	appInstanceMap[destnName] = appInstanceDetails
	appInstanceMapMutex.Unlock()
	lockDebug("AI Write", "register", "Released")
}

func (serv AppAutoScalerService) Load(appDetails []scalerutils.AppDetail) {
	for _, appDetail := range appDetails {
		serv.Register(appDetail)
	}
}

func getDetail(target string) string {

	var output string
	curTime := time.Now().Unix()
	
	lockDebug("AI Read", "detail", "Locking")
	appInstanceMapMutex.RLock()	
	appInstance := appInstanceMap[target]	
	appInstanceMapMutex.RUnlock()
	lockDebug("AI Read", "detail", "Released")
	
	if (curTime - appInstance.LastFetchTime) < cache_expiration {
		
		output = fmt.Sprintf("{ 'queue' : '%s', 'app' : '%s', 'instances': '%d'}\n", target, appInstance.AppName, appInstance.Instances)
		
		return output
	}

	lockDebug("AD Read", "detail", "Locking")
    appDetailsMapMutex.RLock()
	appDetail := appDetailsMap[target]
	appDetailsMapMutex.RUnlock()
	lockDebug("AD Read", "detail", "Released")
	
	appGuidInstanceDetail := cfutils.FindApp(appDetail.Org, appDetail.Space, appDetail.AppName)

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
		
		// Acquire Lock before updating...
		lockDebug("AI Write", "update as not-found", "Locking")
		appInstanceMapMutex.Lock()
		appInstanceMap[target] = appInstance
		appInstanceMapMutex.Unlock()
		lockDebug("AI Write", "update as not-found", "Released")
		return output
	}

	// Update Last fetch time
	appInstance.LastFetchTime = time.Now().Unix()

	// Acquire Lock before updating...
	lockDebug("AI Write", "update detail", "Locking")
	appInstanceMapMutex.Lock()
	appInstanceMap[target] = appInstance
	appInstanceMapMutex.Unlock()
	lockDebug("AI Write", "update detail", "Released")

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

	lockDebug("AI Write", "deregister", "Locking")
	appInstanceMapMutex.Lock()
	deleteAppInstanceDetail(target)
	appInstanceMapMutex.Unlock()
	lockDebug("AI Write", "deregister", "Released")
	
	appDetail := appDetailsMap[target]
	if (scalerutils.IsDBEnabled()) {
		scalerutils.Delete(appDetail)
	}
	
	lockDebug("AD Write", "deregister", "Locking")
	appDetailsMapMutex.Lock()
	deleteAppDetail(target)
	appDetailsMapMutex.Unlock()
	lockDebug("AD Write", "deregister", "Released")
	
}

func (serv AppAutoScalerService) Scale(target string, increments int, up bool) {

	lockDebug("AI Write", "scale", "Locking")
	appInstanceMapMutex.Lock()
	appInstanceDetails := appInstanceMap[target]
	if isAppInstanceNil(appInstanceDetails) {
		fmt.Printf("Error!! App Instance not found for %s\n", target)
		appInstanceMapMutex.Unlock()
		lockDebug("AI Write", "scale", "Released")
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
	lockDebug("AI Write", "scale", "Released")

	fmt.Printf("\nScaling[%s] App : { 'queue' : '%s', 'app' : '%s', 'instances': '%d'}\n", direction, target, appInstanceDetails.AppName, appInstanceDetails.Instances)

	cfutils.ScaleApp(appInstanceDetails.AppName, appInstanceDetails.Instances)
}

func (serv AppAutoScalerService) ScaleUp(postdata string, target string, increments int) {
	serv.Scale(target, increments, true)
}

func (serv AppAutoScalerService) ScaleDown(postdata string, target string, increments int) {
	serv.Scale(target, increments, false)
}
