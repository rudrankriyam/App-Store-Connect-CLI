package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/99designs/keyring"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/config"
)

const (
	keyringService    = "asc"
	keyringItemPrefix = "asc:credential:"
	legacyKeychain    = "asc"
)

// Credential represents stored API credentials
type Credential struct {
	Name           string `json:"name"`
	KeyID          string `json:"key_id"`
	IssuerID       string `json:"issuer_id"`
	PrivateKeyPath string `json:"private_key_path"`
	IsDefault      bool   `json:"is_default"`
}

// Credentials stores multiple credentials
type Credentials struct {
	DefaultKey string       `json:"default_key"`
	Keys       []Credential `json:"keys"`
}

type credentialPayload struct {
	KeyID          string `json:"key_id"`
	IssuerID       string `json:"issuer_id"`
	PrivateKeyPath string `json:"private_key_path"`
}

func keyringConfig(keychainName string) keyring.Config {
	cfg := keyring.Config{
		ServiceName:                    keyringService,
		KeychainTrustApplication:       true,
		KeychainSynchronizable:         false,
		KeychainAccessibleWhenUnlocked: true,
		AllowedBackends: []keyring.BackendType{
			keyring.KeychainBackend,
			keyring.WinCredBackend,
			keyring.SecretServiceBackend,
			keyring.KWalletBackend,
			keyring.KeyCtlBackend,
		},
	}
	if keychainName != "" {
		cfg.KeychainName = keychainName
	}
	return cfg
}

var keyringOpener = func() (keyring.Keyring, error) {
	return keyring.Open(keyringConfig(""))
}

var legacyKeyringOpener = func() (keyring.Keyring, error) {
	return keyring.Open(keyringConfig(legacyKeychain))
}

// ValidateKeyFile validates that the private key file exists and is valid
func ValidateKeyFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat key file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("private key path is a directory")
	}
	if info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("private key file is too permissive; run: chmod 600 %q", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	// Parse the PEM block
	block, _ := pem.Decode(data)
	if block == nil {
		return fmt.Errorf("invalid PEM data")
	}

	// Try to parse as PKCS8 (App Store Connect keys are ECDSA)
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if _, ok := key.(*ecdsa.PrivateKey); ok {
			return nil
		}
		return fmt.Errorf("private key is not ECDSA")
	}

	// Try SEC1 EC private key as fallback
	if _, err := x509.ParseECPrivateKey(block.Bytes); err != nil {
		return fmt.Errorf("invalid private key format: %w", err)
	}

	return nil
}

// LoadPrivateKey loads the private key from the file
func LoadPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM data")
	}

	// Try PKCS8 first
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		ecdsaKey, ok := key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not ECDSA")
		}
		return ecdsaKey, nil
	}

	// Try SEC1 EC private key as fallback
	ecdsaKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	return ecdsaKey, nil
}

// StoreCredentials stores credentials in the keychain when available.
func StoreCredentials(name, keyID, issuerID, keyPath string) error {
	payload := credentialPayload{
		KeyID:          keyID,
		IssuerID:       issuerID,
		PrivateKeyPath: keyPath,
	}

	if err := storeInKeychain(name, payload); err == nil {
		// Successfully stored in keychain - clean up config file for security
		if err := clearConfigCredentials(); err != nil && !errors.Is(err, config.ErrNotFound) {
			// Log but don't fail - keychain is the authoritative storage
			_ = err
		}
		return saveDefaultName(name)
	} else if !isKeyringUnavailable(err) {
		return err
	}

	return storeInConfig(name, payload)
}

// StoreCredentialsConfig stores credentials in the config file only.
func StoreCredentialsConfig(name, keyID, issuerID, keyPath string) error {
	payload := credentialPayload{
		KeyID:          keyID,
		IssuerID:       issuerID,
		PrivateKeyPath: keyPath,
	}
	path, err := config.GlobalPath()
	if err != nil {
		return err
	}
	return storeInConfigAt(name, payload, path)
}

// StoreCredentialsConfigAt stores credentials in the specified config file.
func StoreCredentialsConfigAt(name, keyID, issuerID, keyPath, configPath string) error {
	payload := credentialPayload{
		KeyID:          keyID,
		IssuerID:       issuerID,
		PrivateKeyPath: keyPath,
	}
	return storeInConfigAt(name, payload, configPath)
}

