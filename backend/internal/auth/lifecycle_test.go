package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type lifecycleRepositoryStub struct {
	getIdentityByEmail                    func(context.Context, string) (Identity, error)
	createLoginChallenge                  func(context.Context, CreateLoginChallengeInput) (Challenge, error)
	getLatestPendingLoginChallengeByEmail func(context.Context, string) (Challenge, error)
	incrementLoginChallengeAttemptCount   func(context.Context, uuid.UUID) (Challenge, error)
	consumeLoginChallenge                 func(context.Context, uuid.UUID, time.Time) (Challenge, error)
	recordIdentityAuthentication          func(context.Context, RecordIdentityAuthenticationInput) (Identity, error)
	createSession                         func(context.Context, CreateSessionInput) (SessionRecord, error)
	createUserWithEmailIdentityAndSession func(context.Context, CreateUserWithEmailIdentityAndSessionInput) (SessionRecord, error)
	revokeActiveSessionByTokenHash        func(context.Context, string, time.Time) (SessionRecord, error)
}

func (s lifecycleRepositoryStub) GetIdentityByEmail(ctx context.Context, emailNormalized string) (Identity, error) {
	return s.getIdentityByEmail(ctx, emailNormalized)
}

func (s lifecycleRepositoryStub) CreateLoginChallenge(ctx context.Context, input CreateLoginChallengeInput) (Challenge, error) {
	return s.createLoginChallenge(ctx, input)
}

func (s lifecycleRepositoryStub) GetLatestPendingLoginChallengeByEmail(ctx context.Context, emailNormalized string) (Challenge, error) {
	return s.getLatestPendingLoginChallengeByEmail(ctx, emailNormalized)
}

func (s lifecycleRepositoryStub) IncrementLoginChallengeAttemptCount(ctx context.Context, id uuid.UUID) (Challenge, error) {
	return s.incrementLoginChallengeAttemptCount(ctx, id)
}

func (s lifecycleRepositoryStub) ConsumeLoginChallenge(ctx context.Context, id uuid.UUID, consumedAt time.Time) (Challenge, error) {
	return s.consumeLoginChallenge(ctx, id, consumedAt)
}

func (s lifecycleRepositoryStub) RecordIdentityAuthentication(ctx context.Context, input RecordIdentityAuthenticationInput) (Identity, error) {
	return s.recordIdentityAuthentication(ctx, input)
}

func (s lifecycleRepositoryStub) CreateSession(ctx context.Context, input CreateSessionInput) (SessionRecord, error) {
	return s.createSession(ctx, input)
}

func (s lifecycleRepositoryStub) CreateUserWithEmailIdentityAndSession(ctx context.Context, input CreateUserWithEmailIdentityAndSessionInput) (SessionRecord, error) {
	return s.createUserWithEmailIdentityAndSession(ctx, input)
}

func (s lifecycleRepositoryStub) RevokeActiveSessionByTokenHash(ctx context.Context, sessionTokenHash string, revokedAt time.Time) (SessionRecord, error) {
	return s.revokeActiveSessionByTokenHash(ctx, sessionTokenHash, revokedAt)
}

func TestIssueSignInChallengeReturnsEmailNotFound(t *testing.T) {
	t.Parallel()

	lifecycle := NewLifecycle(lifecycleRepositoryStub{
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			return Identity{}, ErrIdentityNotFound
		},
		createLoginChallenge: func(context.Context, CreateLoginChallengeInput) (Challenge, error) {
			t.Fatal("CreateLoginChallenge() should not be called")
			return Challenge{}, nil
		},
	})

	if _, err := lifecycle.IssueSignInChallenge(context.Background(), "fan@example.com"); !errors.Is(err, ErrEmailNotFound) {
		t.Fatalf("IssueSignInChallenge() error got %v want %v", err, ErrEmailNotFound)
	}
}

