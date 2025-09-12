package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vctrl/currency-service/gateway/internal/config"
)

func TestUserPasswordManager_CreatePassword(t *testing.T) {
	tests := []struct {
		name          string
		password      string
		policy        config.PasswordPolicy
		expectedError bool
		errorMessage  string
	}{
		{
			name:     "valid password with minimum length",
			password: "123456",
			policy: config.PasswordPolicy{
				GlobalSalt:            "test-salt",
				MinSize:               6,
				MaxSize:               20,
				RequiredSymbolsRegExp: "",
				RequiredSymbolsHint:   "",
			},
			expectedError: false,
		},
		{
			name:     "valid password with maximum length",
			password: "12345678901234567890",
			policy: config.PasswordPolicy{
				GlobalSalt:            "test-salt",
				MinSize:               6,
				MaxSize:               20,
				RequiredSymbolsRegExp: "",
				RequiredSymbolsHint:   "",
			},
			expectedError: false,
		},
		{
			name:     "valid password with special characters",
			password: "Test123!",
			policy: config.PasswordPolicy{
				GlobalSalt:            "test-salt",
				MinSize:               6,
				MaxSize:               20,
				RequiredSymbolsRegExp: `[!@#]`,
				RequiredSymbolsHint:   "Password must contain !, @, or #",
			},
			expectedError: false,
		},
		{
			name:     "password too short",
			password: "12345",
			policy: config.PasswordPolicy{
				GlobalSalt:            "test-salt",
				MinSize:               6,
				MaxSize:               20,
				RequiredSymbolsRegExp: "",
				RequiredSymbolsHint:   "",
			},
			expectedError: true,
			errorMessage:  "short password",
		},
		{
			name:     "password too long",
			password: "123456789012345678901",
			policy: config.PasswordPolicy{
				GlobalSalt:            "test-salt",
				MinSize:               6,
				MaxSize:               20,
				RequiredSymbolsRegExp: "",
				RequiredSymbolsHint:   "",
			},
			expectedError: true,
			errorMessage:  "long password",
		},
		{
			name:     "password without required special characters",
			password: "Test123",
			policy: config.PasswordPolicy{
				GlobalSalt:            "test-salt",
				MinSize:               6,
				MaxSize:               20,
				RequiredSymbolsRegExp: `[!@#]`,
				RequiredSymbolsHint:   "Password must contain !, @, or #",
			},
			expectedError: true,
			errorMessage:  "Password must contain !, @, or #",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passwordManager := NewUserPasswordManager(tt.policy)

			result, err := passwordManager.CreatePassword(tt.password)

			if tt.expectedError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Empty(t, result)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, result)
				assert.IsType(t, ProtectedPassword(""), result)

				// Check Argon2 format
				resultStr := result.String()
				assert.Contains(t, resultStr, "$argon2id$")
				assert.Contains(t, resultStr, "v=19")
				assert.Contains(t, resultStr, "m=65536") // memory = 64 * 1024
				assert.Contains(t, resultStr, "t=1")     // time = 1
				assert.Contains(t, resultStr, "p=4")     // threads = 4
			}
		})
	}
}

