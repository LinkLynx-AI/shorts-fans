package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aws/smithy-go"
	"github.com/google/uuid"
)

type fanAuthProviderStub struct {
	confirmPasswordReset func(context.Context, string, string, string) error
	confirmSignUp        func(context.Context, string, string) error
	resendSignUpCode     func(context.Context, string) error
	signIn               func(context.Context, string, string) (CognitoSessionInput, error)
	signUp               func(context.Context, string, string) error
	startPasswordReset   func(context.Context, string) error
}

func (s fanAuthProviderStub) ConfirmPasswordReset(ctx context.Context, email string, confirmationCode string, newPassword string) error {
	return s.confirmPasswordReset(ctx, email, confirmationCode, newPassword)
}

func (s fanAuthProviderStub) ConfirmSignUp(ctx context.Context, email string, confirmationCode string) error {
	return s.confirmSignUp(ctx, email, confirmationCode)
}

func (s fanAuthProviderStub) ResendSignUpCode(ctx context.Context, email string) error {
	return s.resendSignUpCode(ctx, email)
}

func (s fanAuthProviderStub) SignIn(ctx context.Context, email string, password string) (CognitoSessionInput, error) {
	return s.signIn(ctx, email, password)
}

func (s fanAuthProviderStub) SignUp(ctx context.Context, email string, password string) error {
	return s.signUp(ctx, email, password)
}

func (s fanAuthProviderStub) StartPasswordReset(ctx context.Context, email string) error {
	return s.startPasswordReset(ctx, email)
}

type fanAuthSessionManagerStub struct {
	startSession       func(context.Context, CognitoSessionInput) (AuthenticatedSession, error)
	startSignUpSession func(context.Context, CognitoSessionInput, string, string) (AuthenticatedSession, error)
}

func (s fanAuthSessionManagerStub) StartSession(ctx context.Context, input CognitoSessionInput) (AuthenticatedSession, error) {
	return s.startSession(ctx, input)
}

func (s fanAuthSessionManagerStub) StartSignUpSession(
	ctx context.Context,
	input CognitoSessionInput,
	displayName string,
	handle string,
) (AuthenticatedSession, error) {
	return s.startSignUpSession(ctx, input, displayName, handle)
}

type fanAuthRepositoryStub struct {
	getIdentityByProviderAndSubject func(context.Context, string, string) (Identity, error)
	getIdentityByNormalizedEmail func(context.Context, string) (Identity, error)
	getPreferredEmailByUserID    func(context.Context, uuid.UUID) (string, error)
	handleExists                 func(context.Context, string) (bool, error)
	revokeActiveSession          func(context.Context, string, time.Time) (SessionRecord, error)
	updateActiveModeByTokenHash  func(context.Context, string, ActiveMode) (SessionRecord, error)
}

func (s fanAuthRepositoryStub) GetIdentityByProviderAndSubject(ctx context.Context, provider string, providerSubject string) (Identity, error) {
	return s.getIdentityByProviderAndSubject(ctx, provider, providerSubject)
}

func (s fanAuthRepositoryStub) GetIdentityByNormalizedEmail(ctx context.Context, emailNormalized string) (Identity, error) {
	return s.getIdentityByNormalizedEmail(ctx, emailNormalized)
}

func (s fanAuthRepositoryStub) GetPreferredEmailByUserID(ctx context.Context, userID uuid.UUID) (string, error) {
	return s.getPreferredEmailByUserID(ctx, userID)
}

func (s fanAuthRepositoryStub) HandleExists(ctx context.Context, handle string) (bool, error) {
	return s.handleExists(ctx, handle)
}

func (s fanAuthRepositoryStub) RevokeActiveSessionByTokenHash(ctx context.Context, sessionTokenHash string, revokedAt time.Time) (SessionRecord, error) {
	return s.revokeActiveSession(ctx, sessionTokenHash, revokedAt)
}

func (s fanAuthRepositoryStub) UpdateActiveModeByTokenHash(
	ctx context.Context,
	sessionTokenHash string,
	activeMode ActiveMode,
) (SessionRecord, error) {
	return s.updateActiveModeByTokenHash(ctx, sessionTokenHash, activeMode)
}

type signUpDraftStoreStub struct {
	deleteDraft func(context.Context, string) error
	getDraft    func(context.Context, string) (SignUpDraft, error)
	saveDraft   func(context.Context, string, SignUpDraft, time.Duration) error
}

func (s signUpDraftStoreStub) DeleteDraft(ctx context.Context, email string) error {
	return s.deleteDraft(ctx, email)
}

func (s signUpDraftStoreStub) GetDraft(ctx context.Context, email string) (SignUpDraft, error) {
	return s.getDraft(ctx, email)
}

func (s signUpDraftStoreStub) SaveDraft(ctx context.Context, email string, draft SignUpDraft, ttl time.Duration) error {
	return s.saveDraft(ctx, email, draft, ttl)
}

type authCooldownStoreStub struct {
	release     func(context.Context, string) error
	tryActivate func(context.Context, string, time.Duration) (bool, error)
}

func (s authCooldownStoreStub) Release(ctx context.Context, key string) error {
	return s.release(ctx, key)
}

func (s authCooldownStoreStub) TryActivate(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return s.tryActivate(ctx, key, ttl)
}

type fanAuthViewerReaderStub struct {
	readCurrentViewer func(context.Context, string) (Bootstrap, error)
}

func (s fanAuthViewerReaderStub) ReadCurrentViewer(ctx context.Context, rawSessionToken string) (Bootstrap, error) {
	return s.readCurrentViewer(ctx, rawSessionToken)
}

