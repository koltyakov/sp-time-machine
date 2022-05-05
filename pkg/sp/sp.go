package sp

import (
	"github.com/koltyakov/gosip"
	"github.com/koltyakov/gosip/api"
	strategy "github.com/koltyakov/gosip/auth/saml" // see more https://go.spflow.com/auth/overview
)

/**

Depending on the SharePoint environment and use case, auth strategy (https://go.spflow.com/auth/strategies)
can be different. For a production installation Azure Certificate Auth
(https://go.spflow.com/auth/custom-auth/azure-certificate-auth) might be preferred.

*/

// Binds SharePoint API client
func NewSP(connStr, masterKey string) (*api.SP, error) {
	auth, err := parseConnStr(connStr, masterKey)
	if err != nil {
		return nil, err
	}
	client := &gosip.SPClient{AuthCnfg: auth}
	sp := api.NewSP(client)
	return sp, nil
}

func parseConnStr(connStr, masterKey string) (*strategy.AuthCnfg, error) {
	auth := &strategy.AuthCnfg{}
	auth.SetMasterkey(masterKey)
	auth.ParseConfig([]byte(connStr))
	return auth, nil
}