func TestUserPasswordManager_VerifyPassword(t *testing.T) {
	policy := config.PasswordPolicy{
		GlobalSalt:            "test-salt",
		MinSize:               6,
		MaxSize:               20,
		RequiredSymbolsRegExp: "",
		RequiredSymbolsHint:   "",
	}

	passwordManager := NewUserPasswordManager(policy)

	tests := []struct {
		name             string
		originalPassword string
		verifyPassword   string
		expectedError    bool
	}{
		{
			name:             "correct password verification",
			originalPassword: "testpassword123",
			verifyPassword:   "testpassword123",
			expectedError:    false,
		},
		{
			name:             "incorrect password verification",
			originalPassword: "testpassword123",
			verifyPassword:   "wrongpassword",
			expectedError:    true,
		},
		{
			name:             "empty password verification",
			originalPassword: "testpassword123",
			verifyPassword:   "",
			expectedError:    true,
		},
		{
			name:             "case sensitive verification",
			originalPassword: "TestPassword123",
			verifyPassword:   "testpassword123",
			expectedError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create protected password
			protectedPassword, err := passwordManager.CreatePassword(tt.originalPassword)
			require.NoError(t, err)

			// Verify password
			err = passwordManager.VerifyPassword(tt.verifyPassword, protectedPassword)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUserPasswordManager_VerifyPassword_InvalidFormat(t *testing.T) {
	policy := config.PasswordPolicy{
		GlobalSalt:            "test-salt",
		MinSize:               6,
		MaxSize:               20,
		RequiredSymbolsRegExp: "",
		RequiredSymbolsHint:   "",
	}

	passwordManager := NewUserPasswordManager(policy)

	tests := []struct {
		name              string
		protectedPassword ProtectedPassword
		verifyPassword    string
	}{
		{
			name:              "invalid format - too few parts",
			protectedPassword: ProtectedPassword("$argon2id$v=19$m=65536,t=1,p=4$salt"),
			verifyPassword:    "testpassword",
		},
		{
			name:              "invalid format - too many parts",
			protectedPassword: ProtectedPassword("$argon2id$v=19$m=65536,t=1,p=4$salt$hash$extra"),
			verifyPassword:    "testpassword",
		},
		{
			name:              "invalid base64 salt",
			protectedPassword: ProtectedPassword("$argon2id$v=19$m=65536,t=1,p=4$invalid-salt$hash"),
			verifyPassword:    "testpassword",
		},
		{
			name:              "invalid base64 hash",
			protectedPassword: ProtectedPassword("$argon2id$v=19$m=65536,t=1,p=4$salt$invalid-hash"),
			verifyPassword:    "testpassword",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := passwordManager.VerifyPassword(tt.verifyPassword, tt.protectedPassword)
			assert.Error(t, err)
		})
	}
}

func TestUserPasswordManager_Validate(t *testing.T) {
	tests := []struct {
		name          string
		password      string
		policy        config.PasswordPolicy
		expectedError bool
		errorMessage  string
	}{
		{
			name:     "valid password - minimum length",
			password: "123456",
			policy: config.PasswordPolicy{
				MinSize:               6,
				MaxSize:               20,
				RequiredSymbolsRegExp: "",
				RequiredSymbolsHint:   "",
			},
			expectedError: false,
		},
		{
			name:     "valid password - maximum length",
			password: "12345678901234567890",
			policy: config.PasswordPolicy{
				MinSize:               6,
				MaxSize:               20,
				RequiredSymbolsRegExp: "",
				RequiredSymbolsHint:   "",
			},
			expectedError: false,
		},
		{
			name:     "valid password - with special characters",
			password: "Test123@",
			policy: config.PasswordPolicy{
				MinSize:               6,
				MaxSize:               20,
				RequiredSymbolsRegExp: `[!@#]`,
				RequiredSymbolsHint:   "Password must contain !, @, or #",
			},
			expectedError: false,
		},
		{
			name:     "password too short",
			password: "12345",
			policy: config.PasswordPolicy{
				MinSize:               6,
				MaxSize:               20,
				RequiredSymbolsRegExp: "",
				RequiredSymbolsHint:   "",
			},
			expectedError: true,
			errorMessage:  "short password",
		},
		{
			name:     "password too long",
			password: "123456789012345678901",
			policy: config.PasswordPolicy{
				MinSize:               6,
				MaxSize:               20,
				RequiredSymbolsRegExp: "",
				RequiredSymbolsHint:   "",
			},
			expectedError: true,
			errorMessage:  "long password",
		},
		{
			name:     "password without required special characters",
			password: "Test123",
			policy: config.PasswordPolicy{
				MinSize:               6,
				MaxSize:               20,
				RequiredSymbolsRegExp: `[!@#]`,
				RequiredSymbolsHint:   "Password must contain !, @, or #",
			},
			expectedError: true,
			errorMessage:  "Password must contain !, @, or #",
		},
		{
			name:     "no regexp validation when regexp is empty",
			password: "Test123",
			policy: config.PasswordPolicy{
				MinSize:               6,
				MaxSize:               20,
				RequiredSymbolsRegExp: "",
				RequiredSymbolsHint:   "Password must contain !, @, or #",
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passwordManager := NewUserPasswordManager(tt.policy)

			err := passwordManager.validate(tt.password)

			if tt.expectedError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewUserPasswordManager(t *testing.T) {
	tests := []struct {
		name   string
		policy config.PasswordPolicy
	}{
		{
			name: "policy with regexp",
			policy: config.PasswordPolicy{
				GlobalSalt:            "test-salt",
				MinSize:               6,
				MaxSize:               20,
				RequiredSymbolsRegExp: `[!@#]`,
				RequiredSymbolsHint:   "Password must contain !, @, or #",
			},
		},
		{
			name: "policy without regexp",
			policy: config.PasswordPolicy{
				GlobalSalt:            "test-salt",
				MinSize:               6,
				MaxSize:               20,
				RequiredSymbolsRegExp: "",
				RequiredSymbolsHint:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passwordManager := NewUserPasswordManager(tt.policy)

			assert.NotNil(t, passwordManager)
			assert.Equal(t, tt.policy, passwordManager.passwordPolicy)

			if tt.policy.RequiredSymbolsRegExp != "" {
				assert.NotNil(t, passwordManager.requiredSymbolsCheck)
			} else {
				assert.Nil(t, passwordManager.requiredSymbolsCheck)
			}
		})
	}
}

// Same password should produce different hashes
func TestUserPasswordManager_CreatePassword_Consistency(t *testing.T) {
	policy := config.PasswordPolicy{
		GlobalSalt:            "test-salt",
		MinSize:               6,
		MaxSize:               20,
		RequiredSymbolsRegExp: "",
		RequiredSymbolsHint:   "",
	}

	passwordManager := NewUserPasswordManager(policy)
	password := "testpassword123"

	// Create password twice
	hash1, err1 := passwordManager.CreatePassword(password)
	require.NoError(t, err1)

	hash2, err2 := passwordManager.CreatePassword(password)
	require.NoError(t, err2)

	// Hashes should be different (random salt)
	assert.NotEqual(t, hash1, hash2)

	// Both should verify correctly
	require.NoError(t, passwordManager.VerifyPassword(password, hash1))
	require.NoError(t, passwordManager.VerifyPassword(password, hash2))
}
