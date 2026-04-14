package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type cognitoSessionRepositoryStub struct {
	getIdentityByProviderAndSubject           func(context.Context, string, string) (Identity, error)
	getIdentityByEmail                        func(context.Context, string) (Identity, error)
	createIdentity                            func(context.Context, CreateIdentityInput) (Identity, error)
	createSession                             func(context.Context, CreateSessionInput) (SessionRecord, error)
	createUserWithIdentityAndSession          func(context.Context, CreateUserWithIdentityAndSessionInput) (SessionRecord, error)
	recordIdentityAuthentication              func(context.Context, RecordIdentityAuthenticationInput) (Identity, error)
	refreshSessionRecentAuthenticatedAtByHash func(context.Context, string, time.Time) (SessionRecord, error)
}

func (s cognitoSessionRepositoryStub) GetIdentityByProviderAndSubject(
	ctx context.Context,
	provider string,
	providerSubject string,
) (Identity, error) {
	return s.getIdentityByProviderAndSubject(ctx, provider, providerSubject)
}

func (s cognitoSessionRepositoryStub) GetIdentityByEmail(ctx context.Context, emailNormalized string) (Identity, error) {
	return s.getIdentityByEmail(ctx, emailNormalized)
}

func (s cognitoSessionRepositoryStub) CreateIdentity(ctx context.Context, input CreateIdentityInput) (Identity, error) {
	return s.createIdentity(ctx, input)
}

func (s cognitoSessionRepositoryStub) CreateSession(ctx context.Context, input CreateSessionInput) (SessionRecord, error) {
	return s.createSession(ctx, input)
}

func (s cognitoSessionRepositoryStub) CreateUserWithIdentityAndSession(
	ctx context.Context,
	input CreateUserWithIdentityAndSessionInput,
) (SessionRecord, error) {
	return s.createUserWithIdentityAndSession(ctx, input)
}

func (s cognitoSessionRepositoryStub) RecordIdentityAuthentication(
	ctx context.Context,
	input RecordIdentityAuthenticationInput,
) (Identity, error) {
	return s.recordIdentityAuthentication(ctx, input)
}

func (s cognitoSessionRepositoryStub) RefreshSessionRecentAuthenticatedAtByTokenHash(
	ctx context.Context,
	sessionTokenHash string,
	recentAuthenticatedAt time.Time,
) (SessionRecord, error) {
	return s.refreshSessionRecentAuthenticatedAtByHash(ctx, sessionTokenHash, recentAuthenticatedAt)
}

func TestNewCognitoSessionManagerInitializesDefaults(t *testing.T) {
	t.Parallel()

	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{})
	if manager == nil {
		t.Fatal("NewCognitoSessionManager() = nil")
	}
	if manager.repository == nil {
		t.Fatal("NewCognitoSessionManager() repository = nil")
	}
	if manager.now == nil {
		t.Fatal("NewCognitoSessionManager() now = nil")
	}
	if manager.newSessionToken == nil {
		t.Fatal("NewCognitoSessionManager() newSessionToken = nil")
	}
}

func TestCognitoSessionManagerStartSessionUsesExistingIdentity(t *testing.T) {
	t.Parallel()

	now := time.Unix(1711000000, 0).UTC()
	userID := uuid.New()
	identityID := uuid.New()

	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{
		getIdentityByProviderAndSubject: func(_ context.Context, provider string, providerSubject string) (Identity, error) {
			if provider != identityProviderCognito {
				t.Fatalf("GetIdentityByProviderAndSubject() provider got %q want %q", provider, identityProviderCognito)
			}
			if providerSubject != "cognito-subject" {
				t.Fatalf("GetIdentityByProviderAndSubject() subject got %q want %q", providerSubject, "cognito-subject")
			}

			return Identity{ID: identityID, UserID: userID}, nil
		},
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			t.Fatal("GetIdentityByEmail() should not be called")
			return Identity{}, nil
		},
		createIdentity: func(context.Context, CreateIdentityInput) (Identity, error) {
			t.Fatal("CreateIdentity() should not be called")
			return Identity{}, nil
		},
		recordIdentityAuthentication: func(_ context.Context, input RecordIdentityAuthenticationInput) (Identity, error) {
			if input.ID != identityID {
				t.Fatalf("RecordIdentityAuthentication() id got %s want %s", input.ID, identityID)
			}
			if input.EmailNormalized != "fan@example.com" {
				t.Fatalf("RecordIdentityAuthentication() email got %q want %q", input.EmailNormalized, "fan@example.com")
			}
			if input.VerifiedAt == nil || !input.VerifiedAt.Equal(now) {
				t.Fatalf("RecordIdentityAuthentication() verifiedAt got %v want %s", input.VerifiedAt, now)
			}

			return Identity{}, nil
		},
		createSession: func(_ context.Context, input CreateSessionInput) (SessionRecord, error) {
			if input.UserID != userID {
				t.Fatalf("CreateSession() userID got %s want %s", input.UserID, userID)
			}
			if input.RecentAuthenticatedAt != now {
				t.Fatalf("CreateSession() recent_authenticated_at got %s want %s", input.RecentAuthenticatedAt, now)
			}
			if input.SessionTokenHash != HashSessionToken("session-token") {
				t.Fatalf("CreateSession() session hash got %q want %q", input.SessionTokenHash, HashSessionToken("session-token"))
			}

			return SessionRecord{ExpiresAt: input.ExpiresAt}, nil
		},
		createUserWithIdentityAndSession: func(context.Context, CreateUserWithIdentityAndSessionInput) (SessionRecord, error) {
			t.Fatal("CreateUserWithIdentityAndSession() should not be called")
			return SessionRecord{}, nil
		},
		refreshSessionRecentAuthenticatedAtByHash: func(context.Context, string, time.Time) (SessionRecord, error) {
			t.Fatal("RefreshSessionRecentAuthenticatedAtByTokenHash() should not be called")
			return SessionRecord{}, nil
		},
	})
	manager.now = func() time.Time { return now }
	manager.newSessionToken = func() (string, error) { return "session-token", nil }

	got, err := manager.StartSession(context.Background(), CognitoSessionInput{
		Subject:         "cognito-subject",
		Email:           "fan@example.com",
		EmailVerified:   true,
		AuthenticatedAt: now,
	})
	if err != nil {
		t.Fatalf("StartSession() error = %v, want nil", err)
	}
	if got.Token != "session-token" {
		t.Fatalf("StartSession() token got %q want %q", got.Token, "session-token")
	}
}

