package pcf

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/vault-plugin-auth-pcf/signatures"
	"github.com/hashicorp/vault/api"
)

type CLIHandler struct{}

func (h *CLIHandler) Auth(c *api.Client, m map[string]string) (*api.Secret, error) {
	mount, ok := m["mount"]
	if !ok {
		mount = "pcf"
	}

	role := m["role"]
	if role == "" {
		return nil, errors.New(`"role" is required`)
	}

	pathToInstanceCert := m["cf_instance_cert"]
	if pathToInstanceCert == "" {
		pathToInstanceCert = os.Getenv(EnvVarInstanceCertificate)
	}
	if pathToInstanceCert == "" {
		return nil, errors.New(`"cf_instance_cert" is required`)
	}

	pathToInstanceKey := m["cf_instance_key"]
	if pathToInstanceKey == "" {
		pathToInstanceKey = os.Getenv(EnvVarInstanceKey)
	}
	if pathToInstanceKey == "" {
		return nil, errors.New(`"cf_instance_key" is required`)
	}

	certBytes, err := ioutil.ReadFile(pathToInstanceCert)
	if err != nil {
		return nil, err
	}
	certificate := string(certBytes)

	signingTime := time.Now().UTC()
	signatureData := &signatures.SignatureData{
		SigningTime: signingTime,
		Role:        role,
		Certificate: certificate,
	}
	signature, err := signatures.Sign(pathToInstanceKey, signatureData)
	if err != nil {
		return nil, err
	}

	loginData := map[string]interface{}{
		"role":         role,
		"certificate":  certificate,
		"signing_time": signingTime.Format(signatures.TimeFormat),
		"signature":    signature,
	}

	path := fmt.Sprintf("auth/%s/login", mount)

	secret, err := c.Logical().Write(path, loginData)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, errors.New("empty response from credential provider")
	}
	return secret, nil
}

func (h *CLIHandler) Help() string {
	help := `
Usage: vault login -method=pcf [CONFIG K=V...]

  The PCF auth method allows users to authenticate using PCF's instance identity service.

  The PCF credentials may be specified explicitly via the command line:

      $ vault login -method=pcf role=...

  This will automatically pull from the CF_INSTANCE_CERT and CF_INSTANCE_KEY values
  in your local environment. If they're not available or you wish to override them, 
  they may also be supplied explicitly:

      $ vault login -method=pcf role=... cf_instance_cert=... cf_instance_key=...

Configuration:

  cf_instance_cert=<string>
      Explicit value to use for the path to the PCF instance certificate.

  cf_instance_key=<string>
      Explicit value to use for the path to the PCF instance key.

  mount=<string>
      Path where the PCF credential method is mounted. This is usually provided
      via the -path flag in the "vault login" command, but it can be specified
      here as well. If specified here, it takes precedence over the value for
      -path. The default value is "pcf".

  role=<string>
      Name of the role to request a token against
`

	return strings.TrimSpace(help)
}
