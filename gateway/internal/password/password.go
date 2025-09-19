package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/vctrl/currency-service/gateway/internal/config"
	"golang.org/x/crypto/argon2"
)

const (
	time    = uint32(1)
	memory  = uint32(64 * 1024)
	threads = uint8(4)
	keyLen  = uint32(32)
)

type ProtectedPassword string

func (p ProtectedPassword) String() string {
	return string(p)
}

type ProtectedPasswordManager interface {
	CreatePassword(planTextPassword string) (ProtectedPassword, error)
	VerifyPassword(planTextPassword string, protectedPassword ProtectedPassword) error
}

type UserPasswordManager struct {
	passwordPolicy       config.PasswordPolicy
	requiredSymbolsCheck *regexp.Regexp
}

func NewUserPasswordManager(passwordPolicy config.PasswordPolicy) (*UserPasswordManager, error) {
	passwordManager := UserPasswordManager{passwordPolicy: passwordPolicy}
	if passwordPolicy.RequiredSymbolsRegExp != "" {
		requiredSymbolsCheck, err := regexp.Compile(passwordPolicy.RequiredSymbolsRegExp)
		if err != nil {
			return nil, err
		}
		passwordManager.requiredSymbolsCheck = requiredSymbolsCheck
	}

	return &passwordManager, nil
}

func (m *UserPasswordManager) CreatePassword(planTextPassword string) (ProtectedPassword, error) {
	if err := m.validate(planTextPassword); err != nil {
		return "", err
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(planTextPassword+m.passwordPolicy.GlobalSalt), salt, time, memory, threads, keyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	protected := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memory, time, threads, b64Salt, b64Hash)

	return ProtectedPassword(protected), nil
}

func (m *UserPasswordManager) VerifyPassword(planTextPassword string, protectedPassword ProtectedPassword) error {
	parts := strings.Split(protectedPassword.String(), "$")
	if len(parts) != 6 {
		return ErrProtectedPasswordFormat
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return err
	}

	newHash := argon2.IDKey([]byte(planTextPassword+m.passwordPolicy.GlobalSalt), salt, time, memory, threads, keyLen)
	if subtle.ConstantTimeCompare(hash, newHash) == 1 {
		return nil
	}

	return ErrInvalidPassword
}

func (m *UserPasswordManager) validate(planTextPassword string) error {
	if m.passwordPolicy.MinSize != 0 && len(planTextPassword) < m.passwordPolicy.MinSize {
		return ErrShortPassword
	}

	if m.passwordPolicy.MaxSize != 0 && len(planTextPassword) > m.passwordPolicy.MaxSize {
		return ErrLongPassword
	}

	if m.requiredSymbolsCheck == nil {
		return nil
	}

	if m.requiredSymbolsCheck.MatchString(planTextPassword) {
		return nil
	} else {
		return NewErrRequiredSymbols(m.passwordPolicy.RequiredSymbolsHint)
	}
}
