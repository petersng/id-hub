// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package graphql

import (
	"time"

	"github.com/joincivil/id-hub/pkg/utils"
)

type Claim struct {
	ID      *string  `json:"id"`
	Context *string  `json:"context"`
	Types   []string `json:"types"`
}

type ClaimGetRequestInput struct {
	ID *string `json:"id"`
}

type ClaimGetResponse struct {
	Claim *Claim `json:"claim"`
}

type DidDocAuthenticationInput struct {
	PublicKey *DidDocPublicKeyInput `json:"publicKey"`
	IDOnly    *bool                 `json:"idOnly"`
}

type DidDocPublicKeyInput struct {
	ID                 *string `json:"id"`
	Type               *string `json:"type"`
	Controller         *string `json:"controller"`
	PublicKeyPem       *string `json:"publicKeyPem"`
	PublicKeyJwk       *string `json:"publicKeyJwk"`
	PublicKeyHex       *string `json:"publicKeyHex"`
	PublicKeyBase64    *string `json:"publicKeyBase64"`
	PublicKeyBase58    *string `json:"publicKeyBase58"`
	PublicKeyMultibase *string `json:"publicKeyMultibase"`
	EthereumAddress    *string `json:"ethereumAddress"`
}

type DidDocServiceInput struct {
	ID              *string         `json:"id"`
	Type            *string         `json:"type"`
	Description     *string         `json:"description"`
	PublicKey       *string         `json:"publicKey"`
	ServiceEndpoint *utils.AnyValue `json:"serviceEndpoint"`
}

type DidGetRequestInput struct {
	Did *string `json:"did"`
}

type DidSaveRequestInput struct {
	Did             *string                      `json:"did"`
	PublicKeys      []*DidDocPublicKeyInput      `json:"publicKeys"`
	Authentications []*DidDocAuthenticationInput `json:"authentications"`
	Services        []*DidDocServiceInput        `json:"services"`
	Proof           *LinkedDataProofInput        `json:"proof"`
}

type LinkedDataProofInput struct {
	Type       *string    `json:"type"`
	Creator    *string    `json:"creator"`
	Created    *time.Time `json:"created"`
	ProofValue *string    `json:"proofValue"`
	Domain     *string    `json:"domain"`
	Nonce      *string    `json:"nonce"`
}
