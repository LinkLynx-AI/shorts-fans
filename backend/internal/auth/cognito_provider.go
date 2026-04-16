package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

// CognitoConfig は Cognito public API client の設定です。
type CognitoConfig struct {
	Region           string
	UserPoolClientID string
}

// Validate は Cognito client 設定が利用可能か検証します。
func (c CognitoConfig) Validate() error {
	if strings.TrimSpace(c.Region) == "" {
		return fmt.Errorf("region is required")
	}
	if strings.TrimSpace(c.UserPoolClientID) == "" {
		return fmt.Errorf("user pool client id is required")
	}

	return nil
}

type cognitoPublicAPI interface {
	ConfirmForgotPassword(ctx context.Context, params *cognitoidentityprovider.ConfirmForgotPasswordInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ConfirmForgotPasswordOutput, error)
	ConfirmSignUp(ctx context.Context, params *cognitoidentityprovider.ConfirmSignUpInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ConfirmSignUpOutput, error)
	ForgotPassword(ctx context.Context, params *cognitoidentityprovider.ForgotPasswordInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ForgotPasswordOutput, error)
	InitiateAuth(ctx context.Context, params *cognitoidentityprovider.InitiateAuthInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.InitiateAuthOutput, error)
	ResendConfirmationCode(ctx context.Context, params *cognitoidentityprovider.ResendConfirmationCodeInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ResendConfirmationCodeOutput, error)
	SignUp(ctx context.Context, params *cognitoidentityprovider.SignUpInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.SignUpOutput, error)
}

// CognitoClient は fan auth が使う Cognito public API を包みます。
type CognitoClient struct {
	api          cognitoPublicAPI
	userPoolID   string
	userClientID string
	now          func() time.Time
}

// NewCognitoClient は AWS SDK を使って Cognito client を構築します。
func NewCognitoClient(ctx context.Context, cfg CognitoConfig) (*CognitoClient, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	return newCognitoClient(cognitoidentityprovider.NewFromConfig(awsCfg), cfg.UserPoolClientID), nil
}

func newCognitoClient(api cognitoPublicAPI, userClientID string) *CognitoClient {
	return &CognitoClient{
		api:          api,
		userClientID: strings.TrimSpace(userClientID),
		now:          time.Now,
	}
}

// SignIn は Cognito username-password auth を実行し、principal を返します。
func (c *CognitoClient) SignIn(ctx context.Context, email string, password string) (CognitoSessionInput, error) {
	if c == nil || c.api == nil {
		return CognitoSessionInput{}, fmt.Errorf("cognito client is not initialized")
	}

	output, err := c.api.InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: cognitotypes.AuthFlowTypeUserPasswordAuth,
		ClientId: &c.userClientID,
		AuthParameters: map[string]string{
			"USERNAME": email,
			"PASSWORD": password,
		},
	})
	if err != nil {
		return CognitoSessionInput{}, err
	}
	if output.AuthenticationResult == nil || output.AuthenticationResult.IdToken == nil {
		return CognitoSessionInput{}, fmt.Errorf("cognito sign in did not return an id token")
	}

	return parseCognitoSessionInput(*output.AuthenticationResult.IdToken, c.now().UTC())
}

// SignUp は Cognito sign up を開始します。
func (c *CognitoClient) SignUp(ctx context.Context, email string, password string) error {
	if c == nil || c.api == nil {
		return fmt.Errorf("cognito client is not initialized")
	}

	_, err := c.api.SignUp(ctx, &cognitoidentityprovider.SignUpInput{
		ClientId: &c.userClientID,
		Username: &email,
		Password: &password,
		UserAttributes: []cognitotypes.AttributeType{
			{Name: stringPointer("email"), Value: &email},
		},
	})

	return err
}