func TestCognitoSessionManagerStartSessionBridgesLegacyIdentity(t *testing.T) {
	t.Parallel()

	now := time.Unix(1711000100, 0).UTC()
	userID := uuid.New()
	identityID := uuid.New()

	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{
		getIdentityByProviderAndSubject: func(context.Context, string, string) (Identity, error) {
			return Identity{}, ErrIdentityNotFound
		},
		getIdentityByEmail: func(_ context.Context, emailNormalized string) (Identity, error) {
			if emailNormalized != "fan@example.com" {
				t.Fatalf("GetIdentityByEmail() email got %q want %q", emailNormalized, "fan@example.com")
			}

			return Identity{UserID: userID}, nil
		},
		createIdentity: func(_ context.Context, input CreateIdentityInput) (Identity, error) {
			if input.UserID != userID {
				t.Fatalf("CreateIdentity() userID got %s want %s", input.UserID, userID)
			}
			if input.Provider != identityProviderCognito {
				t.Fatalf("CreateIdentity() provider got %q want %q", input.Provider, identityProviderCognito)
			}
			if input.ProviderSubject != "cognito-subject" {
				t.Fatalf("CreateIdentity() subject got %q want %q", input.ProviderSubject, "cognito-subject")
			}
			if input.EmailNormalized == nil || *input.EmailNormalized != "fan@example.com" {
				t.Fatalf("CreateIdentity() email got %v want %q", input.EmailNormalized, "fan@example.com")
			}

			return Identity{ID: identityID, UserID: userID}, nil
		},
		recordIdentityAuthentication: func(_ context.Context, input RecordIdentityAuthenticationInput) (Identity, error) {
			if input.ID != identityID {
				t.Fatalf("RecordIdentityAuthentication() id got %s want %s", input.ID, identityID)
			}

			return Identity{}, nil
		},
		createSession: func(_ context.Context, input CreateSessionInput) (SessionRecord, error) {
			if input.UserID != userID {
				t.Fatalf("CreateSession() userID got %s want %s", input.UserID, userID)
			}

			return SessionRecord{ExpiresAt: input.ExpiresAt}, nil
		},
		createUserWithIdentityAndSession: func(context.Context, CreateUserWithIdentityAndSessionInput) (SessionRecord, error) {
			t.Fatal("CreateUserWithIdentityAndSession() should not be called")
			return SessionRecord{}, nil
		},
		refreshSessionRecentAuthenticatedAtByHash: func(context.Context, string, time.Time) (SessionRecord, error) {
			t.Fatal("RefreshSessionRecentAuthenticatedAtByTokenHash() should not be called")
			return SessionRecord{}, nil
		},
	})
	manager.now = func() time.Time { return now }
	manager.newSessionToken = func() (string, error) { return "session-token", nil }

	if _, err := manager.StartSession(context.Background(), CognitoSessionInput{
		Subject:         "cognito-subject",
		Email:           "fan@example.com",
		EmailVerified:   true,
		AuthenticatedAt: now,
	}); err != nil {
		t.Fatalf("StartSession() error = %v, want nil", err)
	}
}