func TestNewLifecycleInitializesDefaultTokenGenerators(t *testing.T) {
	t.Parallel()

	lifecycle := NewLifecycle(lifecycleRepositoryStub{})

	challengeToken, err := lifecycle.newChallengeToken()
	if err != nil {
		t.Fatalf("newChallengeToken() error = %v, want nil", err)
	}
	if challengeToken == "" {
		t.Fatal("newChallengeToken() returned empty token")
	}

	sessionToken, err := lifecycle.newSessionToken()
	if err != nil {
		t.Fatalf("newSessionToken() error = %v, want nil", err)
	}
	if sessionToken == "" {
		t.Fatal("newSessionToken() returned empty token")
	}
}

func TestIssueSignUpChallengeReturnsEmailAlreadyRegistered(t *testing.T) {
	t.Parallel()

	lifecycle := NewLifecycle(lifecycleRepositoryStub{
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			return Identity{ID: uuid.New()}, nil
		},
		createLoginChallenge: func(context.Context, CreateLoginChallengeInput) (Challenge, error) {
			t.Fatal("CreateLoginChallenge() should not be called")
			return Challenge{}, nil
		},
	})

	if _, err := lifecycle.IssueSignUpChallenge(context.Background(), "fan@example.com"); !errors.Is(err, ErrEmailAlreadyRegistered) {
		t.Fatalf("IssueSignUpChallenge() error got %v want %v", err, ErrEmailAlreadyRegistered)
	}
}

func TestIssueSignInChallengeCreatesChallenge(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	lifecycle := NewLifecycle(lifecycleRepositoryStub{
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			return Identity{ID: uuid.New()}, nil
		},
		createLoginChallenge: func(_ context.Context, input CreateLoginChallengeInput) (Challenge, error) {
			if input.EmailNormalized != "fan@example.com" {
				t.Fatalf("CreateLoginChallenge() email got %q want %q", input.EmailNormalized, "fan@example.com")
			}
			if input.ChallengeTokenHash != HashChallengeToken("challenge-token") {
				t.Fatalf("CreateLoginChallenge() challenge hash got %q want %q", input.ChallengeTokenHash, HashChallengeToken("challenge-token"))
			}
			if !input.ExpiresAt.Equal(now.Add(defaultChallengeTTL)) {
				t.Fatalf("CreateLoginChallenge() expires_at got %s want %s", input.ExpiresAt, now.Add(defaultChallengeTTL))
			}

			return Challenge{ExpiresAt: input.ExpiresAt}, nil
		},
	})
	lifecycle.now = func() time.Time { return now }
	lifecycle.newChallengeToken = func() (string, error) { return "challenge-token", nil }

	got, err := lifecycle.IssueSignInChallenge(context.Background(), "fan@example.com")
	if err != nil {
		t.Fatalf("IssueSignInChallenge() error = %v, want nil", err)
	}
	if got.Token != "challenge-token" {
		t.Fatalf("IssueSignInChallenge() token got %q want %q", got.Token, "challenge-token")
	}
}

func TestIssueSignUpChallengeCreatesChallenge(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000001, 0).UTC()
	lifecycle := NewLifecycle(lifecycleRepositoryStub{
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			return Identity{}, ErrIdentityNotFound
		},
		createLoginChallenge: func(_ context.Context, input CreateLoginChallengeInput) (Challenge, error) {
			if input.EmailNormalized != "fan@example.com" {
				t.Fatalf("CreateLoginChallenge() email got %q want %q", input.EmailNormalized, "fan@example.com")
			}
			if input.ChallengeTokenHash != HashChallengeToken("challenge-token") {
				t.Fatalf("CreateLoginChallenge() challenge hash got %q want %q", input.ChallengeTokenHash, HashChallengeToken("challenge-token"))
			}

			return Challenge{ExpiresAt: input.ExpiresAt}, nil
		},
	})
	lifecycle.now = func() time.Time { return now }
	lifecycle.newChallengeToken = func() (string, error) { return "challenge-token", nil }

	got, err := lifecycle.IssueSignUpChallenge(context.Background(), "fan@example.com")
	if err != nil {
		t.Fatalf("IssueSignUpChallenge() error = %v, want nil", err)
	}
	if got.Token != "challenge-token" {
		t.Fatalf("IssueSignUpChallenge() token got %q want %q", got.Token, "challenge-token")
	}
}

