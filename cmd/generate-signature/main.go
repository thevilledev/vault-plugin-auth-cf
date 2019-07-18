package main

/*

Usage:

	export CF_INSTANCE_CERT=path/to/instance.crt
	export CF_INSTANCE_KEY=path/to/instance.key
	export SIGNING_TIME=$(date -u)
	export ROLE='test-role'
	generate-signature

To use it for directly logging into Vault:

	export CF_INSTANCE_CERT=path/to/instance.crt
	export CF_INSTANCE_KEY=path/to/instance.key
	export SIGNING_TIME=$(date -u)
	export ROLE='test-role'
	vault write auth/vault-plugin-auth-pcf/login \
		role=$ROLE \
		certificate=$CF_INSTANCE_CERT \
		signing_time=SIGNING_TIME \
		signature=$(generate-signature)
*/

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/hashicorp/vault-plugin-auth-pcf/signatures"
)

func main() {
	signingTimeRaw := os.Getenv("SIGNING_TIME")
	signingTime, err := time.Parse(signatures.TimeFormat, signingTimeRaw)
	if err != nil {
		log.Fatal(err)
	}

	pathToInstanceCert := os.Getenv("CF_INSTANCE_CERT")
	pathToInstanceKey := os.Getenv("CF_INSTANCE_KEY")
	roleName := os.Getenv("ROLE")

	instanceCertBytes, err := ioutil.ReadFile(pathToInstanceCert)
	if err != nil {
		log.Fatal(err)
	}

	signature, err := signatures.Sign(pathToInstanceKey, &signatures.SignatureData{
		SigningTime:            signingTime,
		CFInstanceCertContents: string(instanceCertBytes),
		Role:                   roleName,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Print(signature)

	sigData := &signatures.SignatureData{
		SigningTime:            signingTime,
		Role:                   roleName,
		CFInstanceCertContents: string(instanceCertBytes),
	}
	if _, err := signatures.Verify(signature, sigData); err != nil {
		log.Fatal(err)
	}
}