func TestCognitoSessionManagerStartSessionCreatesNewUser(t *testing.T) {
	t.Parallel()

	now := time.Unix(1711000200, 0).UTC()

	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{
		getIdentityByProviderAndSubject: func(context.Context, string, string) (Identity, error) {
			return Identity{}, ErrIdentityNotFound
		},
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			return Identity{}, ErrIdentityNotFound
		},
		createIdentity: func(context.Context, CreateIdentityInput) (Identity, error) {
			t.Fatal("CreateIdentity() should not be called")
			return Identity{}, nil
		},
		createSession: func(context.Context, CreateSessionInput) (SessionRecord, error) {
			t.Fatal("CreateSession() should not be called")
			return SessionRecord{}, nil
		},
		createUserWithIdentityAndSession: func(_ context.Context, input CreateUserWithIdentityAndSessionInput) (SessionRecord, error) {
			if input.Provider != identityProviderCognito {
				t.Fatalf("CreateUserWithIdentityAndSession() provider got %q want %q", input.Provider, identityProviderCognito)
			}
			if input.ProviderSubject != "cognito-subject" {
				t.Fatalf("CreateUserWithIdentityAndSession() subject got %q want %q", input.ProviderSubject, "cognito-subject")
			}
			if input.EmailNormalized == nil || *input.EmailNormalized != "fan@example.com" {
				t.Fatalf("CreateUserWithIdentityAndSession() email got %v want %q", input.EmailNormalized, "fan@example.com")
			}
			if input.RecentAuthenticatedAt != now {
				t.Fatalf("CreateUserWithIdentityAndSession() recent_authenticated_at got %s want %s", input.RecentAuthenticatedAt, now)
			}

			return SessionRecord{ExpiresAt: input.ExpiresAt}, nil
		},
		recordIdentityAuthentication: func(context.Context, RecordIdentityAuthenticationInput) (Identity, error) {
			t.Fatal("RecordIdentityAuthentication() should not be called")
			return Identity{}, nil
		},
		refreshSessionRecentAuthenticatedAtByHash: func(context.Context, string, time.Time) (SessionRecord, error) {
			t.Fatal("RefreshSessionRecentAuthenticatedAtByTokenHash() should not be called")
			return SessionRecord{}, nil
		},
	})
	manager.now = func() time.Time { return now }
	manager.newSessionToken = func() (string, error) { return "session-token", nil }

	got, err := manager.StartSession(context.Background(), CognitoSessionInput{
		Subject:         "cognito-subject",
		Email:           "fan@example.com",
		EmailVerified:   true,
		AuthenticatedAt: now,
	})
	if err != nil {
		t.Fatalf("StartSession() error = %v, want nil", err)
	}
	if got.Token != "session-token" {
		t.Fatalf("StartSession() token got %q want %q", got.Token, "session-token")
	}
}

func TestCognitoSessionManagerStartSessionBridgesLegacyIdentityAfterCreateConflict(t *testing.T) {
	t.Parallel()

	now := time.Unix(1711000210, 0).UTC()
	userID := uuid.New()
	identityID := uuid.New()
	getIdentityByEmailCallCount := 0

	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{
		getIdentityByProviderAndSubject: func(context.Context, string, string) (Identity, error) {
			return Identity{}, ErrIdentityNotFound
		},
		getIdentityByEmail: func(_ context.Context, emailNormalized string) (Identity, error) {
			getIdentityByEmailCallCount++
			if emailNormalized != "fan@example.com" {
				t.Fatalf("GetIdentityByEmail() email got %q want %q", emailNormalized, "fan@example.com")
			}
			if getIdentityByEmailCallCount == 1 {
				return Identity{}, ErrIdentityNotFound
			}

			return Identity{UserID: userID}, nil
		},
		createIdentity: func(_ context.Context, input CreateIdentityInput) (Identity, error) {
			if input.UserID != userID {
				t.Fatalf("CreateIdentity() userID got %s want %s", input.UserID, userID)
			}

			return Identity{ID: identityID, UserID: userID}, nil
		},
		createSession: func(_ context.Context, input CreateSessionInput) (SessionRecord, error) {
			if input.UserID != userID {
				t.Fatalf("CreateSession() userID got %s want %s", input.UserID, userID)
			}
			if input.SessionTokenHash != HashSessionToken("session-token") {
				t.Fatalf("CreateSession() session hash got %q want %q", input.SessionTokenHash, HashSessionToken("session-token"))
			}
			if input.RecentAuthenticatedAt != now {
				t.Fatalf("CreateSession() recent_authenticated_at got %s want %s", input.RecentAuthenticatedAt, now)
			}

			return SessionRecord{ExpiresAt: input.ExpiresAt}, nil
		},
		createUserWithIdentityAndSession: func(context.Context, CreateUserWithIdentityAndSessionInput) (SessionRecord, error) {
			return SessionRecord{}, ErrIdentityAlreadyExists
		},
		recordIdentityAuthentication: func(_ context.Context, input RecordIdentityAuthenticationInput) (Identity, error) {
			if input.ID != identityID {
				t.Fatalf("RecordIdentityAuthentication() id got %s want %s", input.ID, identityID)
			}

			return Identity{}, nil
		},
		refreshSessionRecentAuthenticatedAtByHash: func(context.Context, string, time.Time) (SessionRecord, error) {
			t.Fatal("RefreshSessionRecentAuthenticatedAtByTokenHash() should not be called")
			return SessionRecord{}, nil
		},
	})
	manager.now = func() time.Time { return now }
	manager.newSessionToken = func() (string, error) { return "session-token", nil }

	got, err := manager.StartSession(context.Background(), CognitoSessionInput{
		Subject:         "cognito-subject",
		Email:           "fan@example.com",
		EmailVerified:   true,
		AuthenticatedAt: now,
	})
	if err != nil {
		t.Fatalf("StartSession() error = %v, want nil", err)
	}
	if got.Token != "session-token" {
		t.Fatalf("StartSession() token got %q want %q", got.Token, "session-token")
	}
	if getIdentityByEmailCallCount != 2 {
		t.Fatalf("GetIdentityByEmail() call count got %d want %d", getIdentityByEmailCallCount, 2)
	}
}

