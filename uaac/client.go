package uaac

import (
    "crypto/tls"
	"fmt"
    "io"
    "io/ioutil"
    "net/http"
    "sync"
)

var _client *http.Client
var mutex sync.Mutex

// Http Client, wraps go's http.Client for our usecase
type Client struct {
    Host              string
    BasicAuthUsername string
    BasicAuthPassword string
    AccessToken       string
    VerifySSL         bool
}

func NewClient(host string, verifySSL bool) Client {
    return Client{
        Host:      host,
        VerifySSL: verifySSL,
    }
}

func (client Client) WithBasicAuthCredentials(clientID, clientSecret string) Client {
    client.BasicAuthUsername = clientID
    client.BasicAuthPassword = clientSecret
    client.AccessToken = ""
    return client
}

func (client Client) WithAuthorizationToken(token string) Client {
    client.BasicAuthUsername = ""
    client.BasicAuthPassword = ""
    client.AccessToken = token
    return client
}

func GetClient(client Client) *http.Client {
    mutex.Lock()
    defer mutex.Unlock()

    if _client == nil {
        _client = &http.Client{
            Transport: &http.Transport{
                TLSClientConfig: client.TLSConfig(),
            },
        }
    }

    return _client
}

// Make request with the given basic auth and ssl settings, returns reponse code and body as a byte array
func (client Client) MakeRequest(method, path string, requestBody io.Reader) (int, []byte, error) {
    url := client.Host + path
    request, err := http.NewRequest(method, url, requestBody)
    if err != nil {
        return 0, nil, err
    }
    if client.BasicAuthUsername != "" {
        request.SetBasicAuth(client.BasicAuthUsername, client.BasicAuthPassword)
    } else {
        request.Header.Set("Authorization", "Bearer "+client.AccessToken)
    }
    request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	fmt.Printf("\nConnecting to UAA with Request: %#v\n", request)
    httpClient := GetClient(client)
    response, err := httpClient.Do(request)
    if err != nil {
        return 0, nil, err
    }
	
	fmt.Printf("\nGot UAA Response: %#v\n", response)

    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return response.StatusCode, body, err
    }

    return response.StatusCode, body, nil
}

func (client Client) TLSConfig() *tls.Config {
    return &tls.Config{
        InsecureSkipVerify: !client.VerifySSL,
    }
}