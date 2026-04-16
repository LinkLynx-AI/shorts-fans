package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

type cognitoPublicAPIStub struct {
	confirmForgotPassword func(context.Context, *cognitoidentityprovider.ConfirmForgotPasswordInput, ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ConfirmForgotPasswordOutput, error)
	confirmSignUp         func(context.Context, *cognitoidentityprovider.ConfirmSignUpInput, ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ConfirmSignUpOutput, error)
	forgotPassword        func(context.Context, *cognitoidentityprovider.ForgotPasswordInput, ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ForgotPasswordOutput, error)
	initiateAuth          func(context.Context, *cognitoidentityprovider.InitiateAuthInput, ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.InitiateAuthOutput, error)
	resendConfirmation    func(context.Context, *cognitoidentityprovider.ResendConfirmationCodeInput, ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ResendConfirmationCodeOutput, error)
	signUp                func(context.Context, *cognitoidentityprovider.SignUpInput, ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.SignUpOutput, error)
}

func (s cognitoPublicAPIStub) ConfirmForgotPassword(
	ctx context.Context,
	params *cognitoidentityprovider.ConfirmForgotPasswordInput,
	optFns ...func(*cognitoidentityprovider.Options),
) (*cognitoidentityprovider.ConfirmForgotPasswordOutput, error) {
	return s.confirmForgotPassword(ctx, params, optFns...)
}

func (s cognitoPublicAPIStub) ConfirmSignUp(
	ctx context.Context,
	params *cognitoidentityprovider.ConfirmSignUpInput,
	optFns ...func(*cognitoidentityprovider.Options),
) (*cognitoidentityprovider.ConfirmSignUpOutput, error) {
	return s.confirmSignUp(ctx, params, optFns...)
}

func (s cognitoPublicAPIStub) ForgotPassword(
	ctx context.Context,
	params *cognitoidentityprovider.ForgotPasswordInput,
	optFns ...func(*cognitoidentityprovider.Options),
) (*cognitoidentityprovider.ForgotPasswordOutput, error) {
	return s.forgotPassword(ctx, params, optFns...)
}

func (s cognitoPublicAPIStub) InitiateAuth(
	ctx context.Context,
	params *cognitoidentityprovider.InitiateAuthInput,
	optFns ...func(*cognitoidentityprovider.Options),
) (*cognitoidentityprovider.InitiateAuthOutput, error) {
	return s.initiateAuth(ctx, params, optFns...)
}

func (s cognitoPublicAPIStub) ResendConfirmationCode(
	ctx context.Context,
	params *cognitoidentityprovider.ResendConfirmationCodeInput,
	optFns ...func(*cognitoidentityprovider.Options),
) (*cognitoidentityprovider.ResendConfirmationCodeOutput, error) {
	return s.resendConfirmation(ctx, params, optFns...)
}

func (s cognitoPublicAPIStub) SignUp(
	ctx context.Context,
	params *cognitoidentityprovider.SignUpInput,
	optFns ...func(*cognitoidentityprovider.Options),
) (*cognitoidentityprovider.SignUpOutput, error) {
	return s.signUp(ctx, params, optFns...)
}

func TestCognitoConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     CognitoConfig
		wantErr string
	}{
		{
			name: "valid",
			cfg: CognitoConfig{
				Region:           "ap-northeast-1",
				UserPoolClientID: "client-id",
			},
		},
		{
			name: "missing region",
			cfg: CognitoConfig{
				UserPoolClientID: "client-id",
			},
			wantErr: "region is required",
		},
		{
			name: "missing client id",
			cfg: CognitoConfig{
				Region: "ap-northeast-1",
			},
			wantErr: "user pool client id is required",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.cfg.Validate()
			if tt.wantErr == "" && err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("Validate() error got %v want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestNewCognitoClientTrimsClientID(t *testing.T) {
	t.Parallel()

	client := newCognitoClient(cognitoPublicAPIStub{}, " client-id ")
	if client == nil {
		t.Fatal("newCognitoClient() = nil, want initialized client")
	}
	if client.userClientID != "client-id" {
		t.Fatalf("newCognitoClient() client id got %q want %q", client.userClientID, "client-id")
	}
}

func TestNewCognitoClientValidatesConfigBeforeAWSLoad(t *testing.T) {
	t.Parallel()

	if _, err := NewCognitoClient(context.Background(), CognitoConfig{}); err == nil || !strings.Contains(err.Error(), "region is required") {
		t.Fatalf("NewCognitoClient() error got %v want config validation error", err)
	}
}

func TestNewCognitoClientBuildsAWSClient(t *testing.T) {
	t.Parallel()

	client, err := NewCognitoClient(context.Background(), CognitoConfig{
		Region:           "ap-northeast-1",
		UserPoolClientID: "client-id",
	})
	if err != nil {
		t.Fatalf("NewCognitoClient() error = %v, want nil", err)
	}
	if client == nil {
		t.Fatal("NewCognitoClient() = nil, want initialized client")
	}
	if client.userClientID != "client-id" {
		t.Fatalf("NewCognitoClient() client id got %q want %q", client.userClientID, "client-id")
	}
}

func TestCognitoClientMethodsRequireInitialization(t *testing.T) {
	t.Parallel()

	client := &CognitoClient{}
	ctx := context.Background()

	if _, err := client.SignIn(ctx, "fan@example.com", "VeryStrongPass123!"); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("SignIn() error got %v want initialization error", err)
	}
	if err := client.SignUp(ctx, "fan@example.com", "VeryStrongPass123!"); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("SignUp() error got %v want initialization error", err)
	}
	if err := client.ConfirmSignUp(ctx, "fan@example.com", "123456"); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("ConfirmSignUp() error got %v want initialization error", err)
	}
	if err := client.ResendSignUpCode(ctx, "fan@example.com"); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("ResendSignUpCode() error got %v want initialization error", err)
	}
	if err := client.StartPasswordReset(ctx, "fan@example.com"); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("StartPasswordReset() error got %v want initialization error", err)
	}
	if err := client.ConfirmPasswordReset(ctx, "fan@example.com", "123456", "AnotherStrongPass456!"); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("ConfirmPasswordReset() error got %v want initialization error", err)
	}
}