func TestCognitoSessionManagerStartSessionRejectsUnverifiedEmail(t *testing.T) {
	t.Parallel()

	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{})

	if _, err := manager.StartSession(context.Background(), CognitoSessionInput{
		Subject:       "cognito-subject",
		Email:         "fan@example.com",
		EmailVerified: false,
	}); !errors.Is(err, ErrEmailVerificationRequired) {
		t.Fatalf("StartSession() error got %v want %v", err, ErrEmailVerificationRequired)
	}
}

func TestCognitoSessionManagerStartSessionRejectsBlankSubject(t *testing.T) {
	t.Parallel()

	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{})

	if _, err := manager.StartSession(context.Background(), CognitoSessionInput{
		Subject:       "   ",
		Email:         "fan@example.com",
		EmailVerified: true,
	}); !errors.Is(err, ErrInvalidProviderSubject) {
		t.Fatalf("StartSession() error got %v want %v", err, ErrInvalidProviderSubject)
	}
}

func TestCognitoSessionManagerStartSessionRejectsInvalidEmail(t *testing.T) {
	t.Parallel()

	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{})

	if _, err := manager.StartSession(context.Background(), CognitoSessionInput{
		Subject:       "cognito-subject",
		Email:         "not-an-email",
		EmailVerified: true,
	}); err == nil {
		t.Fatal("StartSession() error = nil, want invalid email error")
	}
}

func TestCognitoSessionManagerStartSessionUsesNowWhenAuthenticatedAtZero(t *testing.T) {
	t.Parallel()

	now := time.Unix(1711000250, 0).UTC()
	userID := uuid.New()
	identityID := uuid.New()

	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{
		getIdentityByProviderAndSubject: func(context.Context, string, string) (Identity, error) {
			return Identity{ID: identityID, UserID: userID}, nil
		},
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			t.Fatal("GetIdentityByEmail() should not be called")
			return Identity{}, nil
		},
		createIdentity: func(context.Context, CreateIdentityInput) (Identity, error) {
			t.Fatal("CreateIdentity() should not be called")
			return Identity{}, nil
		},
		recordIdentityAuthentication: func(_ context.Context, input RecordIdentityAuthenticationInput) (Identity, error) {
			if input.VerifiedAt == nil || !input.VerifiedAt.Equal(now) {
				t.Fatalf("RecordIdentityAuthentication() verifiedAt got %v want %s", input.VerifiedAt, now)
			}

			return Identity{}, nil
		},
		createSession: func(_ context.Context, input CreateSessionInput) (SessionRecord, error) {
			if input.RecentAuthenticatedAt != now {
				t.Fatalf("CreateSession() recent_authenticated_at got %s want %s", input.RecentAuthenticatedAt, now)
			}
			if !input.ExpiresAt.Equal(now.Add(defaultSessionTTL)) {
				t.Fatalf("CreateSession() expires_at got %s want %s", input.ExpiresAt, now.Add(defaultSessionTTL))
			}

			return SessionRecord{ExpiresAt: input.ExpiresAt}, nil
		},
		createUserWithIdentityAndSession: func(context.Context, CreateUserWithIdentityAndSessionInput) (SessionRecord, error) {
			t.Fatal("CreateUserWithIdentityAndSession() should not be called")
			return SessionRecord{}, nil
		},
		refreshSessionRecentAuthenticatedAtByHash: func(context.Context, string, time.Time) (SessionRecord, error) {
			t.Fatal("RefreshSessionRecentAuthenticatedAtByTokenHash() should not be called")
			return SessionRecord{}, nil
		},
	})
	manager.now = func() time.Time { return now }
	manager.newSessionToken = func() (string, error) { return "session-token", nil }

	if _, err := manager.StartSession(context.Background(), CognitoSessionInput{
		Subject:       "cognito-subject",
		Email:         "fan@example.com",
		EmailVerified: true,
	}); err != nil {
		t.Fatalf("StartSession() error = %v, want nil", err)
	}
}

