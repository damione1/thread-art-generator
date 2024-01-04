package token

import (
	"testing"
	"time"

	"github.com/Damione1/thread-art-generator/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestPasetoMaker(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	userID := util.RandomUUID()
	duration := time.Minute

	IssuedTime := time.Now()
	ExpireTime := time.Now().Add(duration)

	token, payload, err := maker.CreateToken(userID, duration)

	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err = maker.ValidateToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, userID, payload.UserID)
	require.WithinDuration(t, IssuedTime, payload.IssuedTime, time.Second)
	require.WithinDuration(t, ExpireTime, payload.ExpireTime, time.Second)
}