func TestCognitoClientSignInParsesIDToken(t *testing.T) {
	t.Parallel()

	fallback := time.Unix(1712000300, 0).UTC()
	client := newCognitoClient(cognitoPublicAPIStub{
		initiateAuth: func(_ context.Context, params *cognitoidentityprovider.InitiateAuthInput, _ ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.InitiateAuthOutput, error) {
			if params.AuthFlow != cognitotypes.AuthFlowTypeUserPasswordAuth {
				t.Fatalf("InitiateAuth() auth flow got %q want %q", params.AuthFlow, cognitotypes.AuthFlowTypeUserPasswordAuth)
			}
			if params.ClientId == nil || *params.ClientId != "client-id" {
				t.Fatalf("InitiateAuth() client id got %v want %q", params.ClientId, "client-id")
			}
			if got := params.AuthParameters["USERNAME"]; got != "fan@example.com" {
				t.Fatalf("InitiateAuth() username got %q want %q", got, "fan@example.com")
			}
			if got := params.AuthParameters["PASSWORD"]; got != "VeryStrongPass123!" {
				t.Fatalf("InitiateAuth() password got %q want %q", got, "VeryStrongPass123!")
			}

			idToken := testIDToken(t, map[string]any{
				"sub":            "cognito-subject",
				"email":          "fan@example.com",
				"email_verified": "true",
			})
			return &cognitoidentityprovider.InitiateAuthOutput{
				AuthenticationResult: &cognitotypes.AuthenticationResultType{
					IdToken: &idToken,
				},
			}, nil
		},
	}, "client-id")
	client.now = func() time.Time { return fallback }

	got, err := client.SignIn(context.Background(), "fan@example.com", "VeryStrongPass123!")
	if err != nil {
		t.Fatalf("SignIn() error = %v, want nil", err)
	}
	if got.Subject != "cognito-subject" {
		t.Fatalf("SignIn() subject got %q want %q", got.Subject, "cognito-subject")
	}
	if got.Email != "fan@example.com" {
		t.Fatalf("SignIn() email got %q want %q", got.Email, "fan@example.com")
	}
	if !got.EmailVerified {
		t.Fatal("SignIn() email verified got false want true")
	}
	if !got.AuthenticatedAt.Equal(fallback) {
		t.Fatalf("SignIn() authenticated at got %s want %s", got.AuthenticatedAt, fallback)
	}
}

func TestCognitoClientSignInRequiresIDToken(t *testing.T) {
	t.Parallel()

	client := newCognitoClient(cognitoPublicAPIStub{
		initiateAuth: func(context.Context, *cognitoidentityprovider.InitiateAuthInput, ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.InitiateAuthOutput, error) {
			return &cognitoidentityprovider.InitiateAuthOutput{}, nil
		},
	}, "client-id")

	if _, err := client.SignIn(context.Background(), "fan@example.com", "VeryStrongPass123!"); err == nil || !strings.Contains(err.Error(), "did not return an id token") {
		t.Fatalf("SignIn() error got %v want missing id token error", err)
	}
}

