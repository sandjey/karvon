package otpstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrOTPCooldown          = errors.New("otp cooldown")
	ErrOTPExpired           = errors.New("otp expired")
	ErrOTPInvalid           = errors.New("otp invalid")
	ErrOTPMaxAttempts       = errors.New("otp max attempts exceeded")
	ErrOTPRateLimited       = errors.New("otp rate limited")
	ErrOTPVerifyRateLimited = errors.New("otp verify attempts per phone exceeded")
)

// OTPCooldownError means a resend was requested before the cooldown elapsed.
type OTPCooldownError struct{ RetryAfter time.Duration }

func (e *OTPCooldownError) Error() string        { return "otp cooldown" }
func (e *OTPCooldownError) Is(target error) bool { return target == ErrOTPCooldown }

// OTPRateLimitError means the per-phone send limit for the window was hit.
type OTPRateLimitError struct{ RetryAfter time.Duration }

func (e *OTPRateLimitError) Error() string        { return "otp rate limited" }
func (e *OTPRateLimitError) Is(target error) bool { return target == ErrOTPRateLimited }

// OTPConfig configures an OTPStore.
type OTPConfig struct {
	Secret string
	TTL    time.Duration // OTP code TTL (e.g. 5m)

	// Cooldown is the base resend cooldown. After EscalateAfter sends within
	// AttemptWindow, CooldownLong is used instead. Set EscalateAfter=0 to disable.
	Cooldown      time.Duration
	CooldownLong  time.Duration
	EscalateAfter int
	AttemptWindow time.Duration

	MaxAttempts int

	SendLimitPerPhone int64
	SendWindow        time.Duration

	VerifyAttemptsPerPhone int64
	VerifyAttemptsWindow   time.Duration

	// UniversalOTP bypasses verification for QA/testing.
	UniversalOTP string

	Prefix string
}

type OTPStore struct {
	rdb                    *redis.Client
	secret                 string
	ttl                    time.Duration
	cooldown               time.Duration
	cooldownLong           time.Duration
	escalateAfter          int
	attemptWindow          time.Duration
	maxAttempts            int
	sendLimitPerPhone      int64
	sendWindow             time.Duration
	verifyAttemptsPerPhone int64
	verifyAttemptsWindow   time.Duration
	universalOTP           string
	prefix                 string
}

func NewOTPStore(rdb *redis.Client, cfg OTPConfig) *OTPStore {
	if cfg.SendWindow <= 0 {
		cfg.SendWindow = time.Hour
	}
	if cfg.AttemptWindow <= 0 {
		cfg.AttemptWindow = time.Hour
	}
	if cfg.TTL <= 0 {
		cfg.TTL = 5 * time.Minute
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 5
	}
	return &OTPStore{
		rdb:                    rdb,
		secret:                 cfg.Secret,
		ttl:                    cfg.TTL,
		cooldown:               cfg.Cooldown,
		cooldownLong:           cfg.CooldownLong,
		escalateAfter:          cfg.EscalateAfter,
		attemptWindow:          cfg.AttemptWindow,
		maxAttempts:            cfg.MaxAttempts,
		sendLimitPerPhone:      cfg.SendLimitPerPhone,
		sendWindow:             cfg.SendWindow,
		verifyAttemptsPerPhone: cfg.VerifyAttemptsPerPhone,
		verifyAttemptsWindow:   cfg.VerifyAttemptsWindow,
		universalOTP:           cfg.UniversalOTP,
		prefix:                 cfg.Prefix,
	}
}

func (s *OTPStore) otpKey(phone string) string          { return s.prefix + "otp:" + phone }
func (s *OTPStore) cooldownKey(phone string) string     { return s.prefix + "otp:cooldown:" + phone }
func (s *OTPStore) sendCountKey(phone string) string    { return s.prefix + "otp:send_count:" + phone }
func (s *OTPStore) resendAttemptsKey(phone string) string {
	return s.prefix + "otp:resend_attempts:" + phone
}
func (s *OTPStore) verifyAttemptsKey(phone string) string {
	return s.prefix + "otp:verify_attempts:" + phone
}

func (s *OTPStore) hash(phone, code string) string {
	sum := sha256.Sum256([]byte(phone + ":" + code + ":" + s.secret))
	return hex.EncodeToString(sum[:])
}

type OTPRecord struct {
	RequestID string
}

// CheckSendAllowed reports whether a new OTP may be sent to phone right now
// WITHOUT mutating any counter. Returns *OTPCooldownError or *OTPRateLimitError.
func (s *OTPStore) CheckSendAllowed(ctx context.Context, phone string) error {
	if s.cooldown > 0 {
		if ttl, err := s.rdb.TTL(ctx, s.cooldownKey(phone)).Result(); err == nil && ttl > 0 {
			return &OTPCooldownError{RetryAfter: ttl}
		}
	}
	if s.sendLimitPerPhone > 0 {
		if err := s.checkLimit(ctx, s.sendCountKey(phone), s.sendLimitPerPhone); err != nil {
			return err
		}
	}
	return nil
}

