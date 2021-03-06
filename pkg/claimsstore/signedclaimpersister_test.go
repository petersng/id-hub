package claimsstore_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joincivil/go-common/pkg/article"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/did/ethuri"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	"github.com/joincivil/id-hub/pkg/testutils"
)

func setupConnection() (*gorm.DB, error) {
	db, err := testutils.GetTestDBConnection()
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(
		&ethuri.PostgresDocument{},
		&claimsstore.SignedClaimPostgres{},
		&claimsstore.Node{},
		&claimsstore.RootCommit{},
		&claimsstore.JWTClaimPostgres{}).Error
	if err != nil {
		return nil, err
	}

	return db, nil
}

func makeContentCredential() *claimtypes.ContentCredential {
	subj := claimtypes.ContentCredentialSubject{
		ID:       "https://ap.com/article/1",
		Metadata: article.Metadata{},
	}
	proof := linkeddata.Proof{
		Type:       "EcdsaSecp256k1Signature2019",
		Creator:    "did:ethuri:apethuriabcd1234",
		Created:    time.Date(2000, 2, 1, 12, 30, 0, 0, time.UTC),
		ProofValue: "0xthisisasignatureandisprobablysuperlong",
	}
	proofSlice := make([]interface{}, 0)
	proofSlice = append(proofSlice, proof)
	return &claimtypes.ContentCredential{
		Context:           []string{"https://something.com/some/stuff/v1"},
		Type:              []claimtypes.CredentialType{claimtypes.VerifiableCredentialType, claimtypes.ContentCredentialType},
		CredentialSubject: subj,
		Issuer:            "did:ethuri:apethuriabcd1234",
		IssuanceDate:      time.Date(2000, 2, 1, 12, 30, 0, 0, time.UTC),
		Proof:             proofSlice,
	}
}

func TestToCredential(t *testing.T) {
	credentialSubject := json.RawMessage(`{
		"id": "https://ap.com/article/1",
		"metadata": {
			"title": "an article of great importance",
			"revisionContentHash": "0xabcdefg",
			"revisionContentUrl": "https://ap.com/article/1",
			"canonicalUrl": "https://ap.com/article-great-importance",
			"slug": "article-great-importance",
			"description": "the most important thing ever written on the subject",
			"contributors": [
				{
					"role": "author",
					"name": "Janine Great-Journalist"
				}
			],
			"images": [
				{
					"w": 64,
					"h": 64,
					"url": "https://ap.com/images/1.jpg",
					"hash": "0xthisimagewasimportant"
				}
			],
			"tags": [
				"abc",
				"fgh"
			],
			"primaryTag": "abc",
			"revisionDate": "2010-01-01T19:23:24Z",
			"originalPublishDate": "2009-12-12T20:00:00Z",
			"opinion": false,
			"civilSchemaVersion": "v0.0.1"
		}
	}`)
	proof := json.RawMessage(`{
		"type": "EcdsaSecp256k1Signature2019",
		"creator": "did:ethuri:apethuriabcd1234",
		"created": "2019-01-01T19:23:24Z",
		"proofValue": "0xthisisasignatureandisprobablysuperlong",
		"nonce": "some-unique-value-1"
	}`)
	signedClaim := &claimsstore.SignedClaimPostgres{
		IssuanceDate:      time.Now(),
		Type:              claimtypes.ContentCredentialType,
		CredentialSubject: postgres.Jsonb{RawMessage: credentialSubject},
		Issuer:            "did:ethuri:apethuriabcd1234",
		Proof:             postgres.Jsonb{RawMessage: proof},
		Hash:              "0xblahblahblah",
	}

	cred, err := signedClaim.ToCredential()
	if err != nil {
		t.Errorf("failed to convert the claim: %v", err)
	}
	cc, ok := cred.(*claimtypes.ContentCredential)
	if !ok {
		t.Errorf("should have been a content credential")
	}
	if cc.CredentialSubject.Metadata.Title != "an article of great importance" {
		t.Errorf("didn't hydrate right")
	}
}

func TestFromContentCredential(t *testing.T) {
	cred := makeContentCredential()
	signedClaim := &claimsstore.SignedClaimPostgres{}

	err := signedClaim.FromContentCredential(cred)

	if err != nil {
		t.Errorf("Error creating claim: %v", err)
	}

	if signedClaim.Type != claimtypes.ContentCredentialType {
		t.Errorf("wrong claim type")
	}
	if signedClaim.Hash != "1b2098bc8d129cae501aa1aa4fe8d92d3452a1e18b72267f77c354aba2b1e609196a" {
		t.Errorf("hash does not match expected")
	}
}

func TestSignedClaimPersister(t *testing.T) {
	db, err := setupConnection()
	if err != nil {
		t.Errorf("failed to set up db connection")
	}

	cleaner := testutils.DeleteCreatedEntities(db)
	defer cleaner()

	persister := claimsstore.NewSignedClaimPGPersister(db)
	cred := makeContentCredential()

	hash, err := persister.AddCredential(cred)
	if err != nil {
		t.Errorf("error adding claim: %v", err)
	}

	retrievedCred, err := persister.GetCredentialByHash(hash[4:])
	if err != nil {
		t.Errorf("error getting claim: %v", err)
	}
	_, err = retrievedCred.FindLinkedDataProof()
	if err != nil {
		t.Errorf("error retrieving linked data proof from slice: %v", err)
	}
	retrievedCred2, err := persister.GetCredentialByMultihash(hash)
	if err != nil {
		t.Errorf("error getting claim: %v", err)
	}
	_, err = retrievedCred2.FindLinkedDataProof()
	if err != nil {
		t.Errorf("error retrieving linked data proof from slice: %v", err)
	}
}