func TestCognitoSessionManagerStartSessionReusesConcurrentCognitoIdentity(t *testing.T) {
	t.Parallel()

	now := time.Unix(1711000275, 0).UTC()
	userID := uuid.New()
	identityID := uuid.New()
	getIdentityCallCount := 0

	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{
		getIdentityByProviderAndSubject: func(_ context.Context, provider string, providerSubject string) (Identity, error) {
			getIdentityCallCount++
			if provider != identityProviderCognito {
				t.Fatalf("GetIdentityByProviderAndSubject() provider got %q want %q", provider, identityProviderCognito)
			}
			if providerSubject != "cognito-subject" {
				t.Fatalf("GetIdentityByProviderAndSubject() subject got %q want %q", providerSubject, "cognito-subject")
			}
			if getIdentityCallCount == 1 {
				return Identity{}, ErrIdentityNotFound
			}

			return Identity{ID: identityID, UserID: userID}, nil
		},
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			return Identity{UserID: userID}, nil
		},
		createIdentity: func(context.Context, CreateIdentityInput) (Identity, error) {
			return Identity{}, ErrIdentityAlreadyExists
		},
		recordIdentityAuthentication: func(_ context.Context, input RecordIdentityAuthenticationInput) (Identity, error) {
			if input.ID != identityID {
				t.Fatalf("RecordIdentityAuthentication() id got %s want %s", input.ID, identityID)
			}

			return Identity{}, nil
		},
		createSession: func(_ context.Context, input CreateSessionInput) (SessionRecord, error) {
			if input.UserID != userID {
				t.Fatalf("CreateSession() userID got %s want %s", input.UserID, userID)
			}

			return SessionRecord{ExpiresAt: input.ExpiresAt}, nil
		},
		createUserWithIdentityAndSession: func(context.Context, CreateUserWithIdentityAndSessionInput) (SessionRecord, error) {
			t.Fatal("CreateUserWithIdentityAndSession() should not be called")
			return SessionRecord{}, nil
		},
		refreshSessionRecentAuthenticatedAtByHash: func(context.Context, string, time.Time) (SessionRecord, error) {
			t.Fatal("RefreshSessionRecentAuthenticatedAtByTokenHash() should not be called")
			return SessionRecord{}, nil
		},
	})
	manager.now = func() time.Time { return now }
	manager.newSessionToken = func() (string, error) { return "session-token", nil }

	if _, err := manager.StartSession(context.Background(), CognitoSessionInput{
		Subject:         "cognito-subject",
		Email:           "fan@example.com",
		EmailVerified:   true,
		AuthenticatedAt: now,
	}); err != nil {
		t.Fatalf("StartSession() error = %v, want nil", err)
	}
	if getIdentityCallCount != 2 {
		t.Fatalf("GetIdentityByProviderAndSubject() call count got %d want %d", getIdentityCallCount, 2)
	}
}

func TestCognitoSessionManagerStartSessionWrapsCreateUserTokenError(t *testing.T) {
	t.Parallel()

	now := time.Unix(1711000280, 0).UTC()
	expectedErr := errors.New("token generation failed")

	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{
		getIdentityByProviderAndSubject: func(context.Context, string, string) (Identity, error) {
			return Identity{}, ErrIdentityNotFound
		},
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			return Identity{}, ErrIdentityNotFound
		},
		createIdentity: func(context.Context, CreateIdentityInput) (Identity, error) {
			t.Fatal("CreateIdentity() should not be called")
			return Identity{}, nil
		},
		createSession: func(context.Context, CreateSessionInput) (SessionRecord, error) {
			t.Fatal("CreateSession() should not be called")
			return SessionRecord{}, nil
		},
		createUserWithIdentityAndSession: func(context.Context, CreateUserWithIdentityAndSessionInput) (SessionRecord, error) {
			t.Fatal("CreateUserWithIdentityAndSession() should not be called")
			return SessionRecord{}, nil
		},
		recordIdentityAuthentication: func(context.Context, RecordIdentityAuthenticationInput) (Identity, error) {
			t.Fatal("RecordIdentityAuthentication() should not be called")
			return Identity{}, nil
		},
		refreshSessionRecentAuthenticatedAtByHash: func(context.Context, string, time.Time) (SessionRecord, error) {
			t.Fatal("RefreshSessionRecentAuthenticatedAtByTokenHash() should not be called")
			return SessionRecord{}, nil
		},
	})
	manager.now = func() time.Time { return now }
	manager.newSessionToken = func() (string, error) { return "", expectedErr }

	if _, err := manager.StartSession(context.Background(), CognitoSessionInput{
		Subject:         "cognito-subject",
		Email:           "fan@example.com",
		EmailVerified:   true,
		AuthenticatedAt: now,
	}); !errors.Is(err, expectedErr) {
		t.Fatalf("StartSession() error got %v want %v", err, expectedErr)
	}
}

func TestCognitoSessionManagerStartSessionWrapsSubjectLookupError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("subject lookup failed")
	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{
		getIdentityByProviderAndSubject: func(context.Context, string, string) (Identity, error) {
			return Identity{}, expectedErr
		},
	})

	if _, err := manager.StartSession(context.Background(), CognitoSessionInput{
		Subject:       "cognito-subject",
		Email:         "fan@example.com",
		EmailVerified: true,
	}); !errors.Is(err, expectedErr) {
		t.Fatalf("StartSession() error got %v want %v", err, expectedErr)
	}
}

func TestCognitoSessionManagerStartSessionWrapsLegacyEmailLookupError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("legacy email lookup failed")
	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{
		getIdentityByProviderAndSubject: func(context.Context, string, string) (Identity, error) {
			return Identity{}, ErrIdentityNotFound
		},
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			return Identity{}, expectedErr
		},
	})

	if _, err := manager.StartSession(context.Background(), CognitoSessionInput{
		Subject:       "cognito-subject",
		Email:         "fan@example.com",
		EmailVerified: true,
	}); !errors.Is(err, expectedErr) {
		t.Fatalf("StartSession() error got %v want %v", err, expectedErr)
	}
}