func TestStartSignInSessionCreatesSession(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000100, 0).UTC()
	verifiedAt := now.Add(-time.Hour)
	identityID := uuid.New()
	userID := uuid.New()
	challengeID := uuid.New()

	lifecycle := NewLifecycle(lifecycleRepositoryStub{
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			return Identity{
				ID:         identityID,
				UserID:     userID,
				VerifiedAt: &verifiedAt,
			}, nil
		},
		getLatestPendingLoginChallengeByEmail: func(context.Context, string) (Challenge, error) {
			return Challenge{
				ID:                 challengeID,
				ChallengeTokenHash: HashChallengeToken("challenge-token"),
			}, nil
		},
		incrementLoginChallengeAttemptCount: func(context.Context, uuid.UUID) (Challenge, error) {
			t.Fatal("IncrementLoginChallengeAttemptCount() should not be called")
			return Challenge{}, nil
		},
		consumeLoginChallenge: func(_ context.Context, id uuid.UUID, consumedAt time.Time) (Challenge, error) {
			if id != challengeID {
				t.Fatalf("ConsumeLoginChallenge() id got %s want %s", id, challengeID)
			}
			if !consumedAt.Equal(now) {
				t.Fatalf("ConsumeLoginChallenge() consumedAt got %s want %s", consumedAt, now)
			}

			return Challenge{ID: id}, nil
		},
		recordIdentityAuthentication: func(_ context.Context, input RecordIdentityAuthenticationInput) (Identity, error) {
			if input.ID != identityID {
				t.Fatalf("RecordIdentityAuthentication() id got %s want %s", input.ID, identityID)
			}
			if input.VerifiedAt == nil || !input.VerifiedAt.Equal(verifiedAt) {
				t.Fatalf("RecordIdentityAuthentication() verified_at got %v want %s", input.VerifiedAt, verifiedAt)
			}

			return Identity{}, nil
		},
		createSession: func(_ context.Context, input CreateSessionInput) (SessionRecord, error) {
			if input.UserID != userID {
				t.Fatalf("CreateSession() user_id got %s want %s", input.UserID, userID)
			}
			if input.ActiveMode != ActiveModeFan {
				t.Fatalf("CreateSession() active_mode got %q want %q", input.ActiveMode, ActiveModeFan)
			}
			if input.SessionTokenHash != HashSessionToken("session-token") {
				t.Fatalf("CreateSession() token hash got %q want %q", input.SessionTokenHash, HashSessionToken("session-token"))
			}

			return SessionRecord{ExpiresAt: input.ExpiresAt}, nil
		},
	})
	lifecycle.now = func() time.Time { return now }
	lifecycle.newSessionToken = func() (string, error) { return "session-token", nil }

	got, err := lifecycle.StartSignInSession(context.Background(), "fan@example.com", "challenge-token")
	if err != nil {
		t.Fatalf("StartSignInSession() error = %v, want nil", err)
	}
	if got.Token != "session-token" {
		t.Fatalf("StartSignInSession() token got %q want %q", got.Token, "session-token")
	}
	if !got.ExpiresAt.Equal(now.Add(defaultSessionTTL)) {
		t.Fatalf("StartSignInSession() expires_at got %s want %s", got.ExpiresAt, now.Add(defaultSessionTTL))
	}
}