func TestFanAuthServiceSignInUsesDraftProfileWhenPresent(t *testing.T) {
	t.Parallel()

	service := NewFanAuthService(
		fanAuthProviderStub{
			signIn: func(_ context.Context, email string, password string) (CognitoSessionInput, error) {
				if email != "fan@example.com" {
					t.Fatalf("SignIn() email got %q want %q", email, "fan@example.com")
				}
				if password != "VeryStrongPass123!" {
					t.Fatalf("SignIn() password got %q want %q", password, "VeryStrongPass123!")
				}

				return CognitoSessionInput{
					Subject:         "cognito-subject",
					Email:           email,
					EmailVerified:   true,
					AuthenticatedAt: time.Unix(1712000000, 0).UTC(),
				}, nil
			},
		},
		fanAuthSessionManagerStub{
			startSession: func(context.Context, CognitoSessionInput) (AuthenticatedSession, error) {
				t.Fatal("StartSession() should not be called when draft exists")
				return AuthenticatedSession{}, nil
			},
			startSignUpSession: func(_ context.Context, input CognitoSessionInput, displayName string, handle string) (AuthenticatedSession, error) {
				if input.Subject != "cognito-subject" {
					t.Fatalf("StartSignUpSession() subject got %q want %q", input.Subject, "cognito-subject")
				}
				if displayName != "Mina" {
					t.Fatalf("StartSignUpSession() displayName got %q want %q", displayName, "Mina")
				}
				if handle != "mina" {
					t.Fatalf("StartSignUpSession() handle got %q want %q", handle, "mina")
				}

				return AuthenticatedSession{Token: "raw-session-token"}, nil
			},
		},
		fanAuthRepositoryStub{
			getIdentityByProviderAndSubject: func(context.Context, string, string) (Identity, error) { return Identity{}, ErrIdentityNotFound },
			getIdentityByNormalizedEmail: func(context.Context, string) (Identity, error) { return Identity{}, ErrIdentityNotFound },
			getPreferredEmailByUserID:    func(context.Context, uuid.UUID) (string, error) { return "", nil },
			handleExists:                 func(context.Context, string) (bool, error) { return false, nil },
			revokeActiveSession:          func(context.Context, string, time.Time) (SessionRecord, error) { return SessionRecord{}, nil },
		},
		fanAuthViewerReaderStub{
			readCurrentViewer: func(context.Context, string) (Bootstrap, error) { return Bootstrap{}, nil },
		},
		signUpDraftStoreStub{
			getDraft: func(_ context.Context, email string) (SignUpDraft, error) {
				if email != "fan@example.com" {
					t.Fatalf("GetDraft() email got %q want %q", email, "fan@example.com")
				}
				return SignUpDraft{DisplayName: "Mina", Handle: "mina", Password: "VeryStrongPass123!"}, nil
			},
			deleteDraft: func(_ context.Context, email string) error {
				if email != "fan@example.com" {
					t.Fatalf("DeleteDraft() email got %q want %q", email, "fan@example.com")
				}
				return nil
			},
			saveDraft: func(context.Context, string, SignUpDraft, time.Duration) error { return nil },
		},
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	got, err := service.SignIn(context.Background(), "fan@example.com", "VeryStrongPass123!")
	if err != nil {
		t.Fatalf("SignIn() error = %v, want nil", err)
	}
	if got.Token != "raw-session-token" {
		t.Fatalf("SignIn() token got %q want %q", got.Token, "raw-session-token")
	}
}

func TestFanAuthServiceSignInUsesDraftProfileWhenDeleteDraftFails(t *testing.T) {
	t.Parallel()

	service := NewFanAuthService(
		fanAuthProviderStub{
			signIn: func(context.Context, string, string) (CognitoSessionInput, error) {
				return CognitoSessionInput{
					Subject:         "cognito-subject",
					Email:           "fan@example.com",
					EmailVerified:   true,
					AuthenticatedAt: time.Unix(1712000001, 0).UTC(),
				}, nil
			},
		},
		fanAuthSessionManagerStub{
			startSession: func(context.Context, CognitoSessionInput) (AuthenticatedSession, error) {
				t.Fatal("StartSession() should not be called when draft exists")
				return AuthenticatedSession{}, nil
			},
			startSignUpSession: func(context.Context, CognitoSessionInput, string, string) (AuthenticatedSession, error) {
				return AuthenticatedSession{Token: "raw-session-token"}, nil
			},
		},
		defaultFanAuthRepositoryStub(),
		defaultFanAuthViewerReaderStub(),
		signUpDraftStoreStub{
			getDraft: func(context.Context, string) (SignUpDraft, error) {
				return SignUpDraft{
					DisplayName: "Mina",
					Handle:      "mina",
					Password:    "VeryStrongPass123!",
				}, nil
			},
			deleteDraft: func(context.Context, string) error { return errors.New("boom") },
			saveDraft:   func(context.Context, string, SignUpDraft, time.Duration) error { return nil },
		},
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	got, err := service.SignIn(context.Background(), "fan@example.com", "VeryStrongPass123!")
	if err != nil {
		t.Fatalf("SignIn() error = %v, want nil", err)
	}
	if got.Token != "raw-session-token" {
		t.Fatalf("SignIn() token got %q want %q", got.Token, "raw-session-token")
	}
}

func TestFanAuthServiceSignInFallsBackWhenDraftStoreIsUnavailableForExistingIdentity(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	service := NewFanAuthService(
		fanAuthProviderStub{
			signIn: func(context.Context, string, string) (CognitoSessionInput, error) {
				return CognitoSessionInput{
					Subject:         "cognito-subject",
					Email:           "fan@example.com",
					EmailVerified:   true,
					AuthenticatedAt: time.Unix(1712000002, 0).UTC(),
				}, nil
			},
		},
		fanAuthSessionManagerStub{
			startSession: func(_ context.Context, input CognitoSessionInput) (AuthenticatedSession, error) {
				if input.Subject != "cognito-subject" {
					t.Fatalf("StartSession() subject got %q want %q", input.Subject, "cognito-subject")
				}
				return AuthenticatedSession{Token: "regular-session-token"}, nil
			},
			startSignUpSession: func(context.Context, CognitoSessionInput, string, string) (AuthenticatedSession, error) {
				t.Fatal("StartSignUpSession() should not be called when draft lookup fails")
				return AuthenticatedSession{}, nil
			},
		},
		fanAuthRepositoryStub{
			getIdentityByProviderAndSubject: func(context.Context, string, string) (Identity, error) { return Identity{}, ErrIdentityNotFound },
			getIdentityByNormalizedEmail: func(_ context.Context, emailNormalized string) (Identity, error) {
				if emailNormalized != "fan@example.com" {
					t.Fatalf("GetIdentityByNormalizedEmail() email got %q want %q", emailNormalized, "fan@example.com")
				}
				return Identity{ID: uuid.New(), UserID: userID}, nil
			},
			getPreferredEmailByUserID:    func(context.Context, uuid.UUID) (string, error) { return "", ErrIdentityNotFound },
			handleExists:                 func(context.Context, string) (bool, error) { return false, nil },
			revokeActiveSession:          func(context.Context, string, time.Time) (SessionRecord, error) { return SessionRecord{}, nil },
			updateActiveModeByTokenHash:  func(context.Context, string, ActiveMode) (SessionRecord, error) { return SessionRecord{}, nil },
		},
		defaultFanAuthViewerReaderStub(),
		signUpDraftStoreStub{
			getDraft:    func(context.Context, string) (SignUpDraft, error) { return SignUpDraft{}, errors.New("redis unavailable") },
			deleteDraft: func(context.Context, string) error { return nil },
			saveDraft:   func(context.Context, string, SignUpDraft, time.Duration) error { return nil },
		},
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	got, err := service.SignIn(context.Background(), "fan@example.com", "VeryStrongPass123!")
	if err != nil {
		t.Fatalf("SignIn() error = %v, want nil", err)
	}
	if got.Token != "regular-session-token" {
		t.Fatalf("SignIn() token got %q want %q", got.Token, "regular-session-token")
	}
}

func TestFanAuthServiceSignInFallsBackWhenDraftStoreIsUnavailableForExistingSubjectIdentity(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	service := NewFanAuthService(
		fanAuthProviderStub{
			signIn: func(context.Context, string, string) (CognitoSessionInput, error) {
				return CognitoSessionInput{
					Subject:         "cognito-subject",
					Email:           "fan@example.com",
					EmailVerified:   true,
					AuthenticatedAt: time.Unix(1712000003, 0).UTC(),
				}, nil
			},
		},
		fanAuthSessionManagerStub{
			startSession: func(_ context.Context, input CognitoSessionInput) (AuthenticatedSession, error) {
				if input.Subject != "cognito-subject" {
					t.Fatalf("StartSession() subject got %q want %q", input.Subject, "cognito-subject")
				}
				return AuthenticatedSession{Token: "subject-session-token"}, nil
			},
			startSignUpSession: func(context.Context, CognitoSessionInput, string, string) (AuthenticatedSession, error) {
				t.Fatal("StartSignUpSession() should not be called when draft lookup fails")
				return AuthenticatedSession{}, nil
			},
		},
		fanAuthRepositoryStub{
			getIdentityByProviderAndSubject: func(_ context.Context, provider string, providerSubject string) (Identity, error) {
				if provider != identityProviderCognito {
					t.Fatalf("GetIdentityByProviderAndSubject() provider got %q want %q", provider, identityProviderCognito)
				}
				if providerSubject != "cognito-subject" {
					t.Fatalf("GetIdentityByProviderAndSubject() subject got %q want %q", providerSubject, "cognito-subject")
				}
				return Identity{ID: uuid.New(), UserID: userID}, nil
			},
			getIdentityByNormalizedEmail: func(context.Context, string) (Identity, error) {
				t.Fatal("GetIdentityByNormalizedEmail() should not be called when subject identity exists")
				return Identity{}, nil
			},
			getPreferredEmailByUserID:   func(context.Context, uuid.UUID) (string, error) { return "", ErrIdentityNotFound },
			handleExists:                func(context.Context, string) (bool, error) { return false, nil },
			revokeActiveSession:         func(context.Context, string, time.Time) (SessionRecord, error) { return SessionRecord{}, nil },
			updateActiveModeByTokenHash: func(context.Context, string, ActiveMode) (SessionRecord, error) { return SessionRecord{}, nil },
		},
		defaultFanAuthViewerReaderStub(),
		signUpDraftStoreStub{
			getDraft:    func(context.Context, string) (SignUpDraft, error) { return SignUpDraft{}, errors.New("redis unavailable") },
			deleteDraft: func(context.Context, string) error { return nil },
			saveDraft:   func(context.Context, string, SignUpDraft, time.Duration) error { return nil },
		},
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	got, err := service.SignIn(context.Background(), "fan@example.com", "VeryStrongPass123!")
	if err != nil {
		t.Fatalf("SignIn() error = %v, want nil", err)
	}
	if got.Token != "subject-session-token" {
		t.Fatalf("SignIn() token got %q want %q", got.Token, "subject-session-token")
	}
}

func TestFanAuthServiceSignInMapsConfirmationRequired(t *testing.T) {
	t.Parallel()

	service := newTestFanAuthService(
		fanAuthProviderStub{
			signIn: func(context.Context, string, string) (CognitoSessionInput, error) {
				return CognitoSessionInput{}, &smithy.GenericAPIError{Code: "UserNotConfirmedException", Message: "not confirmed"}
			},
		},
	)

	if _, err := service.SignIn(context.Background(), "fan@example.com", "VeryStrongPass123!"); !errors.Is(err, ErrConfirmationRequired) {
		t.Fatalf("SignIn() error got %v want %v", err, ErrConfirmationRequired)
	}
}

func TestFanAuthServiceSignInStartsRegularSessionWithoutDraft(t *testing.T) {
	t.Parallel()

	service := NewFanAuthService(
		fanAuthProviderStub{
			signIn: func(context.Context, string, string) (CognitoSessionInput, error) {
				return CognitoSessionInput{
					Subject:         "cognito-subject",
					Email:           "fan@example.com",
					EmailVerified:   true,
					AuthenticatedAt: time.Unix(1712000600, 0).UTC(),
				}, nil
			},
		},
		fanAuthSessionManagerStub{
			startSession: func(_ context.Context, input CognitoSessionInput) (AuthenticatedSession, error) {
				if input.Subject != "cognito-subject" {
					t.Fatalf("StartSession() subject got %q want %q", input.Subject, "cognito-subject")
				}
				return AuthenticatedSession{Token: "raw-session-token"}, nil
			},
			startSignUpSession: func(context.Context, CognitoSessionInput, string, string) (AuthenticatedSession, error) {
				t.Fatal("StartSignUpSession() should not be called when draft is absent")
				return AuthenticatedSession{}, nil
			},
		},
		defaultFanAuthRepositoryStub(),
		defaultFanAuthViewerReaderStub(),
		defaultSignUpDraftStoreStub(),
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	got, err := service.SignIn(context.Background(), "fan@example.com", "VeryStrongPass123!")
	if err != nil {
		t.Fatalf("SignIn() error = %v, want nil", err)
	}
	if got.Token != "raw-session-token" {
		t.Fatalf("SignIn() token got %q want %q", got.Token, "raw-session-token")
	}
}

func TestFanAuthServiceSignInMapsRateLimited(t *testing.T) {
	t.Parallel()

	service := newTestFanAuthService(
		fanAuthProviderStub{
			signIn: func(context.Context, string, string) (CognitoSessionInput, error) {
				return CognitoSessionInput{}, genericAPIError("TooManyRequestsException")
			},
		},
	)

	if _, err := service.SignIn(context.Background(), "fan@example.com", "VeryStrongPass123!"); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("SignIn() error got %v want %v", err, ErrRateLimited)
	}
}

func TestFanAuthServiceStartSignUpReturnsAcceptedAndStoresDraft(t *testing.T) {
	t.Parallel()

	var savedDraft SignUpDraft
	service := NewFanAuthService(
		fanAuthProviderStub{
			signIn: func(context.Context, string, string) (CognitoSessionInput, error) { return CognitoSessionInput{}, nil },
			signUp: func(_ context.Context, email string, password string) error {
				if email != "fan@example.com" {
					t.Fatalf("SignUp() email got %q want %q", email, "fan@example.com")
				}
				if password != "VeryStrongPass123!" {
					t.Fatalf("SignUp() password got %q want %q", password, "VeryStrongPass123!")
				}
				return nil
			},
		},
		defaultFanAuthSessionManagerStub(),
		fanAuthRepositoryStub{
			getIdentityByNormalizedEmail: func(context.Context, string) (Identity, error) { return Identity{}, ErrIdentityNotFound },
			getPreferredEmailByUserID:    func(context.Context, uuid.UUID) (string, error) { return "", nil },
			handleExists:                 func(_ context.Context, handle string) (bool, error) { return false, nil },
			revokeActiveSession:          func(context.Context, string, time.Time) (SessionRecord, error) { return SessionRecord{}, nil },
		},
		defaultFanAuthViewerReaderStub(),
		signUpDraftStoreStub{
			getDraft:    func(context.Context, string) (SignUpDraft, error) { return SignUpDraft{}, ErrSignUpDraftNotFound },
			deleteDraft: func(context.Context, string) error { return nil },
			saveDraft: func(_ context.Context, email string, draft SignUpDraft, ttl time.Duration) error {
				if email != "fan@example.com" {
					t.Fatalf("SaveDraft() email got %q want %q", email, "fan@example.com")
				}
				if ttl != signUpDraftTTL {
					t.Fatalf("SaveDraft() ttl got %s want %s", ttl, signUpDraftTTL)
				}
				savedDraft = draft
				return nil
			},
		},
		authCooldownStoreStub{
			release: func(context.Context, string) error { return nil },
			tryActivate: func(_ context.Context, key string, ttl time.Duration) (bool, error) {
				if key != "sign_up:fan@example.com" {
					t.Fatalf("TryActivate() key got %q want %q", key, "sign_up:fan@example.com")
				}
				if ttl != fanAuthStartCooldownTTL {
					t.Fatalf("TryActivate() ttl got %s want %s", ttl, fanAuthStartCooldownTTL)
				}
				return true, nil
			},
		},
	)

	got, err := service.StartSignUp(context.Background(), "fan@example.com", "Mina", "@mina", "VeryStrongPass123!")
	if err != nil {
		t.Fatalf("StartSignUp() error = %v, want nil", err)
	}
	if got.NextStep != FanAuthNextStepConfirmSignUp {
		t.Fatalf("StartSignUp() next step got %q want %q", got.NextStep, FanAuthNextStepConfirmSignUp)
	}
	if savedDraft.Handle != "mina" {
		t.Fatalf("StartSignUp() saved handle got %q want %q", savedDraft.Handle, "mina")
	}
	if savedDraft.Password != "VeryStrongPass123!" {
		t.Fatalf("StartSignUp() saved password got %q want %q", savedDraft.Password, "VeryStrongPass123!")
	}
}

func TestFanAuthServiceStartSignUpKeepsExistingDraftWhileCooldownIsActive(t *testing.T) {
	t.Parallel()

	var savedDraft SignUpDraft
	saveDraftCalls := 0
	tryActivateCalled := false
	service := NewFanAuthService(
		fanAuthProviderStub{
			signIn: func(context.Context, string, string) (CognitoSessionInput, error) { return CognitoSessionInput{}, nil },
			signUp: func(context.Context, string, string) error {
				t.Fatal("SignUp() should not be called while cooldown is active")
				return nil
			},
		},
		defaultFanAuthSessionManagerStub(),
		defaultFanAuthRepositoryStub(),
		defaultFanAuthViewerReaderStub(),
		signUpDraftStoreStub{
			getDraft: func(context.Context, string) (SignUpDraft, error) {
				return SignUpDraft{
					DisplayName: "Existing Mina",
					Email:       "fan@example.com",
					ExpiresAt:   time.Unix(1712000900, 0).UTC(),
					Handle:      "existing-mina",
					Password:    "ExistingStrongPass123!",
				}, nil
			},
			deleteDraft: func(context.Context, string) error { return nil },
			saveDraft: func(_ context.Context, _ string, draft SignUpDraft, _ time.Duration) error {
				saveDraftCalls++
				savedDraft = draft
				return nil
			},
		},
		authCooldownStoreStub{
			release: func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) {
				tryActivateCalled = true
				return false, nil
			},
		},
	)

	got, err := service.StartSignUp(context.Background(), "fan@example.com", "Mina", "@mina", "VeryStrongPass123!")
	if err != nil {
		t.Fatalf("StartSignUp() error = %v, want nil", err)
	}
	if got.NextStep != FanAuthNextStepConfirmSignUp {
		t.Fatalf("StartSignUp() next step got %q want %q", got.NextStep, FanAuthNextStepConfirmSignUp)
	}
	if !tryActivateCalled {
		t.Fatal("StartSignUp() did not evaluate cooldown for existing draft")
	}
	if saveDraftCalls != 1 {
		t.Fatalf("StartSignUp() save draft calls got %d want %d", saveDraftCalls, 1)
	}
	if savedDraft.Handle != "existing-mina" {
		t.Fatalf("StartSignUp() saved handle got %q want %q", savedDraft.Handle, "existing-mina")
	}
	if savedDraft.Password != "ExistingStrongPass123!" {
		t.Fatalf("StartSignUp() saved password got %q want %q", savedDraft.Password, "ExistingStrongPass123!")
	}
}

func TestFanAuthServiceStartSignUpResendsWhenDraftExistsAndCooldownElapsed(t *testing.T) {
	t.Parallel()

	var savedDraft SignUpDraft
	saveDraftCalls := 0
	service := NewFanAuthService(
		fanAuthProviderStub{
			signUp: func(_ context.Context, email string, password string) error {
				if email != "fan@example.com" {
					t.Fatalf("SignUp() email got %q want %q", email, "fan@example.com")
				}
				if password != "ExistingStrongPass123!" {
					t.Fatalf("SignUp() password got %q want %q", password, "ExistingStrongPass123!")
				}
				return genericAPIError("UsernameExistsException")
			},
			resendSignUpCode: func(_ context.Context, email string) error {
				if email != "fan@example.com" {
					t.Fatalf("ResendSignUpCode() email got %q want %q", email, "fan@example.com")
				}
				return nil
			},
		},
		defaultFanAuthSessionManagerStub(),
		defaultFanAuthRepositoryStub(),
		defaultFanAuthViewerReaderStub(),
		signUpDraftStoreStub{
			getDraft: func(context.Context, string) (SignUpDraft, error) {
				return SignUpDraft{
					DisplayName: "Existing Mina",
					Email:       "fan@example.com",
					ExpiresAt:   time.Unix(1712000900, 0).UTC(),
					Handle:      "existing-mina",
					Password:    "ExistingStrongPass123!",
				}, nil
			},
			deleteDraft: func(context.Context, string) error { return nil },
			saveDraft: func(_ context.Context, _ string, draft SignUpDraft, ttl time.Duration) error {
				if ttl != signUpDraftTTL {
					t.Fatalf("SaveDraft() ttl got %s want %s", ttl, signUpDraftTTL)
				}
				saveDraftCalls++
				savedDraft = draft
				return nil
			},
		},
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	got, err := service.StartSignUp(context.Background(), "fan@example.com", "Mina", "@mina", "VeryStrongPass123!")
	if err != nil {
		t.Fatalf("StartSignUp() error = %v, want nil", err)
	}
	if got.NextStep != FanAuthNextStepConfirmSignUp {
		t.Fatalf("StartSignUp() next step got %q want %q", got.NextStep, FanAuthNextStepConfirmSignUp)
	}
	if saveDraftCalls != 1 {
		t.Fatalf("StartSignUp() save draft calls got %d want %d", saveDraftCalls, 1)
	}
	if savedDraft.Handle != "existing-mina" {
		t.Fatalf("StartSignUp() saved handle got %q want %q", savedDraft.Handle, "existing-mina")
	}
	if savedDraft.Password != "ExistingStrongPass123!" {
		t.Fatalf("StartSignUp() saved password got %q want %q", savedDraft.Password, "ExistingStrongPass123!")
	}
}

func TestFanAuthServiceStartSignUpResendsConfirmationCodeForExistingCognitoUser(t *testing.T) {
	t.Parallel()

	var savedDraft SignUpDraft
	service := NewFanAuthService(
		fanAuthProviderStub{
			signIn: func(context.Context, string, string) (CognitoSessionInput, error) { return CognitoSessionInput{}, nil },
			signUp: func(context.Context, string, string) error {
				return genericAPIError("UsernameExistsException")
			},
			resendSignUpCode: func(_ context.Context, email string) error {
				if email != "fan@example.com" {
					t.Fatalf("ResendSignUpCode() email got %q want %q", email, "fan@example.com")
				}
				return nil
			},
		},
		defaultFanAuthSessionManagerStub(),
		defaultFanAuthRepositoryStub(),
		defaultFanAuthViewerReaderStub(),
		signUpDraftStoreStub{
			getDraft:    func(context.Context, string) (SignUpDraft, error) { return SignUpDraft{}, ErrSignUpDraftNotFound },
			deleteDraft: func(context.Context, string) error { return nil },
			saveDraft: func(_ context.Context, _ string, draft SignUpDraft, _ time.Duration) error {
				savedDraft = draft
				return nil
			},
		},
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	got, err := service.StartSignUp(context.Background(), "fan@example.com", "Mina", "@mina", "VeryStrongPass123!")
	if err != nil {
		t.Fatalf("StartSignUp() error = %v, want nil", err)
	}
	if got.NextStep != FanAuthNextStepConfirmSignUp {
		t.Fatalf("StartSignUp() next step got %q want %q", got.NextStep, FanAuthNextStepConfirmSignUp)
	}
	if savedDraft.Handle != "mina" {
		t.Fatalf("StartSignUp() saved handle got %q want %q", savedDraft.Handle, "mina")
	}
}

func TestFanAuthServiceStartSignUpStoresDraftWhenResendIsSuppressed(t *testing.T) {
	t.Parallel()

	savedDraft := false
	service := NewFanAuthService(
		fanAuthProviderStub{
			signIn: func(context.Context, string, string) (CognitoSessionInput, error) { return CognitoSessionInput{}, nil },
			signUp: func(context.Context, string, string) error {
				return genericAPIError("UsernameExistsException")
			},
			resendSignUpCode: func(context.Context, string) error {
				return genericAPIError("InvalidParameterException")
			},
		},
		defaultFanAuthSessionManagerStub(),
		defaultFanAuthRepositoryStub(),
		defaultFanAuthViewerReaderStub(),
		signUpDraftStoreStub{
			getDraft:    func(context.Context, string) (SignUpDraft, error) { return SignUpDraft{}, ErrSignUpDraftNotFound },
			deleteDraft: func(context.Context, string) error { return nil },
			saveDraft: func(context.Context, string, SignUpDraft, time.Duration) error {
				savedDraft = true
				return nil
			},
		},
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	got, err := service.StartSignUp(context.Background(), "fan@example.com", "Mina", "@mina", "VeryStrongPass123!")
	if err != nil {
		t.Fatalf("StartSignUp() error = %v, want nil", err)
	}
	if got.NextStep != FanAuthNextStepConfirmSignUp {
		t.Fatalf("StartSignUp() next step got %q want %q", got.NextStep, FanAuthNextStepConfirmSignUp)
	}
	if !savedDraft {
		t.Fatal("StartSignUp() did not persist draft when resend was suppressed")
	}
}

func TestFanAuthServiceStartSignUpMapsProviderError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr error
	}{
		{
			name:    "password policy violation",
			err:     genericAPIError("InvalidPasswordException"),
			wantErr: ErrPasswordPolicyViolation,
		},
		{
			name:    "rate limited",
			err:     genericAPIError("TooManyRequestsException"),
			wantErr: ErrRateLimited,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := newTestFanAuthService(fanAuthProviderStub{
				signUp: func(context.Context, string, string) error {
					return tt.err
				},
			})

			if _, err := service.StartSignUp(context.Background(), "fan@example.com", "Mina", "@mina", "VeryStrongPass123!"); !errors.Is(err, tt.wantErr) {
				t.Fatalf("StartSignUp() error got %v want %v", err, tt.wantErr)
			}
		})
	}
}

func TestFanAuthServiceStartSignUpDeletesNewDraftWhenRemoteSignUpFails(t *testing.T) {
	t.Parallel()

	deleted := false
	service := NewFanAuthService(
		fanAuthProviderStub{
			signUp: func(context.Context, string, string) error {
				return genericAPIError("TooManyRequestsException")
			},
		},
		defaultFanAuthSessionManagerStub(),
		defaultFanAuthRepositoryStub(),
		defaultFanAuthViewerReaderStub(),
		signUpDraftStoreStub{
			getDraft: func(context.Context, string) (SignUpDraft, error) { return SignUpDraft{}, ErrSignUpDraftNotFound },
			deleteDraft: func(context.Context, string) error {
				deleted = true
				return nil
			},
			saveDraft: func(context.Context, string, SignUpDraft, time.Duration) error { return nil },
		},
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	if _, err := service.StartSignUp(context.Background(), "fan@example.com", "Mina", "@mina", "VeryStrongPass123!"); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("StartSignUp() error got %v want %v", err, ErrRateLimited)
	}
	if !deleted {
		t.Fatal("StartSignUp() did not delete newly created draft after remote failure")
	}
}

func TestFanAuthServiceStartSignUpWhileCooldownIsActiveStoresDraft(t *testing.T) {
	t.Parallel()

	saved := false
	service := NewFanAuthService(
		fanAuthProviderStub{
			signUp: func(context.Context, string, string) error {
				t.Fatal("SignUp() should not be called while cooldown is active")
				return nil
			},
		},
		defaultFanAuthSessionManagerStub(),
		defaultFanAuthRepositoryStub(),
		defaultFanAuthViewerReaderStub(),
		signUpDraftStoreStub{
			getDraft:    func(context.Context, string) (SignUpDraft, error) { return SignUpDraft{}, ErrSignUpDraftNotFound },
			deleteDraft: func(context.Context, string) error { return nil },
			saveDraft: func(context.Context, string, SignUpDraft, time.Duration) error {
				saved = true
				return nil
			},
		},
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return false, nil },
		},
	)

	got, err := service.StartSignUp(context.Background(), "fan@example.com", "Mina", "@mina", "VeryStrongPass123!")
	if err != nil {
		t.Fatalf("StartSignUp() error = %v, want nil", err)
	}
	if got.NextStep != FanAuthNextStepConfirmSignUp {
		t.Fatalf("StartSignUp() next step got %q want %q", got.NextStep, FanAuthNextStepConfirmSignUp)
	}
	if !saved {
		t.Fatal("StartSignUp() did not persist draft while cooldown was active")
	}
}

func TestFanAuthServiceStartSignUpRejectsTakenHandle(t *testing.T) {
	t.Parallel()

	service := NewFanAuthService(
		fanAuthProviderStub{
			signUp: func(context.Context, string, string) error {
				t.Fatal("SignUp() should not be called when handle is already taken")
				return nil
			},
		},
		defaultFanAuthSessionManagerStub(),
		fanAuthRepositoryStub{
			getIdentityByNormalizedEmail: func(context.Context, string) (Identity, error) { return Identity{}, ErrIdentityNotFound },
			getPreferredEmailByUserID:    func(context.Context, uuid.UUID) (string, error) { return "", ErrIdentityNotFound },
			handleExists:                 func(context.Context, string) (bool, error) { return true, nil },
			revokeActiveSession:          func(context.Context, string, time.Time) (SessionRecord, error) { return SessionRecord{}, nil },
		},
		defaultFanAuthViewerReaderStub(),
		defaultSignUpDraftStoreStub(),
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	if _, err := service.StartSignUp(context.Background(), "fan@example.com", "Mina", "@mina", "VeryStrongPass123!"); !errors.Is(err, ErrHandleAlreadyTaken) {
		t.Fatalf("StartSignUp() error got %v want %v", err, ErrHandleAlreadyTaken)
	}
}

func TestFanAuthServiceConfirmSignUpUsesDraftAndDeletesIt(t *testing.T) {
	t.Parallel()

	service := NewFanAuthService(
		fanAuthProviderStub{
			confirmSignUp: func(_ context.Context, email string, confirmationCode string) error {
				if email != "fan@example.com" {
					t.Fatalf("ConfirmSignUp() email got %q want %q", email, "fan@example.com")
				}
				if confirmationCode != "123456" {
					t.Fatalf("ConfirmSignUp() confirmationCode got %q want %q", confirmationCode, "123456")
				}
				return nil
			},
			signIn: func(_ context.Context, email string, password string) (CognitoSessionInput, error) {
				if password != "VeryStrongPass123!" {
					t.Fatalf("SignIn() password got %q want %q", password, "VeryStrongPass123!")
				}
				return CognitoSessionInput{
					Subject:         "cognito-subject",
					Email:           email,
					EmailVerified:   true,
					AuthenticatedAt: time.Unix(1712000100, 0).UTC(),
				}, nil
			},
		},
		fanAuthSessionManagerStub{
			startSession: func(context.Context, CognitoSessionInput) (AuthenticatedSession, error) {
				t.Fatal("StartSession() should not be called for sign up confirm")
				return AuthenticatedSession{}, nil
			},
			startSignUpSession: func(_ context.Context, input CognitoSessionInput, displayName string, handle string) (AuthenticatedSession, error) {
				if displayName != "Mina" {
					t.Fatalf("StartSignUpSession() displayName got %q want %q", displayName, "Mina")
				}
				if handle != "mina" {
					t.Fatalf("StartSignUpSession() handle got %q want %q", handle, "mina")
				}
				if input.Subject != "cognito-subject" {
					t.Fatalf("StartSignUpSession() subject got %q want %q", input.Subject, "cognito-subject")
				}
				return AuthenticatedSession{Token: "raw-session-token"}, nil
			},
		},
		fanAuthRepositoryStub{
			getIdentityByNormalizedEmail: func(context.Context, string) (Identity, error) { return Identity{}, ErrIdentityNotFound },
			getPreferredEmailByUserID:    func(context.Context, uuid.UUID) (string, error) { return "", nil },
			handleExists:                 func(context.Context, string) (bool, error) { return false, nil },
			revokeActiveSession:          func(context.Context, string, time.Time) (SessionRecord, error) { return SessionRecord{}, nil },
		},
		defaultFanAuthViewerReaderStub(),
		signUpDraftStoreStub{
			getDraft: func(context.Context, string) (SignUpDraft, error) {
				return SignUpDraft{
					DisplayName: "Mina",
					Handle:      "mina",
					Password:    "VeryStrongPass123!",
				}, nil
			},
			deleteDraft: func(_ context.Context, email string) error {
				if email != "fan@example.com" {
					t.Fatalf("DeleteDraft() email got %q want %q", email, "fan@example.com")
				}
				return nil
			},
			saveDraft: func(context.Context, string, SignUpDraft, time.Duration) error { return nil },
		},
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	got, err := service.ConfirmSignUp(context.Background(), "fan@example.com", "123456")
	if err != nil {
		t.Fatalf("ConfirmSignUp() error = %v, want nil", err)
	}
	if got.Token != "raw-session-token" {
		t.Fatalf("ConfirmSignUp() token got %q want %q", got.Token, "raw-session-token")
	}
}

func TestFanAuthServiceConfirmSignUpWithoutDraftMapsExpired(t *testing.T) {
	t.Parallel()

	service := newTestFanAuthService(
		fanAuthProviderStub{
			confirmSignUp: func(context.Context, string, string) error { return nil },
			signIn:        func(context.Context, string, string) (CognitoSessionInput, error) { return CognitoSessionInput{}, nil },
		},
	)
	service.draftStore = signUpDraftStoreStub{
		getDraft:    func(context.Context, string) (SignUpDraft, error) { return SignUpDraft{}, ErrSignUpDraftNotFound },
		deleteDraft: func(context.Context, string) error { return nil },
		saveDraft:   func(context.Context, string, SignUpDraft, time.Duration) error { return nil },
	}

	if _, err := service.ConfirmSignUp(context.Background(), "fan@example.com", "123456"); !errors.Is(err, ErrConfirmationCodeExpired) {
		t.Fatalf("ConfirmSignUp() error got %v want %v", err, ErrConfirmationCodeExpired)
	}
}

func TestFanAuthServiceConfirmSignUpMapsProviderError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr error
	}{
		{
			name:    "invalid code",
			err:     genericAPIError("CodeMismatchException"),
			wantErr: ErrInvalidConfirmationCode,
		},
		{
			name:    "expired code",
			err:     genericAPIError("ExpiredCodeException"),
			wantErr: ErrConfirmationCodeExpired,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := newTestFanAuthService(fanAuthProviderStub{
				confirmSignUp: func(context.Context, string, string) error {
					return tt.err
				},
				signIn: func(context.Context, string, string) (CognitoSessionInput, error) {
					t.Fatal("SignIn() should not be called when confirmation fails")
					return CognitoSessionInput{}, nil
				},
			})
			service.draftStore = signUpDraftStoreStub{
				getDraft: func(context.Context, string) (SignUpDraft, error) {
					return SignUpDraft{
						DisplayName: "Mina",
						Handle:      "mina",
						Password:    "VeryStrongPass123!",
					}, nil
				},
				deleteDraft: func(context.Context, string) error { return nil },
				saveDraft:   func(context.Context, string, SignUpDraft, time.Duration) error { return nil },
			}

			if _, err := service.ConfirmSignUp(context.Background(), "fan@example.com", "123456"); !errors.Is(err, tt.wantErr) {
				t.Fatalf("ConfirmSignUp() error got %v want %v", err, tt.wantErr)
			}
		})
	}
}

