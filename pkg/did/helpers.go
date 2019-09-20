package did

import (
	"encoding/hex"
	"fmt"
	"time"

	log "github.com/golang/glog"
	"github.com/google/uuid"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/crypto"
	didlib "github.com/ockam-network/did"
)

// GenerateNewDocument generates a new DID and DID document and sets the first public
// key for the DID. Sets the public key into the publicKeys field and adds a
// reference to the key to the authentication field. If firstPK has an empty ID
// field, will populate it with the new DID.
func GenerateNewDocument(firstPK *DocPublicKey, addRefToAuth bool,
	addFragment bool) (*Document, error) {

	newDID, err := GenerateEthURIDID()
	if err != nil {
		return nil, errors.Wrap(err, "error generating new ethuri did")
	}

	// If firstPK ID is not set, then set it to the newly created DID
	if firstPK.ID == nil {
		firstPK.ID = newDID
		firstPK.Controller = CopyDID(newDID)
	}

	if !ValidDocPublicKey(firstPK) {
		return nil, errors.New("invalid doc public key")
	}

	doc, err := InitializeNewDocument(newDID, firstPK, addRefToAuth, addFragment)
	if err != nil {
		return nil, errors.Wrap(err, "error initializing new did document")
	}

	return doc, nil
}

// GenerateEthURIDID generates a new EthURI method DID
func GenerateEthURIDID() (*didlib.DID, error) {
	// Generate a new UUID v4
	newUUID := uuid.New()
	didStr := fmt.Sprintf("%s:%s", EthURISchemeMethod, newUUID.String())
	return didlib.Parse(didStr)
}