// ConfirmSignUp は Cognito sign up confirm を実行します。
func (c *CognitoClient) ConfirmSignUp(ctx context.Context, email string, confirmationCode string) error {
	if c == nil || c.api == nil {
		return fmt.Errorf("cognito client is not initialized")
	}

	_, err := c.api.ConfirmSignUp(ctx, &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         &c.userClientID,
		Username:         &email,
		ConfirmationCode: &confirmationCode,
	})

	return err
}

// ResendSignUpCode は確認コードの再送を要求します。
func (c *CognitoClient) ResendSignUpCode(ctx context.Context, email string) error {
	if c == nil || c.api == nil {
		return fmt.Errorf("cognito client is not initialized")
	}

	_, err := c.api.ResendConfirmationCode(ctx, &cognitoidentityprovider.ResendConfirmationCodeInput{
		ClientId: &c.userClientID,
		Username: &email,
	})

	return err
}

// StartPasswordReset は forgot password flow を開始します。
func (c *CognitoClient) StartPasswordReset(ctx context.Context, email string) error {
	if c == nil || c.api == nil {
		return fmt.Errorf("cognito client is not initialized")
	}

	_, err := c.api.ForgotPassword(ctx, &cognitoidentityprovider.ForgotPasswordInput{
		ClientId: &c.userClientID,
		Username: &email,
	})

	return err
}

// ConfirmPasswordReset は confirmation code つき password reset を確定します。
func (c *CognitoClient) ConfirmPasswordReset(ctx context.Context, email string, confirmationCode string, newPassword string) error {
	if c == nil || c.api == nil {
		return fmt.Errorf("cognito client is not initialized")
	}

	_, err := c.api.ConfirmForgotPassword(ctx, &cognitoidentityprovider.ConfirmForgotPasswordInput{
		ClientId:         &c.userClientID,
		Username:         &email,
		ConfirmationCode: &confirmationCode,
		Password:         &newPassword,
	})

	return err
}

func parseCognitoSessionInput(idToken string, fallback time.Time) (CognitoSessionInput, error) {
	parts := strings.Split(strings.TrimSpace(idToken), ".")
	if len(parts) != 3 {
		return CognitoSessionInput{}, fmt.Errorf("cognito id token is malformed")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return CognitoSessionInput{}, fmt.Errorf("decode cognito id token payload: %w", err)
	}

	var claims map[string]json.RawMessage
	if err := json.Unmarshal(payload, &claims); err != nil {
		return CognitoSessionInput{}, fmt.Errorf("unmarshal cognito id token payload: %w", err)
	}

	subject, err := parseStringClaim(claims, "sub")
	if err != nil {
		return CognitoSessionInput{}, err
	}
	email, err := parseStringClaim(claims, "email")
	if err != nil {
		return CognitoSessionInput{}, err
	}
	emailVerified, err := parseBoolClaim(claims, "email_verified")
	if err != nil {
		return CognitoSessionInput{}, err
	}

	authenticatedAt := fallback.UTC()

	return CognitoSessionInput{
		Subject:         subject,
		Email:           email,
		EmailVerified:   emailVerified,
		AuthenticatedAt: authenticatedAt,
	}, nil
}

func parseStringClaim(claims map[string]json.RawMessage, name string) (string, error) {
	raw, ok := claims[name]
	if !ok {
		return "", fmt.Errorf("cognito id token missing %s claim", name)
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("cognito id token %s claim is invalid: %w", name, err)
	}
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("cognito id token %s claim is empty", name)
	}

	return value, nil
}

func parseBoolClaim(claims map[string]json.RawMessage, name string) (bool, error) {
	raw, ok := claims[name]
	if !ok {
		return false, fmt.Errorf("cognito id token missing %s claim", name)
	}

	var value bool
	if err := json.Unmarshal(raw, &value); err == nil {
		return value, nil
	}

	var stringValue string
	if err := json.Unmarshal(raw, &stringValue); err == nil {
		switch strings.TrimSpace(strings.ToLower(stringValue)) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		}
	}

	return false, fmt.Errorf("cognito id token %s claim is invalid", name)
}

func stringPointer(value string) *string {
	return &value
}
