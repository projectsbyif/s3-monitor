package main

import (
	"crypto"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/glog"
	tcrypto "github.com/google/trillian/crypto"
	"github.com/google/trillian/crypto/keys/der"
	"github.com/projectsbyif/verifiable-cloudtrail/trillianlambda"
)

func main() {
	data, err := ioutil.ReadFile(os.Args[1])

	if err != nil {
		glog.Errorf("Unable to read file: %v", err)
	}

	var publishedData trillianlambda.LogRootVerificationData

	e := json.Unmarshal(data, &publishedData)
	if e != nil {
		glog.Errorf("Unable to unmarshal json: %v", e)
	}

	keyDER, _ := pem.Decode([]byte(publishedData.PublicKey))
	publicKey, err := der.UnmarshalPublicKey(keyDER.Bytes)
	if err != nil {
		glog.Errorf("Unable to Unmarshal public key: %v", err)
	}

	logRoot, err := tcrypto.VerifySignedLogRoot(publicKey, crypto.SHA256, &publishedData.SignedLogRoot)

	if err != nil {
		glog.Errorf("Unable to verify signed Log Root: %v", err)
	} else {
		fmt.Println("Success!")
		fmt.Printf("Log Root: %v \n", logRoot)
	}

}
