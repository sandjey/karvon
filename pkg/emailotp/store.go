package emailotp

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrInvalid = errors.New("email otp invalid")
var ErrExpired = errors.New("email otp expired")

const ttlOTP      = 10 * time.Minute
const ttlVerified = 1 * time.Hour
const prefix      = "ctm:emailotp:"

type Store struct{ rdb *redis.Client }

func NewStore(rdb *redis.Client) *Store { return &Store{rdb: rdb} }

func GenerateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil { return "", err }
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (s *Store) Save(ctx context.Context, email, code string) error {
	return s.rdb.Set(ctx, prefix+"code:"+email, code, ttlOTP).Err()
}

func (s *Store) Verify(ctx context.Context, email, code string) error {
	stored, err := s.rdb.Get(ctx, prefix+"code:"+email).Result()
	if errors.Is(err, redis.Nil) { return ErrExpired }
	if err != nil { return err }
	if stored != code { return ErrInvalid }
	// Mark as verified for 1 hour
	s.rdb.Set(ctx, prefix+"verified:"+email, "1", ttlVerified)
	s.rdb.Del(ctx, prefix+"code:"+email)
	return nil
}

// IsVerified returns true if this email was recently verified via OTP.
func (s *Store) IsVerified(ctx context.Context, email string) bool {
	if email == "" { return false }
	v, _ := s.rdb.Get(ctx, prefix+"verified:"+email).Result()
	return v == "1"
}
