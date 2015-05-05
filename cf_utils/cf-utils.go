package cf_utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	// "log"
	// "os/exec"
	"strconv"
	"strings"
	//"./uaac"
  "github.com/cf-platform-eng/autoscaler/cf_utils/uaac"
	"io/ioutil"
	"net/http"
	"crypto/tls"
)

var (
	auth_token string
	cf_api_endpoint string
	DEBUG bool = false
	uaaEnv uaac.UAAEnvironment
	orgGuidMap = make(map[string]string)
	spaceGuidMap = make(map[string]string)
	appGuidMap = make(map[string]string)
)

func init() {

	// json parsing will get screwed up if CF_TRACE is enabled that would emit non-json contents
	uaaEnv = uaac.UaaEnvironment()
	token, _ := uaac.UAAClient()
	
	// Token comes with braces: { ... } , strip that
	auth_token = fmt.Sprintf("%s", token)
	auth_token = strings.Replace(auth_token, "{", "", -1)
	auth_token = strings.Replace(auth_token, "}", "", -1)
	
	fmt.Sprintf("Auth Token: %s\n", auth_token)
	
	
	//cf_api_endpoint = uaaEnv.Scheme + "://api." + uaaEnv.Domain
	cf_api_endpoint = "https://api." + uaaEnv.Domain
}

/*
func setup_using_cli() {
	setupCfCli()
	setupCfTarget("api.10.244.0.34.xip.io", "admin", "admin") 

	os.Setenv("CF_TRACE", "false")

	var out bytes.Buffer
	cmd := exec.Command("cf", "target")
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		log.Fatal("CF Target not set \n", err)
		os.Exit(-1)
	}

	//var byteArr []byte
	targetStatus := string(out.Bytes())
	tokens := strings.Split(targetStatus, "\n")
	for _, token := range tokens {

		if strings.HasPrefix(token, "API endpoint") {
			api_endpoint = strings.Fields(token)[2]
			fmt.Println("API Endpoint : " + api_endpoint)
		} else if strings.HasPrefix(token, "User") {
			user := strings.Fields(token)[1]
			fmt.Println("Authenticated User : " + user)
		}
	}

	fmt.Println("Finished initializing cf client...\n")
}

func setupCfCli() {
	os.Setenv("APP_DIR", "/home/vcap/app")

	var out bytes.Buffer
	cmd := exec.Command("tar", "-C $APP_DIR/bin/ zxvf $APP_DIR/lib/cf-linux*.tgz")
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		log.Fatal("Unable to locate or extract cf binary tar ball \n", err)
		os.Exit(-1)
	}
	
	os.Setenv("PATH", "/home/vcap/app/bin:$PATH")	
}

func setupCfTarget(cf_api_target string, cf_user string, cf_passwd string) {
	
	os.Setenv("HOME", "/home/vcap")
	
	var out bytes.Buffer
	cmd := exec.Command("cf", " api" , cf_api_target, " --skip-ssl-validation")
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		log.Fatal("Unable to set CF API Target to endpoint \n", err)
		os.Exit(-1)
	}
	
	cmd = exec.Command("cf", " login" , "-u ", cf_user, " -p" , cf_passwd)
	cmd.Stdout = &out

	err = cmd.Run()

	if err != nil {
		log.Fatal("Unable to login into CF \n", err)
		os.Exit(-1)
	}
}

func invoke_cf_curl_get(resource string) []byte {

	
	//EXECUTABLE := "cf"
	//COMMAND := "curl"
	
	EXECUTABLE := "curl"
	OPTION    := " -k "
	ACTION     := " -X GET"
	
	HEADER1 := cf_curl_authorization_header()
	HEADER2 := "-H 'Content-Type: application/x-www-form-urlencoded' "
	
	ENDPOINT := cf_api_endpoint
	VERSION := "/v2"
			
	uri := ENDPOINT + VERSION + resource

	var out bytes.Buffer
	cmd := exec.Command(EXECUTABLE, OPTION, HEADER1, HEADER2, ACTION, uri)
	cmd.Stdout = &out

    fmt.Printf("\nCommand to Execute: %#v\n", cmd)
	err := cmd.Run()

	if err != nil {
		fmt.Printf("\nError with invocation of cf curl: %s\n", err.Error())
		log.Fatal(err)
	}

	//var response utils.SpaceResponse
	//var response utils.SpaceSummaryResponse
	var byteArr []byte
	byteArr = out.Bytes()
	DEBUG := true

	if DEBUG {
		fmt.Printf("\nInvoked Uri: %s\n", uri)
		fmt.Printf("\nCommand Executed: %#v\n", cmd)
		fmt.Printf("Output data: %s\n", byteArr)
	}

	return byteArr
}

func invoke_cf_curl_post(resource string, postData string) {

	//EXECUTABLE := "cf"
	//COMMAND := "curl"
	
	EXECUTABLE := "curl"
	OPTION    := " -k "
	ACTION     := " -X POST"
	
	HEADER := cf_curl_authorization_header()
	
	ENDPOINT := cf_api_endpoint
	VERSION := "/v2"
		
	uri := ENDPOINT + VERSION + resource
	payload := " -d " + postData

	var out bytes.Buffer
	cmd := exec.Command(EXECUTABLE, OPTION, ACTION, HEADER, uri, payload)
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
	}
}


*/

func UAAToken() string {
	return fmt.Sprintf("%s", auth_token)
	
}

func cf_curl_authorization_header() string{
	//return fmt.Sprintf(" -H 'Authorization: bearer %s'", UAAToken())
	return fmt.Sprintf(" -H 'Authorization: bearer %s'", UAAToken())
}