func TestFanAuthServiceConfirmSignUpMapsRateLimitedAutoSignIn(t *testing.T) {
	t.Parallel()

	service := newTestFanAuthService(
		fanAuthProviderStub{
			confirmSignUp: func(context.Context, string, string) error { return nil },
			signIn: func(context.Context, string, string) (CognitoSessionInput, error) {
				return CognitoSessionInput{}, genericAPIError("TooManyRequestsException")
			},
		},
	)
	service.draftStore = signUpDraftStoreStub{
		getDraft: func(context.Context, string) (SignUpDraft, error) {
			return SignUpDraft{
				DisplayName: "Mina",
				Handle:      "mina",
				Password:    "VeryStrongPass123!",
			}, nil
		},
		deleteDraft: func(context.Context, string) error { return nil },
		saveDraft:   func(context.Context, string, SignUpDraft, time.Duration) error { return nil },
	}

	if _, err := service.ConfirmSignUp(context.Background(), "fan@example.com", "123456"); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("ConfirmSignUp() error got %v want %v", err, ErrRateLimited)
	}
}

func TestFanAuthServiceConfirmSignUpIgnoresDeleteDraftError(t *testing.T) {
	t.Parallel()

	service := NewFanAuthService(
		fanAuthProviderStub{
			confirmSignUp: func(context.Context, string, string) error { return nil },
			signIn: func(context.Context, string, string) (CognitoSessionInput, error) {
				return CognitoSessionInput{
					Subject:         "cognito-subject",
					Email:           "fan@example.com",
					EmailVerified:   true,
					AuthenticatedAt: time.Unix(1712000700, 0).UTC(),
				}, nil
			},
		},
		fanAuthSessionManagerStub{
			startSession: func(context.Context, CognitoSessionInput) (AuthenticatedSession, error) {
				t.Fatal("StartSession() should not be called during sign-up confirm")
				return AuthenticatedSession{}, nil
			},
			startSignUpSession: func(context.Context, CognitoSessionInput, string, string) (AuthenticatedSession, error) {
				return AuthenticatedSession{Token: "raw-session-token"}, nil
			},
		},
		defaultFanAuthRepositoryStub(),
		defaultFanAuthViewerReaderStub(),
		signUpDraftStoreStub{
			getDraft: func(context.Context, string) (SignUpDraft, error) {
				return SignUpDraft{
					DisplayName: "Mina",
					Handle:      "mina",
					Password:    "VeryStrongPass123!",
				}, nil
			},
			deleteDraft: func(context.Context, string) error { return errors.New("boom") },
			saveDraft:   func(context.Context, string, SignUpDraft, time.Duration) error { return nil },
		},
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	got, err := service.ConfirmSignUp(context.Background(), "fan@example.com", "123456")
	if err != nil {
		t.Fatalf("ConfirmSignUp() error = %v, want nil", err)
	}
	if got.Token != "raw-session-token" {
		t.Fatalf("ConfirmSignUp() token got %q want %q", got.Token, "raw-session-token")
	}
}

func TestFanAuthServiceStartPasswordResetAcceptsSuppressedProviderError(t *testing.T) {
	t.Parallel()

	service := NewFanAuthService(
		fanAuthProviderStub{
			startPasswordReset: func(context.Context, string) error {
				return &smithy.GenericAPIError{Code: "InvalidParameterException", Message: "suppressed"}
			},
		},
		defaultFanAuthSessionManagerStub(),
		defaultFanAuthRepositoryStub(),
		defaultFanAuthViewerReaderStub(),
		defaultSignUpDraftStoreStub(),
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	got, err := service.StartPasswordReset(context.Background(), "fan@example.com")
	if err != nil {
		t.Fatalf("StartPasswordReset() error = %v, want nil", err)
	}
	if got.NextStep != FanAuthNextStepConfirmPasswordReset {
		t.Fatalf("StartPasswordReset() next step got %q want %q", got.NextStep, FanAuthNextStepConfirmPasswordReset)
	}
}

func TestFanAuthServiceStartPasswordResetMapsRateLimited(t *testing.T) {
	t.Parallel()

	released := false
	service := NewFanAuthService(
		fanAuthProviderStub{
			startPasswordReset: func(context.Context, string) error {
				return genericAPIError("TooManyRequestsException")
			},
		},
		defaultFanAuthSessionManagerStub(),
		defaultFanAuthRepositoryStub(),
		defaultFanAuthViewerReaderStub(),
		defaultSignUpDraftStoreStub(),
		authCooldownStoreStub{
			release: func(context.Context, string) error {
				released = true
				return nil
			},
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	if _, err := service.StartPasswordReset(context.Background(), "fan@example.com"); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("StartPasswordReset() error got %v want %v", err, ErrRateLimited)
	}
	if !released {
		t.Fatal("StartPasswordReset() did not release cooldown after provider error")
	}
}

func TestFanAuthServiceStartPasswordResetSkipsProviderWhileCooldownActive(t *testing.T) {
	t.Parallel()

	service := NewFanAuthService(
		fanAuthProviderStub{
			startPasswordReset: func(context.Context, string) error {
				t.Fatal("StartPasswordReset() provider should not be called while cooldown is active")
				return nil
			},
		},
		defaultFanAuthSessionManagerStub(),
		defaultFanAuthRepositoryStub(),
		defaultFanAuthViewerReaderStub(),
		defaultSignUpDraftStoreStub(),
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return false, nil },
		},
	)

	got, err := service.StartPasswordReset(context.Background(), "fan@example.com")
	if err != nil {
		t.Fatalf("StartPasswordReset() error = %v, want nil", err)
	}
	if got.NextStep != FanAuthNextStepConfirmPasswordReset {
		t.Fatalf("StartPasswordReset() next step got %q want %q", got.NextStep, FanAuthNextStepConfirmPasswordReset)
	}
}

func TestFanAuthServiceConfirmPasswordResetMapsProviderError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr error
	}{
		{
			name:    "invalid code",
			err:     genericAPIError("CodeMismatchException"),
			wantErr: ErrInvalidConfirmationCode,
		},
		{
			name:    "expired code",
			err:     genericAPIError("ExpiredCodeException"),
			wantErr: ErrConfirmationCodeExpired,
		},
		{
			name:    "password policy violation",
			err:     genericAPIError("InvalidPasswordException"),
			wantErr: ErrPasswordPolicyViolation,
		},
		{
			name:    "rate limited",
			err:     genericAPIError("TooManyRequestsException"),
			wantErr: ErrRateLimited,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := newTestFanAuthService(fanAuthProviderStub{
				confirmPasswordReset: func(context.Context, string, string, string) error {
					return tt.err
				},
			})

			err := service.ConfirmPasswordReset(context.Background(), "fan@example.com", "123456", "AnotherStrongPass456!")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ConfirmPasswordReset() error got %v want %v", err, tt.wantErr)
			}
		})
	}
}