func TestCognitoSessionManagerStartSessionWrapsRecordAuthenticationError(t *testing.T) {
	t.Parallel()

	now := time.Unix(1711000290, 0).UTC()
	expectedErr := errors.New("record failed")
	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{
		getIdentityByProviderAndSubject: func(context.Context, string, string) (Identity, error) {
			return Identity{ID: uuid.New(), UserID: uuid.New()}, nil
		},
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			t.Fatal("GetIdentityByEmail() should not be called")
			return Identity{}, nil
		},
		createIdentity: func(context.Context, CreateIdentityInput) (Identity, error) {
			t.Fatal("CreateIdentity() should not be called")
			return Identity{}, nil
		},
		recordIdentityAuthentication: func(context.Context, RecordIdentityAuthenticationInput) (Identity, error) {
			return Identity{}, expectedErr
		},
		createSession: func(context.Context, CreateSessionInput) (SessionRecord, error) {
			t.Fatal("CreateSession() should not be called")
			return SessionRecord{}, nil
		},
		createUserWithIdentityAndSession: func(context.Context, CreateUserWithIdentityAndSessionInput) (SessionRecord, error) {
			t.Fatal("CreateUserWithIdentityAndSession() should not be called")
			return SessionRecord{}, nil
		},
		refreshSessionRecentAuthenticatedAtByHash: func(context.Context, string, time.Time) (SessionRecord, error) {
			t.Fatal("RefreshSessionRecentAuthenticatedAtByTokenHash() should not be called")
			return SessionRecord{}, nil
		},
	})

	if _, err := manager.StartSession(context.Background(), CognitoSessionInput{
		Subject:         "cognito-subject",
		Email:           "fan@example.com",
		EmailVerified:   true,
		AuthenticatedAt: now,
	}); !errors.Is(err, expectedErr) {
		t.Fatalf("StartSession() error got %v want %v", err, expectedErr)
	}
}

func TestCognitoSessionManagerStartSessionWrapsSessionCreateError(t *testing.T) {
	t.Parallel()

	now := time.Unix(1711000295, 0).UTC()
	expectedErr := errors.New("create session failed")
	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{
		getIdentityByProviderAndSubject: func(context.Context, string, string) (Identity, error) {
			return Identity{ID: uuid.New(), UserID: uuid.New()}, nil
		},
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			t.Fatal("GetIdentityByEmail() should not be called")
			return Identity{}, nil
		},
		createIdentity: func(context.Context, CreateIdentityInput) (Identity, error) {
			t.Fatal("CreateIdentity() should not be called")
			return Identity{}, nil
		},
		recordIdentityAuthentication: func(context.Context, RecordIdentityAuthenticationInput) (Identity, error) {
			return Identity{}, nil
		},
		createSession: func(context.Context, CreateSessionInput) (SessionRecord, error) {
			return SessionRecord{}, expectedErr
		},
		createUserWithIdentityAndSession: func(context.Context, CreateUserWithIdentityAndSessionInput) (SessionRecord, error) {
			t.Fatal("CreateUserWithIdentityAndSession() should not be called")
			return SessionRecord{}, nil
		},
		refreshSessionRecentAuthenticatedAtByHash: func(context.Context, string, time.Time) (SessionRecord, error) {
			t.Fatal("RefreshSessionRecentAuthenticatedAtByTokenHash() should not be called")
			return SessionRecord{}, nil
		},
	})
	manager.now = func() time.Time { return now }
	manager.newSessionToken = func() (string, error) { return "session-token", nil }

	if _, err := manager.StartSession(context.Background(), CognitoSessionInput{
		Subject:         "cognito-subject",
		Email:           "fan@example.com",
		EmailVerified:   true,
		AuthenticatedAt: now,
	}); !errors.Is(err, expectedErr) {
		t.Fatalf("StartSession() error got %v want %v", err, expectedErr)
	}
}

func TestCognitoSessionManagerStartSessionWrapsCreateUserError(t *testing.T) {
	t.Parallel()

	now := time.Unix(1711000298, 0).UTC()
	expectedErr := errors.New("create user failed")
	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{
		getIdentityByProviderAndSubject: func(context.Context, string, string) (Identity, error) {
			return Identity{}, ErrIdentityNotFound
		},
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			return Identity{}, ErrIdentityNotFound
		},
		createIdentity: func(context.Context, CreateIdentityInput) (Identity, error) {
			t.Fatal("CreateIdentity() should not be called")
			return Identity{}, nil
		},
		createSession: func(context.Context, CreateSessionInput) (SessionRecord, error) {
			t.Fatal("CreateSession() should not be called")
			return SessionRecord{}, nil
		},
		createUserWithIdentityAndSession: func(context.Context, CreateUserWithIdentityAndSessionInput) (SessionRecord, error) {
			return SessionRecord{}, expectedErr
		},
		recordIdentityAuthentication: func(context.Context, RecordIdentityAuthenticationInput) (Identity, error) {
			t.Fatal("RecordIdentityAuthentication() should not be called")
			return Identity{}, nil
		},
		refreshSessionRecentAuthenticatedAtByHash: func(context.Context, string, time.Time) (SessionRecord, error) {
			t.Fatal("RefreshSessionRecentAuthenticatedAtByTokenHash() should not be called")
			return SessionRecord{}, nil
		},
	})
	manager.now = func() time.Time { return now }
	manager.newSessionToken = func() (string, error) { return "session-token", nil }

	if _, err := manager.StartSession(context.Background(), CognitoSessionInput{
		Subject:         "cognito-subject",
		Email:           "fan@example.com",
		EmailVerified:   true,
		AuthenticatedAt: now,
	}); !errors.Is(err, expectedErr) {
		t.Fatalf("StartSession() error got %v want %v", err, expectedErr)
	}
}