func set_headers(req *http.Request) {

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", " Bearer " + UAAToken())
}

func invoke_cf_get(resource string) []byte {
	
	ENDPOINT := cf_api_endpoint
	VERSION := "/v2"
	
	uri := ENDPOINT + VERSION + resource
	DEBUG := false
	
    req, err := http.NewRequest("GET", uri, bytes.NewReader([]byte (resource)))
	
	set_headers(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Set("Authorization", " Bearer " + UAAToken)
	
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify : true},
	}
    client := &http.Client{Transport: tr}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, _ := ioutil.ReadAll(resp.Body)
	if DEBUG {
		fmt.Println("Request :", req)
	    fmt.Println("Response Status:", resp.Status)
	    fmt.Println("Response Headers:", resp.Header)    	
    	fmt.Println("Response Body:", string(body))
	}
	return body	
}

func invoke_cf_with_payload(resource string, action string, postData string) {
	
	
	ENDPOINT := cf_api_endpoint
	VERSION := "/v2"
	
	uri := ENDPOINT + VERSION + resource
	DEBUG := false
	
    req, err := http.NewRequest(action, uri, bytes.NewReader([]byte (postData)))
	
	set_headers(req)
	req.Header.Set("Content-Type", "application/json")
	
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify : true},
	}
    client := &http.Client{Transport: tr}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    _, _ = ioutil.ReadAll(resp.Body)
	if DEBUG {
		fmt.Println("Request :", req)
	    fmt.Println("Response Status:", resp.Status)
	    fmt.Println("Response Headers:", resp.Header)    	
    	//fmt.Println("Response Body:", string(body))
	}
	return
}


func FindApp(orgName string, spaceName string, appName string) map[string]string {

	var appGuidInstances map[string]string

	orgGuid := orgGuidMap[orgName]
	if orgGuid == "" {
		orgGuid = findOrgGuid(orgName)
	}
	orgGuidMap[orgName] = orgGuid

	if orgGuid != "" {
		spaceGuid := spaceGuidMap[spaceName]

		if spaceGuid == "" {
			spaceGuid = findSpaceGuid(spaceName, orgGuid)			
		}
		spaceGuidMap[spaceName] = spaceGuid
		
		if spaceGuid != "" {
			appGuidInstances = findApp(appName, spaceGuid)
			appGuidMap[appName] = appGuidInstances["guid"]
			return appGuidInstances
		}
	}

    fmt.Printf("\nError!! App could not be located!\n")
	return appGuidInstances
}

func ScaleApp(appName string, instances int) {

    appGuid := appGuidMap[appName]
    payload := fmt.Sprintf("{ \"instances\": %d }", instances)
	
	invoke_cf_with_payload("/apps/" + appGuid, "PUT", payload)	    
}

func findOrgGuid(orgName string) string {
	byteArr := invoke_cf_get("/organizations")

	var errorResponse ErrorResponse
	json.Unmarshal(byteArr, &errorResponse)
	if errorResponse.Description == "" {

		var response QueryResponse
		json.Unmarshal(byteArr, &response)

		for _, org := range response.Resources {
			if orgName == org.Entity.Name {
				return org.Metadata.Guid
			}
		}
	}

	fmt.Printf("\nError!! Could not locate org matching: %s\n", orgName)
	return ""
}

func findSpaceGuid(spaceName string, orgGuid string) string {
	byteArr := invoke_cf_get("/organizations/" + orgGuid + "/spaces")
	var response QueryResponse
	json.Unmarshal(byteArr, &response)

	for _, space := range response.Resources {
		if spaceName == space.Entity.Name {
			return space.Metadata.Guid
		}
	}
	fmt.Printf("\nError!! Could not locate space matching: %s\n", spaceName)
	return ""
}

func findApp(appName string, spaceGuid string) map[string]string {
	var result = make(map[string]string)
	byteArr := invoke_cf_get("/spaces/" + spaceGuid + "/apps")
	var response SpaceAppQueryResponse
	json.Unmarshal(byteArr, &response)
	//fmt.Printf("\n\nRaw Apps output is %s\n", byteArr)

	for _, app := range response.Resources {

		if appName == app.Entity.Name {
			fmt.Printf("\nLocated App: %s, guid: %s, instances: %d\n", appName, app.Metadata.Guid, app.Entity.Instances)

			var result = make(map[string]string)
			result["guid"] = app.Metadata.Guid
			result["instances"] = strconv.Itoa(app.Entity.Instances)
			return result
		}
	}

	fmt.Printf("\nError!! Could not locate app matching: %s\n", appName)
	return result
}

/*
func FindAppInstances(appGuid string) int{
  fmt.Printf("Inside FindAppInstances....")

  byteArr := invoke_cf_curl_get("/apps/" + appGuid + "/summary" )

  fmt.Printf("App Summary raw output is %s\n", byteArr)
  //fmt.Printf("App Stats output is %v\n", response)
  //fmt.Println("Instance Details for App: ", response["0"])

  var appSummaryErrorResponse ErrorResponse
  json.Unmarshal(byteArr, &appSummaryErrorResponse)
  //fmt.Printf("App Error Stats output is %v\n", appStateErrorResponse)
  //fmt.Printf("Error description %s", appStateErrorResponse.Description)
  if (appSummaryErrorResponse.Description != "") {

	  var appDetailSummary AppDetail
	  json.Unmarshal(byteArr, &appDetailSummary)
	  fmt.Printf("Got App Summary : %v", appDetailSummary)

	  return appDetailSummary.Instances

  }
  return 0
}
*/