func TestCognitoClientWriteOperationsDelegateToAWSAPI(t *testing.T) {
	t.Parallel()

	client := newCognitoClient(cognitoPublicAPIStub{
		signUp: func(_ context.Context, params *cognitoidentityprovider.SignUpInput, _ ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.SignUpOutput, error) {
			if params.ClientId == nil || *params.ClientId != "client-id" {
				t.Fatalf("SignUp() client id got %v want %q", params.ClientId, "client-id")
			}
			if params.Username == nil || *params.Username != "fan@example.com" {
				t.Fatalf("SignUp() username got %v want %q", params.Username, "fan@example.com")
			}
			if params.Password == nil || *params.Password != "VeryStrongPass123!" {
				t.Fatalf("SignUp() password got %v want %q", params.Password, "VeryStrongPass123!")
			}
			if len(params.UserAttributes) != 1 || params.UserAttributes[0].Name == nil || *params.UserAttributes[0].Name != "email" {
				t.Fatalf("SignUp() user attributes got %+v want email attribute", params.UserAttributes)
			}
			return &cognitoidentityprovider.SignUpOutput{}, nil
		},
		confirmSignUp: func(_ context.Context, params *cognitoidentityprovider.ConfirmSignUpInput, _ ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ConfirmSignUpOutput, error) {
			if params.ConfirmationCode == nil || *params.ConfirmationCode != "123456" {
				t.Fatalf("ConfirmSignUp() confirmation code got %v want %q", params.ConfirmationCode, "123456")
			}
			return &cognitoidentityprovider.ConfirmSignUpOutput{}, nil
		},
		resendConfirmation: func(_ context.Context, params *cognitoidentityprovider.ResendConfirmationCodeInput, _ ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ResendConfirmationCodeOutput, error) {
			if params.Username == nil || *params.Username != "fan@example.com" {
				t.Fatalf("ResendConfirmationCode() username got %v want %q", params.Username, "fan@example.com")
			}
			return &cognitoidentityprovider.ResendConfirmationCodeOutput{}, nil
		},
		forgotPassword: func(_ context.Context, params *cognitoidentityprovider.ForgotPasswordInput, _ ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ForgotPasswordOutput, error) {
			if params.Username == nil || *params.Username != "fan@example.com" {
				t.Fatalf("ForgotPassword() username got %v want %q", params.Username, "fan@example.com")
			}
			return &cognitoidentityprovider.ForgotPasswordOutput{}, nil
		},
		confirmForgotPassword: func(_ context.Context, params *cognitoidentityprovider.ConfirmForgotPasswordInput, _ ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ConfirmForgotPasswordOutput, error) {
			if params.ConfirmationCode == nil || *params.ConfirmationCode != "123456" {
				t.Fatalf("ConfirmForgotPassword() confirmation code got %v want %q", params.ConfirmationCode, "123456")
			}
			if params.Password == nil || *params.Password != "AnotherStrongPass456!" {
				t.Fatalf("ConfirmForgotPassword() password got %v want %q", params.Password, "AnotherStrongPass456!")
			}
			return &cognitoidentityprovider.ConfirmForgotPasswordOutput{}, nil
		},
	}, "client-id")

	ctx := context.Background()
	if err := client.SignUp(ctx, "fan@example.com", "VeryStrongPass123!"); err != nil {
		t.Fatalf("SignUp() error = %v, want nil", err)
	}
	if err := client.ConfirmSignUp(ctx, "fan@example.com", "123456"); err != nil {
		t.Fatalf("ConfirmSignUp() error = %v, want nil", err)
	}
	if err := client.ResendSignUpCode(ctx, "fan@example.com"); err != nil {
		t.Fatalf("ResendSignUpCode() error = %v, want nil", err)
	}
	if err := client.StartPasswordReset(ctx, "fan@example.com"); err != nil {
		t.Fatalf("StartPasswordReset() error = %v, want nil", err)
	}
	if err := client.ConfirmPasswordReset(ctx, "fan@example.com", "123456", "AnotherStrongPass456!"); err != nil {
		t.Fatalf("ConfirmPasswordReset() error = %v, want nil", err)
	}
}

func TestParseCognitoSessionInputRejectsMalformedToken(t *testing.T) {
	t.Parallel()

	if _, err := parseCognitoSessionInput("not-a-jwt", time.Unix(1712000400, 0).UTC()); err == nil || !strings.Contains(err.Error(), "malformed") {
		t.Fatalf("parseCognitoSessionInput() error got %v want malformed token error", err)
	}
}

func TestParseStringClaimRejectsMissingClaim(t *testing.T) {
	t.Parallel()

	_, err := parseStringClaim(map[string]json.RawMessage{}, "sub")
	if err == nil || !strings.Contains(err.Error(), "missing sub claim") {
		t.Fatalf("parseStringClaim() error got %v want missing claim error", err)
	}
}

func TestParseBoolClaimRejectsInvalidValue(t *testing.T) {
	t.Parallel()

	_, err := parseBoolClaim(map[string]json.RawMessage{
		"email_verified": json.RawMessage(`"maybe"`),
	}, "email_verified")
	if err == nil || !strings.Contains(err.Error(), "is invalid") {
		t.Fatalf("parseBoolClaim() error got %v want invalid bool error", err)
	}
}

func testIDToken(t *testing.T, claims map[string]any) string {
	t.Helper()

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v, want nil", err)
	}

	return header + "." + base64.RawURLEncoding.EncodeToString(payload) + ".signature"
}