func TestCognitoSessionManagerStartSessionRecordsAuthenticationWhenRecoveringConflictedUserCreation(t *testing.T) {
	t.Parallel()

	now := time.Unix(1711000299, 0).UTC()
	userID := uuid.New()
	identityID := uuid.New()
	getIdentityByProviderAndSubjectCallCount := 0
	recordAuthenticationCallCount := 0

	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{
		getIdentityByProviderAndSubject: func(_ context.Context, provider string, providerSubject string) (Identity, error) {
			getIdentityByProviderAndSubjectCallCount++
			if provider != identityProviderCognito {
				t.Fatalf("GetIdentityByProviderAndSubject() provider got %q want %q", provider, identityProviderCognito)
			}
			if providerSubject != "cognito-subject" {
				t.Fatalf("GetIdentityByProviderAndSubject() subject got %q want %q", providerSubject, "cognito-subject")
			}
			if getIdentityByProviderAndSubjectCallCount == 1 {
				return Identity{}, ErrIdentityNotFound
			}

			return Identity{ID: identityID, UserID: userID}, nil
		},
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			return Identity{}, ErrIdentityNotFound
		},
		createIdentity: func(context.Context, CreateIdentityInput) (Identity, error) {
			t.Fatal("CreateIdentity() should not be called")
			return Identity{}, nil
		},
		createSession: func(_ context.Context, input CreateSessionInput) (SessionRecord, error) {
			if input.UserID != userID {
				t.Fatalf("CreateSession() userID got %s want %s", input.UserID, userID)
			}
			if !input.RecentAuthenticatedAt.Equal(now) {
				t.Fatalf("CreateSession() recent_authenticated_at got %s want %s", input.RecentAuthenticatedAt, now)
			}

			return SessionRecord{ExpiresAt: input.ExpiresAt}, nil
		},
		createUserWithIdentityAndSession: func(context.Context, CreateUserWithIdentityAndSessionInput) (SessionRecord, error) {
			return SessionRecord{}, ErrIdentityAlreadyExists
		},
		recordIdentityAuthentication: func(_ context.Context, input RecordIdentityAuthenticationInput) (Identity, error) {
			recordAuthenticationCallCount++
			if input.ID != identityID {
				t.Fatalf("RecordIdentityAuthentication() id got %s want %s", input.ID, identityID)
			}
			if input.EmailNormalized != "fan@example.com" {
				t.Fatalf("RecordIdentityAuthentication() email got %q want %q", input.EmailNormalized, "fan@example.com")
			}
			if input.VerifiedAt == nil || !input.VerifiedAt.Equal(now) {
				t.Fatalf("RecordIdentityAuthentication() verifiedAt got %v want %s", input.VerifiedAt, now)
			}
			if !input.LastAuthenticatedAt.Equal(now) {
				t.Fatalf("RecordIdentityAuthentication() lastAuthenticatedAt got %s want %s", input.LastAuthenticatedAt, now)
			}

			return Identity{}, nil
		},
		refreshSessionRecentAuthenticatedAtByHash: func(context.Context, string, time.Time) (SessionRecord, error) {
			t.Fatal("RefreshSessionRecentAuthenticatedAtByTokenHash() should not be called")
			return SessionRecord{}, nil
		},
	})
	manager.now = func() time.Time { return now }
	manager.newSessionToken = func() (string, error) { return "session-token", nil }

	got, err := manager.StartSession(context.Background(), CognitoSessionInput{
		Subject:         "cognito-subject",
		Email:           "fan@example.com",
		EmailVerified:   true,
		AuthenticatedAt: now,
	})
	if err != nil {
		t.Fatalf("StartSession() error = %v, want nil", err)
	}
	if got.Token != "session-token" {
		t.Fatalf("StartSession() token got %q want %q", got.Token, "session-token")
	}
	if getIdentityByProviderAndSubjectCallCount != 2 {
		t.Fatalf("GetIdentityByProviderAndSubject() call count got %d want %d", getIdentityByProviderAndSubjectCallCount, 2)
	}
	if recordAuthenticationCallCount != 1 {
		t.Fatalf("RecordIdentityAuthentication() call count got %d want %d", recordAuthenticationCallCount, 1)
	}
}

func TestCognitoSessionManagerRefreshRecentAuthentication(t *testing.T) {
	t.Parallel()

	now := time.Unix(1711000300, 0).UTC()

	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{
		getIdentityByProviderAndSubject: func(context.Context, string, string) (Identity, error) {
			t.Fatal("GetIdentityByProviderAndSubject() should not be called")
			return Identity{}, nil
		},
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			t.Fatal("GetIdentityByEmail() should not be called")
			return Identity{}, nil
		},
		createIdentity: func(context.Context, CreateIdentityInput) (Identity, error) {
			t.Fatal("CreateIdentity() should not be called")
			return Identity{}, nil
		},
		createSession: func(context.Context, CreateSessionInput) (SessionRecord, error) {
			t.Fatal("CreateSession() should not be called")
			return SessionRecord{}, nil
		},
		createUserWithIdentityAndSession: func(context.Context, CreateUserWithIdentityAndSessionInput) (SessionRecord, error) {
			t.Fatal("CreateUserWithIdentityAndSession() should not be called")
			return SessionRecord{}, nil
		},
		recordIdentityAuthentication: func(context.Context, RecordIdentityAuthenticationInput) (Identity, error) {
			t.Fatal("RecordIdentityAuthentication() should not be called")
			return Identity{}, nil
		},
		refreshSessionRecentAuthenticatedAtByHash: func(_ context.Context, sessionTokenHash string, recentAuthenticatedAt time.Time) (SessionRecord, error) {
			if sessionTokenHash != HashSessionToken("raw-session-token") {
				t.Fatalf("RefreshSessionRecentAuthenticatedAtByTokenHash() token got %q want %q", sessionTokenHash, HashSessionToken("raw-session-token"))
			}
			if recentAuthenticatedAt != now {
				t.Fatalf("RefreshSessionRecentAuthenticatedAtByTokenHash() time got %s want %s", recentAuthenticatedAt, now)
			}

			return SessionRecord{}, nil
		},
	})

	if err := manager.RefreshRecentAuthentication(context.Background(), "  raw-session-token  ", now); err != nil {
		t.Fatalf("RefreshRecentAuthentication() error = %v, want nil", err)
	}
}