// clearConfigCredentials clears credentials from the config file.
// This is called after successfully migrating to keychain storage.
func clearConfigCredentials() error {
	activePath, err := config.Path()
	if err != nil {
		return err
	}
	globalPath, err := config.GlobalPath()
	if err != nil {
		return err
	}
	if err := clearConfigCredentialsAt(activePath); err != nil && !errors.Is(err, config.ErrNotFound) {
		return err
	}
	if !sameConfigPath(activePath, globalPath) {
		if err := clearConfigCredentialsAt(globalPath); err != nil && !errors.Is(err, config.ErrNotFound) {
			return err
		}
	}
	return nil
}

func clearConfigCredentialsAt(path string) error {
	cfg, err := config.LoadAt(path)
	if err != nil {
		return err
	}
	cfg.KeyID = ""
	cfg.IssuerID = ""
	cfg.PrivateKeyPath = ""
	return config.SaveAt(path, cfg)
}

// ListCredentials lists all stored credentials
func ListCredentials() ([]Credential, error) {
	credentials, err := listFromKeychain()
	if err == nil {
		if len(credentials) > 0 {
			return credentials, nil
		}
		// Keychain available but empty - also check config (for --bypass-keychain case)
		configCreds, configErr := listFromConfig()
		if configErr == nil && len(configCreds) > 0 {
			return configCreds, nil
		}
		return credentials, nil
	}
	if !isKeyringUnavailable(err) {
		return nil, err
	}

	return listFromConfig()
}

// RemoveCredentials removes a named credential.
func RemoveCredentials(name string) error {
	err := removeFromKeychain(name)
	if err == nil {
		_ = removeFromLegacyKeychain(name)
		return clearDefaultNameIf(name)
	}
	if isKeyringUnavailable(err) {
		return removeFromConfigIfPresent(name)
	}
	if errors.Is(err, keyring.ErrKeyNotFound) {
		legacyErr := removeFromLegacyKeychain(name)
		if legacyErr == nil {
			return clearDefaultNameIf(name)
		}
		if isKeyringUnavailable(legacyErr) {
			return removeFromConfigIfPresent(name)
		}
		if errors.Is(legacyErr, keyring.ErrKeyNotFound) {
			if err := removeFromConfigIfPresent(name); err != nil {
				return err
			}
			return keyring.ErrKeyNotFound
		}
		return legacyErr
	}
	return err
}

// RemoveAllCredentials removes all stored credentials
func RemoveAllCredentials() error {
	if err := removeAllFromKeychain(); err == nil {
		_ = removeAllFromLegacyKeychain()
		// Clear config credentials but preserve other settings (app_id, timeout, etc.)
		return clearConfigCredentials()
	} else if !isKeyringUnavailable(err) {
		return err
	}
	// Clear config credentials but preserve other settings
	return clearConfigCredentials()
}

func sameConfigPath(left, right string) bool {
	return filepath.Clean(left) == filepath.Clean(right)
}

// GetDefaultCredentials returns the default credentials
func GetDefaultCredentials() (*config.Config, error) {
	credentials, err := listFromKeychain()
	if err == nil {
		name, err := defaultName()
		if err != nil {
			return nil, err
		}
		if name == "" && len(credentials) == 1 {
			name = credentials[0].Name
		}
		for _, cred := range credentials {
			if cred.Name == name {
				return &config.Config{
					KeyID:          cred.KeyID,
					IssuerID:       cred.IssuerID,
					PrivateKeyPath: cred.PrivateKeyPath,
					DefaultKeyName: cred.Name,
				}, nil
			}
		}
		// Keychain available but credentials not found - also check config (for --bypass-keychain case)
		if cfg, configErr := getDefaultFromConfig(); configErr == nil {
			return cfg, nil
		}
		return nil, fmt.Errorf("default credentials not found")
	}
	if !isKeyringUnavailable(err) {
		return nil, err
	}
	return getDefaultFromConfig()
}

func isKeyringUnavailable(err error) bool {
	return errors.Is(err, keyring.ErrNoAvailImpl)
}

func keyringKey(name string) string {
	return keyringItemPrefix + name
}