// InitializeNewDocument generates a simple version of a DID document given
// the DID and an initial public key.
func InitializeNewDocument(did *didlib.DID, firstPK *DocPublicKey, addRefToAuth bool,
	addFragment bool) (*Document, error) {
	if !ValidDocPublicKey(firstPK) {
		return nil, errors.New("invalid doc public key")
	}

	created := time.Now().UTC()
	updated := time.Now().UTC()

	doc := &Document{
		Context:         DefaultDIDContextV1,
		ID:              *did,
		Controller:      did,
		PublicKeys:      []DocPublicKey{},
		Authentications: []DocAuthenicationWrapper{},
		Services:        []DocService{},
		Created:         &created,
		Updated:         &updated,
	}

	err := doc.AddPublicKey(firstPK, addRefToAuth, addFragment)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// CopyDID is a convenience function to make a copy a DID struct
func CopyDID(d *didlib.DID) *didlib.DID {
	cpy, err := didlib.Parse(d.String())
	if err != nil {
		log.Errorf("Error parsing did string for copy: err: %v", err)
	}
	return cpy
}

// ValidDid returns true if the given did string is of a valid DID format
func ValidDid(did string) bool {
	_, err := didlib.Parse(did)
	return err == nil
}

// ValidDocPublicKey ensures that the given DocPublicKey is of a supported type,
// has a valid key for that type and is using the correct public key field.
// Returns true if it is valid, false if not.
func ValidDocPublicKey(pk *DocPublicKey) bool {
	// Supports only Secp256k1 hex keys for now
	// TODO(PN): Add more support here based on our needs
	goodKey := false
	switch pk.Type {
	case LDSuiteTypeSecp256k1Verification:
		if pk.PublicKeyHex == nil || *pk.PublicKeyHex == "" {
			log.Errorf("publicKeyHex is not populated for SECP256k1")
			return false
		}
		// Hex keys do not have 0x prefix
		bys, err := hex.DecodeString(*pk.PublicKeyHex)
		if err != nil {
			log.Errorf("unable to decode pub key hex str for SECP256k1: err: %v", err)
			return false
		}
		// Try to unmarshal to ensure valid key
		_, err = crypto.UnmarshalPubkey(bys)
		if err != nil {
			log.Errorf("invalid pub key value for SECP256k1: err: %v", err)
			return false
		}
		goodKey = true
	}

	// Controller is required for public keys
	if pk.Controller == nil || pk.Controller.String() == "" {
		log.Errorf("controller is required for public key")
		return false
	}

	if pk.ID == nil || pk.ID.String() == "" {
		log.Errorf("id is required for public key")
		return false
	}

	return true && goodKey
}

// PublicKeyInSlice checks to see if a DocPublicKey is in a slice of DocPublicKeys
// XXX(PN): ugh, should be an easier way to handle the key fields here.
func PublicKeyInSlice(pk DocPublicKey, pks []DocPublicKey) bool {
	keyFields := []*string{
		pk.PublicKeyPem,
		pk.PublicKeyJwk,
		pk.PublicKeyHex,
		pk.PublicKeyBase64,
		pk.PublicKeyBase58,
		pk.PublicKeyMultibase,
		pk.EthereumAddress,
	}

	var sKeyFields []*string
	var skf *string

	for _, sPk := range pks {
		sKeyFields = []*string{
			sPk.PublicKeyPem,
			sPk.PublicKeyJwk,
			sPk.PublicKeyHex,
			sPk.PublicKeyBase64,
			sPk.PublicKeyBase58,
			sPk.PublicKeyMultibase,
			sPk.EthereumAddress,
		}

		if pk.Type == sPk.Type {
			for ind, kf := range keyFields {
				skf = sKeyFields[ind]
				if kf != nil && *kf != "" && *kf == *skf {
					return true
				}

			}
		}
	}
	return false
}

// AuthInSlice checks to see if a DocAuthenticationWrapper is in a slice of
// DocAuthenticationWrapper
// XXX(PN): ugh, there should be an easier way to handle the key fields here.
func AuthInSlice(auth DocAuthenicationWrapper, auths []DocAuthenicationWrapper) bool {
	if auth.IDOnly {
		for _, sAuth := range auths {
			if auth.DocPublicKey.ID.String() == sAuth.DocPublicKey.ID.String() {
				return true
			}
		}
		return false
	}

	keyFields := []*string{
		auth.DocPublicKey.PublicKeyPem,
		auth.DocPublicKey.PublicKeyJwk,
		auth.DocPublicKey.PublicKeyHex,
		auth.DocPublicKey.PublicKeyBase64,
		auth.DocPublicKey.PublicKeyBase58,
		auth.DocPublicKey.PublicKeyMultibase,
		auth.DocPublicKey.EthereumAddress,
	}

	var sKeyFields []*string
	var skf *string
	var authKey DocPublicKey
	var sAuthKey DocPublicKey

	for _, sAuth := range auths {
		sKeyFields = []*string{
			sAuth.DocPublicKey.PublicKeyPem,
			sAuth.DocPublicKey.PublicKeyJwk,
			sAuth.DocPublicKey.PublicKeyHex,
			sAuth.DocPublicKey.PublicKeyBase64,
			sAuth.DocPublicKey.PublicKeyBase58,
			sAuth.DocPublicKey.PublicKeyMultibase,
			sAuth.DocPublicKey.EthereumAddress,
		}

		authKey = auth.DocPublicKey
		sAuthKey = sAuth.DocPublicKey

		if authKey.Type == sAuthKey.Type {
			for ind, kf := range keyFields {
				skf = sKeyFields[ind]
				if kf != nil && *kf != "" && *kf == *skf {
					return true
				}

			}
		}
	}
	return false
}

// ServiceInSlice checks to see if a DocService is in a slice of
// DocService
func ServiceInSlice(srv DocService, srvs []DocService) bool {
	var matches bool

	for _, sSrv := range srvs {
		if srv.Type == sSrv.Type {
			if srv.ServiceEndpointURI != nil && sSrv.ServiceEndpointURI != nil &&
				*srv.ServiceEndpointURI == *sSrv.ServiceEndpointURI {
				return true

			} else if srv.ServiceEndpointLD != nil && sSrv.ServiceEndpointLD != nil {
				if len(sSrv.ServiceEndpointLD) == len(srv.ServiceEndpointLD) {
					return true
				}

				matches = true
				for key := range sSrv.ServiceEndpointLD {
					if srv.ServiceEndpointLD[key] != sSrv.ServiceEndpointLD[key] {
						matches = false
					}
				}

				if matches {
					return true
				}
			}
		}
	}
	return false
}
