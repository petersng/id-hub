package claimsstore

import (
	"encoding/hex"

	"github.com/dgrijalva/jwt-go"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jinzhu/gorm"
	"github.com/joincivil/id-hub/pkg/didjwt"
	"github.com/multiformats/go-multihash"
	didlib "github.com/ockam-network/did"
	"github.com/pkg/errors"
)

// JWTClaimPostgres is a model for storing jwt type vcs
type JWTClaimPostgres struct {
	JWT      string `gorm:"not null"`
	Issuer   string `gorm:"not null;index:jwtissuer"`
	Subject  string
	Sender   string `gorm:"not null;index:jwtsender"`
	Hash     string `gorm:"primary_key"`
	Data     string
	Type     string
	IssuedAt int64
}

// TableName sets the name of the table in the db
func (JWTClaimPostgres) TableName() string {
	return "jwt_claims"
}

// TokenToJWTClaimPostgres turns a jwt token into the db model
func TokenToJWTClaimPostgres(token *jwt.Token) (*JWTClaimPostgres, error) {
	hash, err := hashJWT(token.Raw)
	if err != nil {
		return nil, errors.Wrap(err, "failed to hash token")
	}

	claims, ok := token.Claims.(*didjwt.VCClaimsJWT)
	if !ok {
		return nil, errors.New("invalid claims type")
	}

	typ, ok := token.Header["typ"]
	if !ok {
		return nil, errors.New("no type on jwt")
	}

	typS, ok := typ.(string)
	if !ok {
		return nil, errors.New("type is not a string")
	}

	return &JWTClaimPostgres{
		JWT:      token.Raw,
		Issuer:   claims.Issuer,
		Hash:     hash,
		Data:     claims.Data,
		Subject:  claims.Subject,
		IssuedAt: claims.IssuedAt,
		Type:     typS,
	}, nil
}

func hashJWT(token string) (string, error) {
	hash := crypto.Keccak256([]byte(token))
	mHash, err := multihash.EncodeName(hash, "keccak-256")
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(mHash), nil
}

// JWTClaimPGPersister is a postgres persister for JWT claims
type JWTClaimPGPersister struct {
	db            *gorm.DB
	didJWTService *didjwt.Service
}

// NewJWTClaimPGPersister returns a new JWTClaimPGPersister
func NewJWTClaimPGPersister(db *gorm.DB, didJWTService *didjwt.Service) *JWTClaimPGPersister {
	return &JWTClaimPGPersister{
		db:            db,
		didJWTService: didJWTService,
	}
}

// AddJWT adds a new jwt claim to the db
func (p *JWTClaimPGPersister) AddJWT(tokenString string, senderDID *didlib.DID) (*jwt.Token, string, error) {
	token, err := p.didJWTService.ParseJWT(tokenString)
	if err != nil {
		return nil, "", errors.Wrap(err, "addJWT failed to parse token")
	}

	claim, err := TokenToJWTClaimPostgres(token)

	if err != nil {
		return nil, "", errors.Wrap(err, "addJWT failed to make model from token")
	}

	claim.Sender = senderDID.String()

	if err := p.db.Create(claim).Error; err != nil {
		return nil, "", errors.Wrap(err, "addJWT failed to save token to db")
	}

	return token, claim.Hash, nil
}

// GetJWTByHash returns a jwt from it's hash
func (p *JWTClaimPGPersister) GetJWTByHash(hash string) (*jwt.Token, error) {
	bytes, err := hex.DecodeString(hash)
	if err != nil {
		return nil, errors.Wrap(err, "GetJWTByHash failed to decode hash")
	}

	mHash, err := multihash.EncodeName(bytes, "keccak-256")
	if err != nil {
		return nil, errors.Wrap(err, "GetJWTByHash failed to create multihash")
	}

	mHashString := hex.EncodeToString(mHash)
	return p.GetJWTByMultihash(mHashString)
}

// GetJWTByMultihash returns a jwt from it's multihash
func (p *JWTClaimPGPersister) GetJWTByMultihash(mHash string) (*jwt.Token, error) {
	jwtClaim := &JWTClaimPostgres{}
	if err := p.db.Where(&JWTClaimPostgres{Hash: mHash}).First(jwtClaim).Error; err != nil {
		return nil, errors.Wrap(err, "GetJWTByMultihash failed to find claim")
	}
	return p.didJWTService.ParseJWT(jwtClaim.JWT)
}