func storeInKeychain(name string, payload credentialPayload) error {
	kr, err := keyringOpener()
	if err != nil {
		return err
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode credentials: %w", err)
	}
	return kr.Set(keyring.Item{
		Key:   keyringKey(name),
		Data:  data,
		Label: fmt.Sprintf("ASC API Key (%s)", name),
	})
}

func listFromKeychain() ([]Credential, error) {
	kr, err := keyringOpener()
	if err != nil {
		return nil, err
	}
	credentials, err := listFromKeyring(kr)
	if err != nil {
		return nil, err
	}

	legacy, err := listFromLegacyKeychain()
	if err != nil || len(legacy) == 0 {
		return credentials, nil
	}

	existing := make(map[string]struct{}, len(credentials))
	for _, cred := range credentials {
		existing[cred.Name] = struct{}{}
	}

	var toMigrate []Credential
	for _, cred := range legacy {
		if _, ok := existing[cred.Name]; ok {
			_ = removeFromLegacyKeychain(cred.Name)
			continue
		}
		credentials = append(credentials, cred)
		toMigrate = append(toMigrate, cred)
	}

	if len(toMigrate) > 0 {
		migrateLegacyCredentials(toMigrate)
	}
	return credentials, nil
}

func listFromLegacyKeychain() ([]Credential, error) {
	kr, err := legacyKeyringOpener()
	if err != nil {
		return nil, err
	}
	return listFromKeyring(kr)
}

func listFromKeyring(kr keyring.Keyring) ([]Credential, error) {
	keys, err := kr.Keys()
	if err != nil {
		return nil, err
	}

	defaultName, _ := defaultName()
	credentials := []Credential{}
	for _, key := range keys {
		if !strings.HasPrefix(key, keyringItemPrefix) {
			continue
		}
		item, err := kr.Get(key)
		if err != nil {
			if errors.Is(err, keyring.ErrKeyNotFound) {
				continue
			}
			return nil, err
		}
		var payload credentialPayload
		if err := json.Unmarshal(item.Data, &payload); err != nil {
			return nil, fmt.Errorf("invalid keychain entry %q: %w", key, err)
		}
		name := strings.TrimPrefix(key, keyringItemPrefix)
		credentials = append(credentials, Credential{
			Name:           name,
			KeyID:          payload.KeyID,
			IssuerID:       payload.IssuerID,
			PrivateKeyPath: payload.PrivateKeyPath,
			IsDefault:      name == defaultName,
		})
	}

	return credentials, nil
}

func migrateLegacyCredentials(credentials []Credential) {
	for _, cred := range credentials {
		payload := credentialPayload{
			KeyID:          cred.KeyID,
			IssuerID:       cred.IssuerID,
			PrivateKeyPath: cred.PrivateKeyPath,
		}
		if err := storeInKeychain(cred.Name, payload); err != nil {
			continue
		}
		_ = removeFromLegacyKeychain(cred.Name)
	}
}

func removeFromConfigIfPresent(name string) error {
	if err := removeFromConfig(name); err != nil {
		if errors.Is(err, config.ErrNotFound) {
			return nil
		}
		return err
	}
	return nil
}

func removeFromKeychain(name string) error {
	kr, err := keyringOpener()
	if err != nil {
		return err
	}
	return kr.Remove(keyringKey(name))
}

func removeFromLegacyKeychain(name string) error {
	kr, err := legacyKeyringOpener()
	if err != nil {
		return err
	}
	return kr.Remove(keyringKey(name))
}

func removeAllFromKeychain() error {
	kr, err := keyringOpener()
	if err != nil {
		return err
	}
	keys, err := kr.Keys()
	if err != nil {
		return err
	}
	for _, key := range keys {
		if strings.HasPrefix(key, keyringItemPrefix) {
			if err := kr.Remove(key); err != nil {
				return err
			}
		}
	}
	return nil
}

func removeAllFromLegacyKeychain() error {
	kr, err := legacyKeyringOpener()
	if err != nil {
		return err
	}
	keys, err := kr.Keys()
	if err != nil {
		return err
	}
	for _, key := range keys {
		if strings.HasPrefix(key, keyringItemPrefix) {
			if err := kr.Remove(key); err != nil {
				return err
			}
		}
	}
	return nil
}

func storeInConfig(name string, payload credentialPayload) error {
	path, err := config.Path()
	if err != nil {
		return err
	}
	return storeInConfigAt(name, payload, path)
}