func TestStartSignInSessionRejectsInvalidChallenge(t *testing.T) {
	t.Parallel()

	challengeID := uuid.New()
	incremented := false
	lifecycle := NewLifecycle(lifecycleRepositoryStub{
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			return Identity{ID: uuid.New(), UserID: uuid.New()}, nil
		},
		getLatestPendingLoginChallengeByEmail: func(context.Context, string) (Challenge, error) {
			return Challenge{
				ID:                 challengeID,
				ChallengeTokenHash: HashChallengeToken("expected-token"),
			}, nil
		},
		incrementLoginChallengeAttemptCount: func(context.Context, uuid.UUID) (Challenge, error) {
			incremented = true
			return Challenge{}, nil
		},
		consumeLoginChallenge: func(context.Context, uuid.UUID, time.Time) (Challenge, error) {
			t.Fatal("ConsumeLoginChallenge() should not be called")
			return Challenge{}, nil
		},
		recordIdentityAuthentication: func(context.Context, RecordIdentityAuthenticationInput) (Identity, error) {
			t.Fatal("RecordIdentityAuthentication() should not be called")
			return Identity{}, nil
		},
		createSession: func(context.Context, CreateSessionInput) (SessionRecord, error) {
			t.Fatal("CreateSession() should not be called")
			return SessionRecord{}, nil
		},
	})

	if _, err := lifecycle.StartSignInSession(context.Background(), "fan@example.com", "wrong-token"); !errors.Is(err, ErrInvalidChallenge) {
		t.Fatalf("StartSignInSession() error got %v want %v", err, ErrInvalidChallenge)
	}
	if !incremented {
		t.Fatal("StartSignInSession() did not increment challenge attempt count")
	}
}

func TestStartSignInSessionRejectsBlankChallenge(t *testing.T) {
	t.Parallel()

	lifecycle := NewLifecycle(lifecycleRepositoryStub{
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			return Identity{ID: uuid.New(), UserID: uuid.New()}, nil
		},
		getLatestPendingLoginChallengeByEmail: func(context.Context, string) (Challenge, error) {
			t.Fatal("GetLatestPendingLoginChallengeByEmail() should not be called")
			return Challenge{}, nil
		},
	})

	if _, err := lifecycle.StartSignInSession(context.Background(), "fan@example.com", "   "); !errors.Is(err, ErrInvalidChallenge) {
		t.Fatalf("StartSignInSession() error got %v want %v", err, ErrInvalidChallenge)
	}
}

func TestStartSignUpSessionCreatesUserIdentityAndSession(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000200, 0).UTC()
	challengeID := uuid.New()
	lifecycle := NewLifecycle(lifecycleRepositoryStub{
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			return Identity{}, ErrIdentityNotFound
		},
		getLatestPendingLoginChallengeByEmail: func(context.Context, string) (Challenge, error) {
			return Challenge{
				ID:                 challengeID,
				ChallengeTokenHash: HashChallengeToken("sign-up-token"),
			}, nil
		},
		incrementLoginChallengeAttemptCount: func(context.Context, uuid.UUID) (Challenge, error) {
			t.Fatal("IncrementLoginChallengeAttemptCount() should not be called")
			return Challenge{}, nil
		},
		consumeLoginChallenge: func(context.Context, uuid.UUID, time.Time) (Challenge, error) {
			return Challenge{}, nil
		},
		createUserWithEmailIdentityAndSession: func(_ context.Context, input CreateUserWithEmailIdentityAndSessionInput) (SessionRecord, error) {
			if input.EmailNormalized != "fan@example.com" {
				t.Fatalf("CreateUserWithEmailIdentityAndSession() email got %q want %q", input.EmailNormalized, "fan@example.com")
			}
			if input.SessionTokenHash != HashSessionToken("session-token") {
				t.Fatalf("CreateUserWithEmailIdentityAndSession() token hash got %q want %q", input.SessionTokenHash, HashSessionToken("session-token"))
			}

			return SessionRecord{ExpiresAt: input.ExpiresAt}, nil
		},
	})
	lifecycle.now = func() time.Time { return now }
	lifecycle.newSessionToken = func() (string, error) { return "session-token", nil }

	got, err := lifecycle.StartSignUpSession(context.Background(), "fan@example.com", "sign-up-token")
	if err != nil {
		t.Fatalf("StartSignUpSession() error = %v, want nil", err)
	}
	if got.Token != "session-token" {
		t.Fatalf("StartSignUpSession() token got %q want %q", got.Token, "session-token")
	}
}

