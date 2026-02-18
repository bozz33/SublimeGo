package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
)

// MFAConfig configures Multi-Factor Authentication behaviour.
type MFAConfig struct {
	// Issuer is the service name shown in authenticator apps (e.g. "SublimeGo Admin").
	Issuer string
	// Digits is the OTP length (default 6).
	Digits int
	// Period is the TOTP time step in seconds (default 30).
	Period int
	// Skew is the number of periods to check before/after the current one (default 1).
	Skew int
	// SecretLength is the byte length of generated secrets (default 20).
	SecretLength int
	// RecoveryCodeCount is the number of recovery codes to generate (default 8).
	RecoveryCodeCount int
}

// DefaultMFAConfig returns sensible defaults for TOTP-based MFA.
func DefaultMFAConfig() *MFAConfig {
	return &MFAConfig{
		Issuer:            "SublimeGo",
		Digits:            6,
		Period:            30,
		Skew:              1,
		SecretLength:      20,
		RecoveryCodeCount: 8,
	}
}

// MFA provides TOTP-based multi-factor authentication.
type MFA struct {
	cfg *MFAConfig
}

// NewMFA creates an MFA instance with the given config.
// Pass nil to use DefaultMFAConfig.
func NewMFA(cfg *MFAConfig) *MFA {
	if cfg == nil {
		cfg = DefaultMFAConfig()
	}
	if cfg.Digits == 0 {
		cfg.Digits = 6
	}
	if cfg.Period == 0 {
		cfg.Period = 30
	}
	if cfg.Skew == 0 {
		cfg.Skew = 1
	}
	if cfg.SecretLength == 0 {
		cfg.SecretLength = 20
	}
	if cfg.RecoveryCodeCount == 0 {
		cfg.RecoveryCodeCount = 8
	}
	return &MFA{cfg: cfg}
}

// GenerateSecret creates a new random TOTP secret encoded in base32.
func (m *MFA) GenerateSecret() (string, error) {
	secret := make([]byte, m.cfg.SecretLength)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("mfa: generate secret: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret), nil
}

// ProvisioningURI returns the otpauth:// URI for QR code generation.
// Authenticator apps (Google Authenticator, Authy, etc.) scan this URI.
func (m *MFA) ProvisioningURI(secret, accountName string) string {
	return fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s&digits=%d&period=%d",
		m.cfg.Issuer, accountName, secret, m.cfg.Issuer, m.cfg.Digits, m.cfg.Period,
	)
}

// Validate checks if the provided OTP code is valid for the given secret.
// It checks the current time step ± Skew periods to account for clock drift.
func (m *MFA) Validate(secret, code string) bool {
	if code == "" || secret == "" {
		return false
	}

	now := time.Now().Unix()
	counter := now / int64(m.cfg.Period)

	for i := -m.cfg.Skew; i <= m.cfg.Skew; i++ {
		expected := m.generateTOTP(secret, counter+int64(i))
		if hmac.Equal([]byte(expected), []byte(code)) {
			return true
		}
	}
	return false
}

// GenerateRecoveryCodes creates a set of single-use recovery codes.
// Each code is a 10-character alphanumeric string formatted as XXXXX-XXXXX.
func (m *MFA) GenerateRecoveryCodes() ([]string, error) {
	codes := make([]string, m.cfg.RecoveryCodeCount)
	for i := range codes {
		b := make([]byte, 5)
		if _, err := rand.Read(b); err != nil {
			return nil, fmt.Errorf("mfa: generate recovery code: %w", err)
		}
		raw := strings.ToUpper(fmt.Sprintf("%x", b))
		if len(raw) < 10 {
			raw = raw + strings.Repeat("0", 10-len(raw))
		}
		codes[i] = raw[:5] + "-" + raw[5:10]
	}
	return codes, nil
}

// generateTOTP computes the TOTP value for a given counter using HMAC-SHA1.
func (m *MFA) generateTOTP(secret string, counter int64) string {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(
		strings.ToUpper(strings.TrimSpace(secret)),
	)
	if err != nil {
		return ""
	}

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	hash := mac.Sum(nil)

	offset := hash[len(hash)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

	otp := truncated % uint32(math.Pow10(m.cfg.Digits))
	return fmt.Sprintf("%0*d", m.cfg.Digits, otp)
}

// IsMFAEnabled checks if the user has MFA enabled (secret stored in metadata).
func IsMFAEnabled(user *User) bool {
	if user == nil || user.Metadata == nil {
		return false
	}
	secret, ok := user.Metadata["mfa_secret"].(string)
	return ok && secret != ""
}

// SetMFASecret stores the TOTP secret in the user's metadata.
func SetMFASecret(user *User, secret string) {
	if user.Metadata == nil {
		user.Metadata = make(map[string]any)
	}
	user.Metadata["mfa_secret"] = secret
	user.Metadata["mfa_enabled_at"] = time.Now()
}

// GetMFASecret retrieves the TOTP secret from the user's metadata.
func GetMFASecret(user *User) string {
	if user == nil || user.Metadata == nil {
		return ""
	}
	if s, ok := user.Metadata["mfa_secret"].(string); ok {
		return s
	}
	return ""
}

// DisableMFA removes MFA data from the user's metadata.
func DisableMFA(user *User) {
	if user.Metadata == nil {
		return
	}
	delete(user.Metadata, "mfa_secret")
	delete(user.Metadata, "mfa_enabled_at")
	delete(user.Metadata, "mfa_recovery_codes")
}

// SetRecoveryCodes stores recovery codes in the user's metadata.
func SetRecoveryCodes(user *User, codes []string) {
	if user.Metadata == nil {
		user.Metadata = make(map[string]any)
	}
	user.Metadata["mfa_recovery_codes"] = codes
}

// UseRecoveryCode validates and consumes a recovery code.
// Returns true if the code was valid and has been removed.
func UseRecoveryCode(user *User, code string) bool {
	if user.Metadata == nil {
		return false
	}
	raw, ok := user.Metadata["mfa_recovery_codes"]
	if !ok {
		return false
	}
	codes, ok := raw.([]string)
	if !ok {
		return false
	}

	normalised := strings.ToUpper(strings.TrimSpace(code))
	for i, c := range codes {
		if c == normalised {
			// Remove the used code
			codes = append(codes[:i], codes[i+1:]...)
			user.Metadata["mfa_recovery_codes"] = codes
			return true
		}
	}
	return false
}