// SaveOTP stores the OTP hash with TTL and arms the resend cooldown.
// On active cooldown returns *OTPCooldownError; on hit send-limit returns *OTPRateLimitError.
// Returns the cooldown duration that was set so callers can surface the resend wait time.
func (s *OTPStore) SaveOTP(ctx context.Context, phone, code, requestID string) (cooldown time.Duration, err error) {
	// Cooldown check (safety net — CheckSendAllowed should be called first).
	if s.cooldown > 0 {
		if ttl, err := s.rdb.TTL(ctx, s.cooldownKey(phone)).Result(); err == nil && ttl > 0 {
			return 0, &OTPCooldownError{RetryAfter: ttl}
		}
	}

	// Per-phone send rate limit.
	if s.sendLimitPerPhone > 0 {
		if err := s.incrWithLimit(ctx, s.sendCountKey(phone), s.sendLimitPerPhone, s.sendWindow); err != nil {
			return 0, err
		}
	}

	key := s.otpKey(phone)

	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return 0, fmt.Errorf("redis del otp key: %w", err)
	}
	if _, err := s.rdb.HMSet(ctx, key, map[string]interface{}{
		"hash":       s.hash(phone, code),
		"attempts":   "0",
		"request_id": requestID,
	}).Result(); err != nil {
		return 0, fmt.Errorf("redis hmset otp key: %w", err)
	}
	if err := s.rdb.Expire(ctx, key, s.ttl).Err(); err != nil {
		return 0, fmt.Errorf("redis expire otp key: %w", err)
	}

	// Arm the (possibly escalated) resend cooldown.
	if s.cooldown > 0 {
		n, incErr := s.rdb.Incr(ctx, s.resendAttemptsKey(phone)).Result()
		if incErr == nil && n == 1 {
			_ = s.rdb.Expire(ctx, s.resendAttemptsKey(phone), s.attemptWindow).Err()
		}
		dur := s.cooldownFor(n)
		if err := s.rdb.Set(ctx, s.cooldownKey(phone), "1", dur).Err(); err != nil {
			return 0, fmt.Errorf("redis set cooldown key: %w", err)
		}
		return dur, nil
	}

	return 0, nil
}

// Verify checks the OTP code. On success it clears the OTP and resets cooldown
// state so the phone starts fresh on the next sign-in.
func (s *OTPStore) Verify(ctx context.Context, phone, code string) (OTPRecord, error) {
	if s.universalOTP != "" && code == s.universalOTP {
		return OTPRecord{RequestID: "universal-otp"}, nil
	}

	key := s.otpKey(phone)
	vals, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return OTPRecord{}, err
	}
	if len(vals) == 0 {
		return OTPRecord{}, ErrOTPExpired
	}

	attempts, _ := strconv.Atoi(vals["attempts"])
	if attempts >= s.maxAttempts {
		return OTPRecord{}, ErrOTPMaxAttempts
	}

	want := vals["hash"]
	got := s.hash(phone, code)
	if want == "" || got != want {
		// Per-phone verify rate limit (brute-force protection across code rotations).
		if s.verifyAttemptsPerPhone > 0 && s.verifyAttemptsWindow > 0 {
			vkey := s.verifyAttemptsKey(phone)
			n, err := s.rdb.Incr(ctx, vkey).Result()
			if err != nil {
				return OTPRecord{}, err
			}
			if n == 1 {
				_ = s.rdb.Expire(ctx, vkey, s.verifyAttemptsWindow).Err()
			}
			if n > s.verifyAttemptsPerPhone {
				return OTPRecord{}, ErrOTPVerifyRateLimited
			}
		}
		attempts++
		_ = s.rdb.HSet(ctx, key, "attempts", fmt.Sprintf("%d", attempts)).Err()
		if attempts >= s.maxAttempts {
			return OTPRecord{}, ErrOTPMaxAttempts
		}
		return OTPRecord{}, ErrOTPInvalid
	}

	// Success: clear OTP record and reset resend state.
	_ = s.rdb.Del(ctx, key).Err()
	_ = s.rdb.Del(ctx, s.cooldownKey(phone)).Err()
	_ = s.rdb.Del(ctx, s.resendAttemptsKey(phone)).Err()

	return OTPRecord{RequestID: vals["request_id"]}, nil
}

func (s *OTPStore) incrWithLimit(ctx context.Context, key string, limit int64, window time.Duration) error {
	n, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if n == 1 {
		_ = s.rdb.Expire(ctx, key, window).Err()
	}
	if n > limit {
		retry := window
		if ttl, terr := s.rdb.TTL(ctx, key).Result(); terr == nil && ttl > 0 {
			retry = ttl
		}
		return &OTPRateLimitError{RetryAfter: retry}
	}
	return nil
}

func (s *OTPStore) checkLimit(ctx context.Context, key string, limit int64) error {
	n, err := s.rdb.Get(ctx, key).Int64()
	if err != nil {
		return nil // fail open on Redis read error; SaveOTP enforces authoritatively
	}
	if n >= limit {
		retry := s.sendWindow
		if ttl, terr := s.rdb.TTL(ctx, key).Result(); terr == nil && ttl > 0 {
			retry = ttl
		}
		return &OTPRateLimitError{RetryAfter: retry}
	}
	return nil
}

// cooldownFor returns the resend cooldown for the n-th send within AttemptWindow.
// Cyclic escalation: EscalateAfter short cooldowns, then one CooldownLong, repeating.
func (s *OTPStore) cooldownFor(n int64) time.Duration {
	if s.escalateAfter > 0 && s.cooldownLong > 0 {
		period := int64(s.escalateAfter) + 1
		if n > 0 && n%period == 0 {
			return s.cooldownLong
		}
	}
	return s.cooldown
}
