package asc

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/auth"
)

func init() {
	// Seed the random number generator for jitter
	rand.Seed(time.Now().UnixNano())
}

const (
	// BaseURL is the App Store Connect API base URL
	BaseURL = "https://api.appstoreconnect.apple.com"
	// DefaultTimeout is the default request timeout
	DefaultTimeout = 30 * time.Second
	// DefaultUploadTimeout is the default timeout for upload operations.
	DefaultUploadTimeout = 60 * time.Second
	tokenLifetime        = 20 * time.Minute

	// Retry defaults
	DefaultMaxRetries = 3
	DefaultBaseDelay  = 1 * time.Second
	DefaultMaxDelay   = 30 * time.Second
)

// RetryableError is returned when a request can be retried (e.g., rate limiting).
type RetryableError struct {
	Err        error
	RetryAfter time.Duration
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// IsRetryable checks if an error indicates the request can be retried.
func IsRetryable(err error) bool {
	var re *RetryableError
	return errors.As(err, &re)
}

// GetRetryAfter extracts the retry-after duration from an error.
func GetRetryAfter(err error) time.Duration {
	var re *RetryableError
	if errors.As(err, &re) {
		return re.RetryAfter
	}
	return 0
}

// RetryOptions configures retry behavior.
//   - MaxRetries: Number of retry attempts. 0 = no retries (fail fast),
//     negative = use DefaultMaxRetries.
//   - BaseDelay: Initial delay between retries (with exponential backoff).
//   - MaxDelay: Maximum delay cap for backoff.
type RetryOptions struct {
	MaxRetries int           // 0=disabled, negative=default, positive=retry count
	BaseDelay  time.Duration // Initial delay for exponential backoff
	MaxDelay   time.Duration // Maximum delay cap
}

// ResolveRetryOptions returns retry options, optionally overridden by env vars.
func ResolveRetryOptions() RetryOptions {
	opts := RetryOptions{
		MaxRetries: DefaultMaxRetries,
		BaseDelay:  DefaultBaseDelay,
		MaxDelay:   DefaultMaxDelay,
	}

	if override := strings.TrimSpace(os.Getenv("ASC_MAX_RETRIES")); override != "" {
		if parsed, err := strconv.Atoi(override); err == nil && parsed >= 0 {
			opts.MaxRetries = parsed
		}
	}
	if override := strings.TrimSpace(os.Getenv("ASC_BASE_DELAY")); override != "" {
		if parsed, err := time.ParseDuration(override); err == nil && parsed > 0 {
			opts.BaseDelay = parsed
		}
	}
	if override := strings.TrimSpace(os.Getenv("ASC_MAX_DELAY")); override != "" {
		if parsed, err := time.ParseDuration(override); err == nil && parsed > 0 {
			opts.MaxDelay = parsed
		}
	}
	return opts
}

// WithRetry executes a function with retry logic for rate limiting.
// It uses exponential backoff with jitter and respects Retry-After headers.
func WithRetry[T any](ctx context.Context, fn func() (T, error), opts RetryOptions) (T, error) {
	var zero T

	// If MaxRetries is negative, use the default; if zero, fail on first error
	if opts.MaxRetries < 0 {
		opts.MaxRetries = DefaultMaxRetries
	}
	if opts.MaxRetries == 0 {
		return fn()
	}

	if opts.BaseDelay <= 0 {
		opts.BaseDelay = DefaultBaseDelay
	}
	if opts.MaxDelay <= 0 {
		opts.MaxDelay = DefaultMaxDelay
	}

	retryCount := 0

	for {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		// Check if error is retryable
		if !IsRetryable(err) {
			return zero, err
		}

		// Check if we've exceeded max retries
		if retryCount >= opts.MaxRetries {
			return zero, fmt.Errorf("retry limit exceeded after %d retries: %w", retryCount+1, err)
		}

		// Calculate delay
		delay := GetRetryAfter(err)
		if delay == 0 {
			// Exponential backoff with jitter, capped to prevent overflow
			expDelay := opts.BaseDelay
			if retryCount > 0 && retryCount < 31 { // Prevent overflow for reasonable retry counts
				expDelay = opts.BaseDelay * time.Duration(1<<retryCount)
			}
			if expDelay > opts.MaxDelay || expDelay <= 0 {
				expDelay = opts.MaxDelay
			}
			// Add jitter: ±25% of the delay
			jitter := float64(expDelay) * 0.25 * (2*rand.Float64() - 1)
			delay = expDelay + time.Duration(jitter)
			if delay < 0 {
				delay = expDelay / 2 // minimum delay
			}
		}

		if shouldLogRetries() {
			fmt.Fprintf(os.Stderr, "retrying in %s (attempt %d/%d): %v\n", delay, retryCount+1, opts.MaxRetries, err)
		}

		retryCount++

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return zero, fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(delay):
			// Continue to next retry
		}
	}
}

func shouldLogRetries() bool {
	return strings.TrimSpace(os.Getenv("ASC_RETRY_LOG")) != ""
}

// ResolveTimeout returns the request timeout, optionally overridden by env vars.
func ResolveTimeout() time.Duration {
	return ResolveTimeoutWithDefault(DefaultTimeout)
}

// ResolveUploadTimeout returns the upload timeout, optionally overridden by env vars.
func ResolveUploadTimeout() time.Duration {
	return resolveTimeoutWithDefaultAndEnv(DefaultUploadTimeout, "ASC_UPLOAD_TIMEOUT", "ASC_UPLOAD_TIMEOUT_SECONDS")
}

// ResolveTimeoutWithDefault returns the request timeout using a custom default.
// ASC_TIMEOUT and ASC_TIMEOUT_SECONDS override the default when set.
func ResolveTimeoutWithDefault(defaultTimeout time.Duration) time.Duration {
	return resolveTimeoutWithDefaultAndEnv(defaultTimeout, "ASC_TIMEOUT", "ASC_TIMEOUT_SECONDS")
}

func resolveTimeoutWithDefaultAndEnv(defaultTimeout time.Duration, durationEnv, secondsEnv string) time.Duration {
	timeout := defaultTimeout
	if override := strings.TrimSpace(os.Getenv(durationEnv)); override != "" {
		if parsed, err := time.ParseDuration(override); err == nil && parsed > 0 {
			timeout = parsed
		}
	} else if override := strings.TrimSpace(os.Getenv(secondsEnv)); override != "" {
		if parsed, err := time.ParseDuration(override + "s"); err == nil && parsed > 0 {
			timeout = parsed
		}
	}
	return timeout
}

// Client is an App Store Connect API client
type Client struct {
	httpClient *http.Client
	keyID      string
	issuerID   string
	privateKey *ecdsa.PrivateKey
}

// NewClient creates a new ASC client
func NewClient(keyID, issuerID, privateKeyPath string) (*Client, error) {
	if err := auth.ValidateKeyFile(privateKeyPath); err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	key, err := auth.LoadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: ResolveTimeout(),
		},
		keyID:      keyID,
		issuerID:   issuerID,
		privateKey: key,
	}, nil
}