func TestStartSignUpSessionMapsExistingIdentityToConflict(t *testing.T) {
	t.Parallel()

	lifecycle := NewLifecycle(lifecycleRepositoryStub{
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			return Identity{ID: uuid.New()}, nil
		},
	})

	if _, err := lifecycle.StartSignUpSession(context.Background(), "fan@example.com", "challenge-token"); !errors.Is(err, ErrEmailAlreadyRegistered) {
		t.Fatalf("StartSignUpSession() error got %v want %v", err, ErrEmailAlreadyRegistered)
	}
}

func TestStartSignUpSessionMapsIdentityWriteConflict(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000201, 0).UTC()
	challengeID := uuid.New()
	lifecycle := NewLifecycle(lifecycleRepositoryStub{
		getIdentityByEmail: func(context.Context, string) (Identity, error) {
			return Identity{}, ErrIdentityNotFound
		},
		getLatestPendingLoginChallengeByEmail: func(context.Context, string) (Challenge, error) {
			return Challenge{
				ID:                 challengeID,
				ChallengeTokenHash: HashChallengeToken("sign-up-token"),
			}, nil
		},
		consumeLoginChallenge: func(context.Context, uuid.UUID, time.Time) (Challenge, error) {
			return Challenge{}, nil
		},
		createUserWithEmailIdentityAndSession: func(context.Context, CreateUserWithEmailIdentityAndSessionInput) (SessionRecord, error) {
			return SessionRecord{}, ErrIdentityAlreadyExists
		},
	})
	lifecycle.now = func() time.Time { return now }
	lifecycle.newSessionToken = func() (string, error) { return "session-token", nil }

	if _, err := lifecycle.StartSignUpSession(context.Background(), "fan@example.com", "sign-up-token"); !errors.Is(err, ErrEmailAlreadyRegistered) {
		t.Fatalf("StartSignUpSession() error got %v want %v", err, ErrEmailAlreadyRegistered)
	}
}

func TestLogoutIgnoresMissingSession(t *testing.T) {
	t.Parallel()

	lifecycle := NewLifecycle(lifecycleRepositoryStub{
		revokeActiveSessionByTokenHash: func(context.Context, string, time.Time) (SessionRecord, error) {
			return SessionRecord{}, ErrSessionNotFound
		},
	})

	if err := lifecycle.Logout(context.Background(), "raw-session-token"); err != nil {
		t.Fatalf("Logout() error = %v, want nil", err)
	}
}

func TestLogoutIgnoresBlankToken(t *testing.T) {
	t.Parallel()

	lifecycle := NewLifecycle(lifecycleRepositoryStub{
		revokeActiveSessionByTokenHash: func(context.Context, string, time.Time) (SessionRecord, error) {
			t.Fatal("RevokeActiveSessionByTokenHash() should not be called")
			return SessionRecord{}, nil
		},
	})

	if err := lifecycle.Logout(context.Background(), "   "); err != nil {
		t.Fatalf("Logout() error = %v, want nil", err)
	}
}

func TestLogoutWrapsRepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("revoke failed")
	lifecycle := NewLifecycle(lifecycleRepositoryStub{
		revokeActiveSessionByTokenHash: func(context.Context, string, time.Time) (SessionRecord, error) {
			return SessionRecord{}, expectedErr
		},
	})

	if err := lifecycle.Logout(context.Background(), "raw-session-token"); !errors.Is(err, expectedErr) {
		t.Fatalf("Logout() error got %v want wrapped %v", err, expectedErr)
	}
}
