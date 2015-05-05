package uaac

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

var uaaEnv UAAEnvironment

func init() {

	scheme := "https"
	verifySSL := false
	uaaHost := "uaa"
	domain := os.Getenv("DOMAIN")
	uaaClientID := os.Getenv("UAA_CLIENT_ID")
	uaaClientSecret := os.Getenv("UAA_CLIENT_SECRET")

	uaaEnv = UAAEnvironment{
		Domain:          domain,
		Scheme:          scheme,
		UAAClientID:     uaaClientID,
		UAAClientSecret: uaaClientSecret,
		UAAHost:         uaaHost,
		VerifySSL:       verifySSL,
	}

	fmt.Printf("\nDEBUG: UAA Environment created: %#v\n", uaaEnv)
}

func UaaEnvironment() UAAEnvironment {
	return uaaEnv
}

func UAAClient() (Token, error) {
	token := NewToken()

	params := url.Values{
		//"grant_type":   {"authorization_code"},
		"grant_type": {"client_credentials"},
	}

	uaaTokenURL := "https://" + uaaEnv.UAAHost +
					 "." + uaaEnv.Domain + "/oauth/token"

	uri, err := url.Parse(uaaTokenURL)
	if err != nil {
		return token, err
	}

	host := uri.Scheme + "://" + uri.Host

	fmt.Printf("\nDEBUG: Basic Auth UAA Client connecting to Host: %s\n", host)

	client := NewClient(host, uaaEnv.VerifySSL)
	client = client.WithBasicAuthCredentials(uaaEnv.UAAClientID, uaaEnv.UAAClientSecret)
	
	_, body, err := client.MakeRequest("POST", uri.RequestURI(), 
	                          strings.NewReader(params.Encode()))
	if err != nil {
		fmt.Printf("\nERROR! Error connecting to UAA: %s\n" + err.Error())
		return token, err
	}

	json.Unmarshal(body, &token)
	fmt.Printf("\nSuccessfully retreived auth token!\n")
	return token, nil
}
