package ethuri

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/linkeddata"
)

// GenerateDIDCli is the logic to handle the generatedid command for CLI
func GenerateDIDCli(pubKeyType linkeddata.SuiteType, pubKeyFile string, didPersister Persister) (*did.Document, error) {
	pubKeyValue, err := pubKeyFromFile(pubKeyFile)
	if err != nil {
		return nil, errors.Wrap(err, "error getting key from file")
	}

	firstPK := &did.DocPublicKey{
		Type:         pubKeyType,
		PublicKeyHex: &pubKeyValue,
	}

	doc, err := GenerateNewDocument(firstPK, true, true)
	if err != nil {
		return nil, errors.Wrap(err, "error initializing new did document")
	}

	bys, err := json.MarshalIndent(doc, "", "    ")
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling document for output")
	}

	fmt.Printf("-- DID --\n")
	fmt.Printf("%v\n", string(bys))

	if didPersister != nil {
		err = didPersister.SaveDocument(doc)
		if err != nil {
			return nil, errors.Wrap(err, "error storing new did to persister")
		}
	}

	return doc, nil
}

func pubKeyFromFile(filename string) (string, error) {
	bys, err := ioutil.ReadFile(filepath.Clean(filename))
	if err != nil {
		return "", err
	}

	key := strings.Trim(string(bys), "\n ")
	return key, nil
}
