package uaac

//import "fmt"

type UAAEnvironment struct {
	Domain              string
	Scheme              string `env:"SCHEME"               `
	UAAClientID         string `env:"UAA_CLIENT_ID"        `
	UAAClientSecret     string `env:"UAA_CLIENT_SECRET"    `
	UAAHost             string `env:"UAA_HOST"             `
	VerifySSL           bool   `env:"VERIFY_SSL"           `
}

/*
func String() string {
	return fmt.Sprintf("{\nDomain: %s, Scheme : %s, UAAClientID: %s, UAAClientSecret: %s, UAAHost: %s, VerifySSL: %s}\n",  Domain, Scheme, UAAClientID, UAAClientSecret, UAAHost, VerifySSL)		
}
*/