func TestFanAuthServiceReAuthenticateStartsRotatedSession(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	updateActiveModeCalled := false
	service := NewFanAuthService(
		fanAuthProviderStub{
			signIn: func(_ context.Context, email string, password string) (CognitoSessionInput, error) {
				if email != "fan@example.com" {
					t.Fatalf("SignIn() email got %q want %q", email, "fan@example.com")
				}
				if password != "VeryStrongPass123!" {
					t.Fatalf("SignIn() password got %q want %q", password, "VeryStrongPass123!")
				}
				return CognitoSessionInput{
					Subject:         "cognito-subject",
					Email:           email,
					EmailVerified:   true,
					AuthenticatedAt: time.Unix(1712000200, 0).UTC(),
				}, nil
			},
		},
		fanAuthSessionManagerStub{
			startSession: func(_ context.Context, input CognitoSessionInput) (AuthenticatedSession, error) {
				if input.Subject != "cognito-subject" {
					t.Fatalf("StartSession() subject got %q want %q", input.Subject, "cognito-subject")
				}
				return AuthenticatedSession{Token: "rotated-session-token"}, nil
			},
			startSignUpSession: func(context.Context, CognitoSessionInput, string, string) (AuthenticatedSession, error) {
				t.Fatal("StartSignUpSession() should not be called during re-auth")
				return AuthenticatedSession{}, nil
			},
		},
		fanAuthRepositoryStub{
			getIdentityByNormalizedEmail: func(context.Context, string) (Identity, error) { return Identity{}, ErrIdentityNotFound },
			getPreferredEmailByUserID: func(_ context.Context, gotUserID uuid.UUID) (string, error) {
				if gotUserID != userID {
					t.Fatalf("GetPreferredEmailByUserID() userID got %s want %s", gotUserID, userID)
				}
				return "fan@example.com", nil
			},
			handleExists:        func(context.Context, string) (bool, error) { return false, nil },
			revokeActiveSession: func(context.Context, string, time.Time) (SessionRecord, error) { return SessionRecord{}, nil },
			updateActiveModeByTokenHash: func(context.Context, string, ActiveMode) (SessionRecord, error) {
				updateActiveModeCalled = true
				return SessionRecord{}, nil
			},
		},
		fanAuthViewerReaderStub{
			readCurrentViewer: func(_ context.Context, rawSessionToken string) (Bootstrap, error) {
				if rawSessionToken != "raw-session-token" {
					t.Fatalf("ReadCurrentViewer() raw session token got %q want %q", rawSessionToken, "raw-session-token")
				}
				return Bootstrap{
					CurrentViewer: &CurrentViewer{ID: userID, ActiveMode: ActiveModeFan},
				}, nil
			},
		},
		defaultSignUpDraftStoreStub(),
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	got, err := service.ReAuthenticate(context.Background(), "raw-session-token", "VeryStrongPass123!")
	if err != nil {
		t.Fatalf("ReAuthenticate() error = %v, want nil", err)
	}
	if got.Token != "rotated-session-token" {
		t.Fatalf("ReAuthenticate() token got %q want %q", got.Token, "rotated-session-token")
	}
	if updateActiveModeCalled {
		t.Fatal("ReAuthenticate() updated active mode for fan mode session")
	}
}

func TestFanAuthServiceReAuthenticatePreservesCreatorMode(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	updateCalls := 0
	service := NewFanAuthService(
		fanAuthProviderStub{
			signIn: func(context.Context, string, string) (CognitoSessionInput, error) {
				return CognitoSessionInput{
					Subject:         "cognito-subject",
					Email:           "fan@example.com",
					EmailVerified:   true,
					AuthenticatedAt: time.Unix(1712000210, 0).UTC(),
				}, nil
			},
		},
		fanAuthSessionManagerStub{
			startSession: func(context.Context, CognitoSessionInput) (AuthenticatedSession, error) {
				return AuthenticatedSession{Token: "rotated-session-token"}, nil
			},
			startSignUpSession: func(context.Context, CognitoSessionInput, string, string) (AuthenticatedSession, error) {
				t.Fatal("StartSignUpSession() should not be called during re-auth")
				return AuthenticatedSession{}, nil
			},
		},
		fanAuthRepositoryStub{
			getIdentityByNormalizedEmail: func(context.Context, string) (Identity, error) { return Identity{}, ErrIdentityNotFound },
			getPreferredEmailByUserID: func(_ context.Context, gotUserID uuid.UUID) (string, error) {
				if gotUserID != userID {
					t.Fatalf("GetPreferredEmailByUserID() userID got %s want %s", gotUserID, userID)
				}
				return "fan@example.com", nil
			},
			handleExists:        func(context.Context, string) (bool, error) { return false, nil },
			revokeActiveSession: func(context.Context, string, time.Time) (SessionRecord, error) { return SessionRecord{}, nil },
			updateActiveModeByTokenHash: func(_ context.Context, sessionTokenHash string, activeMode ActiveMode) (SessionRecord, error) {
				updateCalls++
				if sessionTokenHash != HashSessionToken("rotated-session-token") {
					t.Fatalf("UpdateActiveModeByTokenHash() token hash got %q want %q", sessionTokenHash, HashSessionToken("rotated-session-token"))
				}
				if activeMode != ActiveModeCreator {
					t.Fatalf("UpdateActiveModeByTokenHash() active mode got %q want %q", activeMode, ActiveModeCreator)
				}
				return SessionRecord{}, nil
			},
		},
		fanAuthViewerReaderStub{
			readCurrentViewer: func(context.Context, string) (Bootstrap, error) {
				return Bootstrap{
					CurrentViewer: &CurrentViewer{ID: userID, ActiveMode: ActiveModeCreator},
				}, nil
			},
		},
		defaultSignUpDraftStoreStub(),
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	got, err := service.ReAuthenticate(context.Background(), "raw-session-token", "VeryStrongPass123!")
	if err != nil {
		t.Fatalf("ReAuthenticate() error = %v, want nil", err)
	}
	if got.Token != "rotated-session-token" {
		t.Fatalf("ReAuthenticate() token got %q want %q", got.Token, "rotated-session-token")
	}
	if updateCalls != 1 {
		t.Fatalf("ReAuthenticate() active mode update calls got %d want %d", updateCalls, 1)
	}
}

func TestFanAuthServiceReAuthenticateRequiresSession(t *testing.T) {
	t.Parallel()

	service := newTestFanAuthService(fanAuthProviderStub{})
	if _, err := service.ReAuthenticate(context.Background(), "", "VeryStrongPass123!"); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("ReAuthenticate() error got %v want %v", err, ErrAuthenticationRequired)
	}
}

func TestFanAuthServiceReAuthenticateRequiresKnownEmail(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	service := NewFanAuthService(
		fanAuthProviderStub{},
		defaultFanAuthSessionManagerStub(),
		fanAuthRepositoryStub{
			getIdentityByNormalizedEmail: func(context.Context, string) (Identity, error) { return Identity{}, ErrIdentityNotFound },
			getPreferredEmailByUserID:    func(context.Context, uuid.UUID) (string, error) { return "", ErrIdentityNotFound },
			handleExists:                 func(context.Context, string) (bool, error) { return false, nil },
			revokeActiveSession:          func(context.Context, string, time.Time) (SessionRecord, error) { return SessionRecord{}, nil },
		},
		fanAuthViewerReaderStub{
			readCurrentViewer: func(context.Context, string) (Bootstrap, error) {
				return Bootstrap{CurrentViewer: &CurrentViewer{ID: userID, ActiveMode: ActiveModeFan}}, nil
			},
		},
		defaultSignUpDraftStoreStub(),
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	if _, err := service.ReAuthenticate(context.Background(), "raw-session-token", "VeryStrongPass123!"); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("ReAuthenticate() error got %v want %v", err, ErrAuthenticationRequired)
	}
}

func TestFanAuthServiceReAuthenticateRequiresCurrentViewer(t *testing.T) {
	t.Parallel()

	service := NewFanAuthService(
		fanAuthProviderStub{},
		defaultFanAuthSessionManagerStub(),
		defaultFanAuthRepositoryStub(),
		fanAuthViewerReaderStub{
			readCurrentViewer: func(context.Context, string) (Bootstrap, error) {
				return Bootstrap{}, nil
			},
		},
		defaultSignUpDraftStoreStub(),
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	if _, err := service.ReAuthenticate(context.Background(), "raw-session-token", "VeryStrongPass123!"); !errors.Is(err, ErrAuthenticationRequired) {
		t.Fatalf("ReAuthenticate() error got %v want %v", err, ErrAuthenticationRequired)
	}
}

func TestFanAuthServiceReAuthenticateMapsSignInError(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	service := NewFanAuthService(
		fanAuthProviderStub{
			signIn: func(context.Context, string, string) (CognitoSessionInput, error) {
				return CognitoSessionInput{}, genericAPIError("NotAuthorizedException")
			},
		},
		defaultFanAuthSessionManagerStub(),
		fanAuthRepositoryStub{
			getIdentityByNormalizedEmail: func(context.Context, string) (Identity, error) { return Identity{}, ErrIdentityNotFound },
			getPreferredEmailByUserID:    func(context.Context, uuid.UUID) (string, error) { return "fan@example.com", nil },
			handleExists:                 func(context.Context, string) (bool, error) { return false, nil },
			revokeActiveSession:          func(context.Context, string, time.Time) (SessionRecord, error) { return SessionRecord{}, nil },
		},
		fanAuthViewerReaderStub{
			readCurrentViewer: func(context.Context, string) (Bootstrap, error) {
				return Bootstrap{CurrentViewer: &CurrentViewer{ID: userID, ActiveMode: ActiveModeFan}}, nil
			},
		},
		defaultSignUpDraftStoreStub(),
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	if _, err := service.ReAuthenticate(context.Background(), "raw-session-token", "VeryStrongPass123!"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("ReAuthenticate() error got %v want %v", err, ErrInvalidCredentials)
	}
}

func TestFanAuthServiceLogoutSuppressesMissingSession(t *testing.T) {
	t.Parallel()

	service := NewFanAuthService(
		fanAuthProviderStub{},
		defaultFanAuthSessionManagerStub(),
		fanAuthRepositoryStub{
			getIdentityByNormalizedEmail: func(context.Context, string) (Identity, error) { return Identity{}, ErrIdentityNotFound },
			getPreferredEmailByUserID:    func(context.Context, uuid.UUID) (string, error) { return "", ErrIdentityNotFound },
			handleExists:                 func(context.Context, string) (bool, error) { return false, nil },
			revokeActiveSession: func(context.Context, string, time.Time) (SessionRecord, error) {
				return SessionRecord{}, ErrSessionNotFound
			},
		},
		defaultFanAuthViewerReaderStub(),
		defaultSignUpDraftStoreStub(),
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	if err := service.Logout(context.Background(), "raw-session-token"); err != nil {
		t.Fatalf("Logout() error = %v, want nil", err)
	}
}

func TestFanAuthServiceLogoutReturnsUnexpectedRepositoryError(t *testing.T) {
	t.Parallel()

	service := NewFanAuthService(
		fanAuthProviderStub{},
		defaultFanAuthSessionManagerStub(),
		fanAuthRepositoryStub{
			getIdentityByNormalizedEmail: func(context.Context, string) (Identity, error) { return Identity{}, ErrIdentityNotFound },
			getPreferredEmailByUserID:    func(context.Context, uuid.UUID) (string, error) { return "", ErrIdentityNotFound },
			handleExists:                 func(context.Context, string) (bool, error) { return false, nil },
			revokeActiveSession: func(context.Context, string, time.Time) (SessionRecord, error) {
				return SessionRecord{}, errors.New("boom")
			},
		},
		defaultFanAuthViewerReaderStub(),
		defaultSignUpDraftStoreStub(),
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)

	if err := service.Logout(context.Background(), "raw-session-token"); err == nil || !strings.Contains(err.Error(), "logout session revoke") {
		t.Fatalf("Logout() error got %v want wrapped repository error", err)
	}
}

func TestFanAuthServiceLogoutAllowsEmptyToken(t *testing.T) {
	t.Parallel()

	service := &FanAuthService{repository: defaultFanAuthRepositoryStub()}
	if err := service.Logout(context.Background(), " "); err != nil {
		t.Fatalf("Logout() error = %v, want nil", err)
	}
}

func TestFanAuthServiceValidateRuntimeRequiresDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		service *FanAuthService
		wantErr string
	}{
		{
			name:    "nil service",
			service: nil,
			wantErr: "fan auth service is nil",
		},
		{
			name:    "missing provider",
			service: &FanAuthService{},
			wantErr: "provider is not initialized",
		},
		{
			name: "missing session manager",
			service: &FanAuthService{
				provider: fanAuthProviderStub{},
			},
			wantErr: "session manager is not initialized",
		},
		{
			name: "missing repository",
			service: &FanAuthService{
				provider:       fanAuthProviderStub{},
				sessionManager: defaultFanAuthSessionManagerStub(),
			},
			wantErr: "repository is not initialized",
		},
		{
			name: "missing viewer reader",
			service: &FanAuthService{
				provider:       fanAuthProviderStub{},
				sessionManager: defaultFanAuthSessionManagerStub(),
				repository:     defaultFanAuthRepositoryStub(),
			},
			wantErr: "viewer reader is not initialized",
		},
		{
			name: "missing draft store",
			service: &FanAuthService{
				provider:       fanAuthProviderStub{},
				sessionManager: defaultFanAuthSessionManagerStub(),
				repository:     defaultFanAuthRepositoryStub(),
				viewerReader:   defaultFanAuthViewerReaderStub(),
			},
			wantErr: "draft store is not initialized",
		},
		{
			name: "missing cooldown store",
			service: &FanAuthService{
				provider:       fanAuthProviderStub{},
				sessionManager: defaultFanAuthSessionManagerStub(),
				repository:     defaultFanAuthRepositoryStub(),
				viewerReader:   defaultFanAuthViewerReaderStub(),
				draftStore:     defaultSignUpDraftStoreStub(),
			},
			wantErr: "cooldown store is not initialized",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.service.validateRuntime()
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateRuntime() error got %v want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestMapSignInError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr error
	}{
		{
			name:    "invalid credentials",
			err:     genericAPIError("NotAuthorizedException"),
			wantErr: ErrInvalidCredentials,
		},
		{
			name:    "confirmation required",
			err:     genericAPIError("UserNotConfirmedException"),
			wantErr: ErrConfirmationRequired,
		},
		{
			name:    "rate limited",
			err:     genericAPIError("TooManyRequestsException"),
			wantErr: ErrRateLimited,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := mapSignInError(tt.err); !errors.Is(got, tt.wantErr) {
				t.Fatalf("mapSignInError() got %v want %v", got, tt.wantErr)
			}
		})
	}
}

func TestMapConfirmSignUpError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr error
	}{
		{
			name:    "invalid code",
			err:     genericAPIError("CodeMismatchException"),
			wantErr: ErrInvalidConfirmationCode,
		},
		{
			name:    "expired code",
			err:     genericAPIError("ExpiredCodeException"),
			wantErr: ErrConfirmationCodeExpired,
		},
		{
			name:    "rate limited",
			err:     genericAPIError("TooManyRequestsException"),
			wantErr: ErrRateLimited,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := mapConfirmSignUpError(tt.err); !errors.Is(got, tt.wantErr) {
				t.Fatalf("mapConfirmSignUpError() got %v want %v", got, tt.wantErr)
			}
		})
	}
}

