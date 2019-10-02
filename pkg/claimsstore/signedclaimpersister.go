package claimsstore

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joincivil/id-hub/pkg/linkeddata"
)

// SignedClaimPostgres represents the schema for signed claims
type SignedClaimPostgres struct {
	IssuanceDate      time.Time
	Type              CredentialType
	CredentialSubject postgres.Jsonb `gorm:"not null"`
	Issuer            string         `gorm:"not null;index:issuer"`
	Proof             postgres.Jsonb `gorm:"not null"`
	Hash              string         `gorm:"primary_key"`
}

// TableName sets the table name for signed claims
func (SignedClaimPostgres) TableName() string {
	return "signed_claims"
}

// ToCredential converts the db type to the model
func (c *SignedClaimPostgres) ToCredential() (*ContentCredential, error) {
	if c.Type != ContentCredentialType {
		return nil, errors.New("Only content credential is currently implemented")
	}
	credential := &ContentCredential{
		Type:    []CredentialType{VerifiableCredentialType, ContentCredentialType},
		Context: []string{"https://www.w3.org/2018/credentials/v1", "https://id.civil.co/credentials/contentcredential/v1"},
		Issuer:  c.Issuer,
		CredentialSchema: CredentialSchema{
			ID:   "https://id.civil.co/credentials/schemas/v1/metadata.json",
			Type: "JsonSchemaValidator2018",
		},
	}
	proof := &linkeddata.Proof{}
	err := json.Unmarshal(c.Proof.RawMessage, proof)
	if err != nil {
		return nil, err
	}

	credential.Proof = *proof

	credSubj := &ContentCredentialSubject{}
	err = json.Unmarshal(c.CredentialSubject.RawMessage, credSubj)

	if err != nil {
		return nil, err
	}

	credential.CredentialSubject = *credSubj

	return credential, nil
}

// FromContentCredential populates the db type from a model
func (c *SignedClaimPostgres) FromContentCredential(cred *ContentCredential) error {
	c.Issuer = cred.Issuer
	c.IssuanceDate = cred.IssuanceDate
	c.Type = ContentCredentialType
	credSubjJSON, err := json.Marshal(cred.CredentialSubject)
	if err != nil {
		return err
	}
	proofJSON, err := json.Marshal(cred.Proof)
	if err != nil {
		return err
	}
	c.CredentialSubject = postgres.Jsonb{RawMessage: credSubjJSON}
	c.Proof = postgres.Jsonb{RawMessage: proofJSON}

	credJSON, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	c.Hash = hex.EncodeToString(crypto.Keccak256(credJSON))
	return nil
}

// SignedClaimPGPersister persister model for signed claims
type SignedClaimPGPersister struct {
	db *gorm.DB
}

// NewSignedClaimPGPersister returns a new SignedClaimPGPersister
func NewSignedClaimPGPersister(db *gorm.DB) *SignedClaimPGPersister {
	return &SignedClaimPGPersister{
		db: db,
	}
}

// AddCredential takes a credential and adds it to the db
func (p *SignedClaimPGPersister) AddCredential(claim *ContentCredential) (string, error) {
	signedClaim := &SignedClaimPostgres{}
	err := signedClaim.FromContentCredential(claim)
	if err != nil {
		return "", err
	}
	if err := p.db.Create(signedClaim).Error; err != nil {
		return "", err
	}
	return signedClaim.Hash, nil
}

// GetCredentialByHash returns a credential from a hash taken from the associated merkle tree claim
func (p *SignedClaimPGPersister) GetCredentialByHash(hash string) (*ContentCredential, error) {
	signedClaim := &SignedClaimPostgres{}
	if err := p.db.Where(&SignedClaimPostgres{Hash: hash}).First(signedClaim).Error; err != nil {
		return nil, err
	}
	return signedClaim.ToCredential()
}