func TestCognitoSessionManagerRefreshRecentAuthenticationUsesNowWhenZero(t *testing.T) {
	t.Parallel()

	now := time.Unix(1711000400, 0).UTC()

	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{
		getIdentityByProviderAndSubject: func(context.Context, string, string) (Identity, error) {
			t.Fatal("GetIdentityByProviderAndSubject() should not be called")
			return Identity{}, nil
		},
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			t.Fatal("GetIdentityByEmail() should not be called")
			return Identity{}, nil
		},
		createIdentity: func(context.Context, CreateIdentityInput) (Identity, error) {
			t.Fatal("CreateIdentity() should not be called")
			return Identity{}, nil
		},
		createSession: func(context.Context, CreateSessionInput) (SessionRecord, error) {
			t.Fatal("CreateSession() should not be called")
			return SessionRecord{}, nil
		},
		createUserWithIdentityAndSession: func(context.Context, CreateUserWithIdentityAndSessionInput) (SessionRecord, error) {
			t.Fatal("CreateUserWithIdentityAndSession() should not be called")
			return SessionRecord{}, nil
		},
		recordIdentityAuthentication: func(context.Context, RecordIdentityAuthenticationInput) (Identity, error) {
			t.Fatal("RecordIdentityAuthentication() should not be called")
			return Identity{}, nil
		},
		refreshSessionRecentAuthenticatedAtByHash: func(_ context.Context, sessionTokenHash string, recentAuthenticatedAt time.Time) (SessionRecord, error) {
			if sessionTokenHash != HashSessionToken("raw-session-token") {
				t.Fatalf("RefreshSessionRecentAuthenticatedAtByTokenHash() token got %q want %q", sessionTokenHash, HashSessionToken("raw-session-token"))
			}
			if recentAuthenticatedAt != now {
				t.Fatalf("RefreshSessionRecentAuthenticatedAtByTokenHash() time got %s want %s", recentAuthenticatedAt, now)
			}

			return SessionRecord{}, nil
		},
	})
	manager.now = func() time.Time { return now }

	if err := manager.RefreshRecentAuthentication(context.Background(), "raw-session-token", time.Time{}); err != nil {
		t.Fatalf("RefreshRecentAuthentication() error = %v, want nil", err)
	}
}

func TestCognitoSessionManagerRefreshRecentAuthenticationRejectsBlankToken(t *testing.T) {
	t.Parallel()

	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{})

	if err := manager.RefreshRecentAuthentication(context.Background(), "   ", time.Now().UTC()); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("RefreshRecentAuthentication() error got %v want %v", err, ErrSessionNotFound)
	}
}

func TestCognitoSessionManagerRefreshRecentAuthenticationWrapsRepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("update failed")

	manager := NewCognitoSessionManager(cognitoSessionRepositoryStub{
		getIdentityByProviderAndSubject: func(context.Context, string, string) (Identity, error) {
			t.Fatal("GetIdentityByProviderAndSubject() should not be called")
			return Identity{}, nil
		},
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			t.Fatal("GetIdentityByEmail() should not be called")
			return Identity{}, nil
		},
		createIdentity: func(context.Context, CreateIdentityInput) (Identity, error) {
			t.Fatal("CreateIdentity() should not be called")
			return Identity{}, nil
		},
		createSession: func(context.Context, CreateSessionInput) (SessionRecord, error) {
			t.Fatal("CreateSession() should not be called")
			return SessionRecord{}, nil
		},
		createUserWithIdentityAndSession: func(context.Context, CreateUserWithIdentityAndSessionInput) (SessionRecord, error) {
			t.Fatal("CreateUserWithIdentityAndSession() should not be called")
			return SessionRecord{}, nil
		},
		recordIdentityAuthentication: func(context.Context, RecordIdentityAuthenticationInput) (Identity, error) {
			t.Fatal("RecordIdentityAuthentication() should not be called")
			return Identity{}, nil
		},
		refreshSessionRecentAuthenticatedAtByHash: func(context.Context, string, time.Time) (SessionRecord, error) {
			return SessionRecord{}, expectedErr
		},
	})

	if err := manager.RefreshRecentAuthentication(context.Background(), "raw-session-token", time.Now().UTC()); !errors.Is(err, expectedErr) {
		t.Fatalf("RefreshRecentAuthentication() error got %v want %v", err, expectedErr)
	}
}