func TestIsRateLimitError(t *testing.T) {
	t.Parallel()

	if !isRateLimitError(genericAPIError("TooManyRequestsException")) {
		t.Fatal("isRateLimitError() got false want true for TooManyRequestsException")
	}
	if isRateLimitError(genericAPIError("UserNotConfirmedException")) {
		t.Fatal("isRateLimitError() got true want false for non-rate-limit error")
	}
}

func newTestFanAuthService(provider fanAuthProviderStub) *FanAuthService {
	return NewFanAuthService(
		provider,
		defaultFanAuthSessionManagerStub(),
		defaultFanAuthRepositoryStub(),
		defaultFanAuthViewerReaderStub(),
		defaultSignUpDraftStoreStub(),
		authCooldownStoreStub{
			release:     func(context.Context, string) error { return nil },
			tryActivate: func(context.Context, string, time.Duration) (bool, error) { return true, nil },
		},
	)
}

func defaultFanAuthSessionManagerStub() fanAuthSessionManagerStub {
	return fanAuthSessionManagerStub{
		startSession: func(context.Context, CognitoSessionInput) (AuthenticatedSession, error) {
			return AuthenticatedSession{}, nil
		},
		startSignUpSession: func(context.Context, CognitoSessionInput, string, string) (AuthenticatedSession, error) {
			return AuthenticatedSession{}, nil
		},
	}
}

