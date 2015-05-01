package cf_utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var api_endpoint string
var orgGuidMap = make(map[string]string)
var spaceGuidMap = make(map[string]string)

//var appGuidMap map[string]string

func init() {

	// json parsing will get screwed up if CF_TRACE is enabled that would emit non-json contents

	//appGuidMap := make(map[string]string)

	os.Setenv("CF_TRACE", "false")

	var out bytes.Buffer
	cmd := exec.Command("cf", "target")
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		log.Fatal("CF Target not set", err)
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

func invoke_cf_curl(resource string) []byte {

	EXECUTABLE := "cf"
	COMMAND := "curl"
	VERSION := "/v2"
	//ACTION     := " -X GET"

	uri := VERSION + resource

	DEBUG := false

	var out bytes.Buffer
	cmd := exec.Command(EXECUTABLE, COMMAND, uri)
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
	}

	//var response utils.SpaceResponse
	//var response utils.SpaceSummaryResponse
	var byteArr []byte
	byteArr = out.Bytes()

	if DEBUG {
		fmt.Printf("\nInvoked Uri: %s\n", uri)
		fmt.Printf("Output data: %s\n", byteArr)
	}

	return byteArr
}

func FindApp(orgName string, spaceName string, appName string) map[string]string {

	var appGuidInstances map[string]string

	orgGuid := orgGuidMap[orgName]
	if orgGuid == "" {
		orgGuid = findOrgGuid(orgName)
		orgGuidMap[orgName] = orgGuid
	}

	if orgGuid != "" {
		spaceGuid := spaceGuidMap[spaceName]

		if spaceGuid == "" {
			spaceGuid = findSpaceGuid(spaceName, orgGuid)
			spaceGuidMap[spaceName] = spaceGuid
		}

		if spaceGuid != "" {
			appGuidInstances = findApp(appName, spaceGuid)
			return appGuidInstances
		}
	}

    fmt.Printf("\nError!! App could not be located!\n")
	return appGuidInstances
}

func findOrgGuid(orgName string) string {
	byteArr := invoke_cf_curl("/organizations")

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
	byteArr := invoke_cf_curl("/organizations/" + orgGuid + "/spaces")
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
	byteArr := invoke_cf_curl("/spaces/" + spaceGuid + "/apps")
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

  byteArr := invoke_cf_curl("/apps/" + appGuid + "/summary" )

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
