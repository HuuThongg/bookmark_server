package auth

import (
	"bookmark/util"
	"errors"
	"fmt"
	"log"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrTokenExpired = errors.New("token is expired")

type PayLoad struct {
	ID        string             `json:"id"`
	AccountID int64              `json:"account_id"`
	IssuedAt  pgtype.Timestamptz `json:"issued_at"`
	Expiry    pgtype.Timestamptz `json:"expiry"`
}

func newPayload(accountID int64, issued_at, expiry time.Time) *PayLoad {
	return &PayLoad{
		ID:        uuid.NewString(),
		AccountID: accountID,
		IssuedAt:  util.ToTimestamptz(issued_at),
		Expiry:    util.ToTimestamptz(expiry),
	}
}

func CreateToken(accountID int64, issued_at time.Time, duration time.Duration) (string, *PayLoad, error) {
	expiry := time.Now().UTC().Add(duration)
	payload := newPayload(accountID, issued_at, expiry)

	token := paseto.NewToken()
	token.SetExpiration(expiry)
	token.SetIssuedAt(payload.IssuedAt.Time)
	token.Set("payload", payload)

	config, err := util.LoadConfig(".")
	if err != nil {
		return "", nil, nil
	}

	secretKey, err := paseto.NewV4AsymmetricSecretKeyFromHex(config.SecretKeyHex)

	if err != nil {
		return "", nil, errors.New("secretKey creation errror")
	}

	signed := token.V4Sign(secretKey, nil)

	return signed, payload, nil

}

func VerifyToken(signed string) (*PayLoad, error) {
	config, err := util.LoadConfig(".")
	if err != nil {
		return nil, nil
	}
	publicKey, err := paseto.NewV4AsymmetricPublicKeyFromHex(config.PublicKeyHex)
	if err != nil {
		err := errors.New("failed to create publicKey")
		return nil, err
	}

	parser := paseto.NewParser()
	fmt.Println("publicKey", publicKey)
	fmt.Println("signed: ", signed)
	token, err := parser.ParseV4Public(publicKey, signed, nil)
	if err != nil {
		log.Printf("token parse error: %v", err.Error())
		return nil, err
	}
	fmt.Println("token parse: ", token)
	var payload PayLoad
	if err := token.Get("payload", &payload); err != nil {
		log.Printf("get payload error: %v", err.Error())
		return nil, err
	}
	if time.Now().UTC().After(payload.Expiry.Time) {
		return nil, ErrTokenExpired
	}
	fmt.Println("payload get")
	return &payload, nil
}