func defaultFanAuthRepositoryStub() fanAuthRepositoryStub {
	return fanAuthRepositoryStub{
		getIdentityByProviderAndSubject: func(context.Context, string, string) (Identity, error) { return Identity{}, ErrIdentityNotFound },
		getIdentityByNormalizedEmail: func(context.Context, string) (Identity, error) { return Identity{}, ErrIdentityNotFound },
		getPreferredEmailByUserID:    func(context.Context, uuid.UUID) (string, error) { return "", ErrIdentityNotFound },
		handleExists:                 func(context.Context, string) (bool, error) { return false, nil },
		revokeActiveSession:          func(context.Context, string, time.Time) (SessionRecord, error) { return SessionRecord{}, nil },
		updateActiveModeByTokenHash:  func(context.Context, string, ActiveMode) (SessionRecord, error) { return SessionRecord{}, nil },
	}
}

func defaultFanAuthViewerReaderStub() fanAuthViewerReaderStub {
	return fanAuthViewerReaderStub{
		readCurrentViewer: func(context.Context, string) (Bootstrap, error) { return Bootstrap{}, nil },
	}
}

func defaultSignUpDraftStoreStub() signUpDraftStoreStub {
	return signUpDraftStoreStub{
		getDraft:    func(context.Context, string) (SignUpDraft, error) { return SignUpDraft{}, ErrSignUpDraftNotFound },
		deleteDraft: func(context.Context, string) error { return nil },
		saveDraft:   func(context.Context, string, SignUpDraft, time.Duration) error { return nil },
	}
}

func genericAPIError(code string) error {
	return &smithy.GenericAPIError{Code: code, Message: code}
}
