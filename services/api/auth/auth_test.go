package auth

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/stretchr/testify/require"
)

const privateKeyStr = `-----BEGIN PRIVATE KEY-----
MIGkAgEBBDAOkc4RjYFA+bpiwJOQKaDjva78VJ5fVk/HZdn0ZavMdO21i+Cg6eys
fD0IQ7B6e6mgBwYFK4EEACKhZANiAAR8nU6p2MvooI6IYi/ZHPZySipcYKvKabp/
jAa5nM2glcGjrWAiXhoWPK4MkpMnFkyl3rKNrc95DrAmJHmFtm2wPj8ErLNC+lsT
R8lCXWTxI56c2fiN7n84jD1kFkQb7Ao=
-----END PRIVATE KEY-----`

func TestTokenGeneration(t *testing.T) {
	r := require.New(t)
	block, _ := pem.Decode([]byte(privateKeyStr))
	x509Encoded := block.Bytes
	privateKey, err := x509.ParseECPrivateKey(x509Encoded)
	r.NoError(err, "cannot decode private key")
	s := ServiceImp{
		privateKey: privateKey,
	}

	beforeGeneration := time.Now()
	userId := uuid.New()
	token, err := s.generateToken(userId)
	r.NoError(err, "cannot decode private key")

	claims, err := s.VerifyToken(context.Background(), token.Access)
	r.NoError(err, "cannot decode private key")

	r.Equal(userId.String(), claims.UserID)
	r.LessOrEqual(beforeGeneration.Add(accessTokenExpirationTime).Unix(), claims.ExpiresAt)
	r.Less(claims.ExpiresAt, beforeGeneration.Add(accessTokenExpirationTime+time.Minute).Unix())
}