func storeInConfigAt(name string, payload credentialPayload, configPath string) error {
	cfg, err := config.LoadAt(configPath)
	if err != nil && err != config.ErrNotFound {
		return err
	}
	if cfg == nil {
		cfg = &config.Config{}
	}
	cfg.KeyID = payload.KeyID
	cfg.IssuerID = payload.IssuerID
	cfg.PrivateKeyPath = payload.PrivateKeyPath
	cfg.DefaultKeyName = name
	return config.SaveAt(configPath, cfg)
}

func hasCompleteCredentials(cfg *config.Config) bool {
	return cfg != nil && cfg.KeyID != "" && cfg.IssuerID != "" && cfg.PrivateKeyPath != ""
}

func hasAnyCredentials(cfg *config.Config) bool {
	if cfg == nil {
		return false
	}
	return cfg.KeyID != "" || cfg.IssuerID != "" || cfg.PrivateKeyPath != ""
}

func loadGlobalConfigForCredentials() (*config.Config, error) {
	if strings.TrimSpace(os.Getenv("ASC_CONFIG_PATH")) != "" {
		return nil, config.ErrNotFound
	}
	path, err := config.GlobalPath()
	if err != nil {
		return nil, err
	}
	return config.LoadAt(path)
}

func listFromConfig() ([]Credential, error) {
	cfg, err := config.Load()
	if err != nil {
		if err == config.ErrNotFound {
			return []Credential{}, nil
		}
		return nil, err
	}
	if !hasCompleteCredentials(cfg) {
		if hasAnyCredentials(cfg) {
			return []Credential{}, nil
		}
		globalCfg, err := loadGlobalConfigForCredentials()
		if err != nil {
			if err == config.ErrNotFound {
				return []Credential{}, nil
			}
			return nil, err
		}
		if !hasCompleteCredentials(globalCfg) {
			return []Credential{}, nil
		}
		cfg = globalCfg
	}
	if cfg.KeyID == "" || cfg.IssuerID == "" || cfg.PrivateKeyPath == "" {
		return []Credential{}, nil
	}
	credentials := []Credential{
		{
			Name:           cfg.DefaultKeyName,
			KeyID:          cfg.KeyID,
			IssuerID:       cfg.IssuerID,
			PrivateKeyPath: cfg.PrivateKeyPath,
			IsDefault:      true,
		},
	}
	if credentials[0].Name == "" {
		credentials[0].Name = "default"
	}
	return credentials, nil
}

func getDefaultFromConfig() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		if err != config.ErrNotFound {
			return nil, err
		}
	}
	if !hasCompleteCredentials(cfg) {
		if hasAnyCredentials(cfg) {
			return nil, fmt.Errorf("incomplete credentials")
		}
		globalCfg, globalErr := loadGlobalConfigForCredentials()
		if globalErr != nil {
			if globalErr == config.ErrNotFound {
				if err == config.ErrNotFound {
					return nil, err
				}
				return nil, fmt.Errorf("incomplete credentials")
			}
			return nil, globalErr
		}
		if !hasCompleteCredentials(globalCfg) {
			return nil, fmt.Errorf("incomplete credentials")
		}
		cfg = globalCfg
	}
	return cfg, nil
}

func saveDefaultName(name string) error {
	cfg, err := config.Load()
	if err != nil && err != config.ErrNotFound {
		return err
	}
	if cfg == nil {
		cfg = &config.Config{}
	}
	cfg.DefaultKeyName = name
	return config.Save(cfg)
}

func defaultName() (string, error) {
	cfg, err := config.Load()
	if err != nil {
		if err == config.ErrNotFound {
			return "", nil
		}
		return "", err
	}
	return cfg.DefaultKeyName, nil
}

func clearDefaultNameIf(name string) error {
	cfg, err := config.Load()
	if err != nil {
		if err == config.ErrNotFound {
			return nil
		}
		return err
	}
	if cfg.DefaultKeyName == name {
		cfg.DefaultKeyName = ""
		return config.Save(cfg)
	}
	return nil
}

func removeFromConfig(name string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.DefaultKeyName == name || name == "" {
		cfg.KeyID = ""
		cfg.IssuerID = ""
		cfg.PrivateKeyPath = ""
		cfg.DefaultKeyName = ""
	}
	return config.Save(cfg)
}
