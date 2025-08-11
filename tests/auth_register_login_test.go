package tests

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kirinyoku/sso-grpc/tests/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/kirinyoku/sso-grpc/internal/proto/auth"
)

const (
	emptyAppID int32  = 0
	appID      int32  = 1
	appSecret  string = "test-secret"

	passDefaultLength int = 16
)

func TestRegisterLogin_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, false, passDefaultLength)

	respReg, err := st.AuthClient.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respLog, err := st.AuthClient.Login(ctx, &pb.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    appID,
	})
	require.NoError(t, err)

	loginTime := time.Now()

	token := respLog.GetToken()
	require.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int32(claims["app_id"].(float64)))
	assert.Equal(t, respReg.GetUserId(), int64(claims["user_id"].(float64)))

	const deltaSeconds = 1

	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)
}

func TestRegisterLogin_DuplicateRegistration(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, false, passDefaultLength)

	respReg, err := st.AuthClient.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respReg, err = st.AuthClient.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.Error(t, err)
	assert.Empty(t, respReg.GetUserId())
	assert.ErrorContains(t, err, "user already exists")
}

func TestRegister_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "register with empty password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "password is required",
		},
		{
			name:        "register with empty email",
			email:       "",
			password:    gofakeit.Password(true, true, true, true, false, passDefaultLength),
			expectedErr: "email is required",
		},
		{
			name:        "register with both empty email and password",
			email:       "",
			password:    "",
			expectedErr: "email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &pb.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestLogin_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		appID       int32
		expectedErr string
	}{
		{
			name:        "login with empty password",
			email:       gofakeit.Email(),
			password:    "",
			appID:       appID,
			expectedErr: "password is required",
		},
		{
			name:        "login with empty email",
			email:       "",
			password:    gofakeit.Password(true, true, true, true, false, passDefaultLength),
			appID:       appID,
			expectedErr: "email is required",
		},
		{
			name:        "login with both empty email and password",
			email:       "",
			password:    "",
			appID:       appID,
			expectedErr: "email is required",
		},
		{
			name:        "login with non matching password",
			email:       gofakeit.Email(),
			password:    gofakeit.Password(true, true, true, true, false, passDefaultLength),
			appID:       appID,
			expectedErr: "invalid credentials",
		},
		{
			name:        "login with empty appID",
			email:       gofakeit.Email(),
			password:    gofakeit.Password(true, true, true, true, false, passDefaultLength),
			appID:       emptyAppID,
			expectedErr: "app_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Login(ctx, &pb.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
				AppId:    tt.appID,
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}
