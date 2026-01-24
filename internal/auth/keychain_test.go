package auth

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/99designs/keyring"
)

func TestValidateKeyFilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "AuthKey.p8")

	writeECDSAPEM(t, keyPath, 0o644, true)

	if err := ValidateKeyFile(keyPath); err == nil {
		t.Fatalf("expected permission error for insecure key file")
	}
}

func TestValidateKeyFileSuccess(t *testing.T) {
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "AuthKey.p8")

	writeECDSAPEM(t, keyPath, 0o600, true)

	if err := ValidateKeyFile(keyPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateKeyFileDirectory(t *testing.T) {
	tempDir := t.TempDir()

	if err := ValidateKeyFile(tempDir); err == nil {
		t.Fatalf("expected error for directory path")
	}
}

func TestLoadPrivateKeyPKCS8(t *testing.T) {
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "AuthKey.p8")

	writeECDSAPEM(t, keyPath, 0o600, true)

	key, err := LoadPrivateKey(keyPath)
	if err != nil {
		t.Fatalf("LoadPrivateKey() error: %v", err)
	}
	if key == nil {
		t.Fatal("expected non-nil key")
	}
}

func TestLoadPrivateKeySEC1(t *testing.T) {
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "AuthKey-EC.p8")

	writeECDSAPEM(t, keyPath, 0o600, false)

	key, err := LoadPrivateKey(keyPath)
	if err != nil {
		t.Fatalf("LoadPrivateKey() error: %v", err)
	}
	if key == nil {
		t.Fatal("expected non-nil key")
	}
}

func TestStoreAndListCredentials(t *testing.T) {
	withArrayKeyring(t)

	if err := StoreCredentials("my-key", "KEY123", "ISS456", "/tmp/AuthKey.p8"); err != nil {
		t.Fatalf("StoreCredentials() error: %v", err)
	}

	creds, err := ListCredentials()
	if err != nil {
		t.Fatalf("ListCredentials() error: %v", err)
	}
	if len(creds) != 1 {
		t.Fatalf("expected 1 credential, got %d", len(creds))
	}
	if creds[0].Name != "my-key" {
		t.Fatalf("expected credential name %q, got %q", "my-key", creds[0].Name)
	}
	if !creds[0].IsDefault {
		t.Fatalf("expected credential to be default")
	}
}

func TestRemoveAllCredentials(t *testing.T) {
	withArrayKeyring(t)

	if err := StoreCredentials("my-key", "KEY123", "ISS456", "/tmp/AuthKey.p8"); err != nil {
		t.Fatalf("StoreCredentials() error: %v", err)
	}

	if err := RemoveAllCredentials(); err != nil {
		t.Fatalf("RemoveAllCredentials() error: %v", err)
	}

	creds, err := ListCredentials()
	if err != nil {
		t.Fatalf("ListCredentials() error: %v", err)
	}
	if len(creds) != 0 {
		t.Fatalf("expected no credentials after removal, got %d", len(creds))
	}
}

func TestStoreCredentialsFallbackToConfig(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	previous := keyringOpener
	keyringOpener = func() (keyring.Keyring, error) {
		return nil, keyring.ErrNoAvailImpl
	}
	t.Cleanup(func() {
		keyringOpener = previous
	})

	if err := StoreCredentials("fallback", "KEY123", "ISS456", "/tmp/AuthKey.p8"); err != nil {
		t.Fatalf("StoreCredentials() error: %v", err)
	}

	creds, err := ListCredentials()
	if err != nil {
		t.Fatalf("ListCredentials() error: %v", err)
	}
	if len(creds) != 1 {
		t.Fatalf("expected 1 credential, got %d", len(creds))
	}
	if creds[0].KeyID != "KEY123" {
		t.Fatalf("expected KeyID KEY123, got %q", creds[0].KeyID)
	}
}

func TestListCredentials_MigratesLegacyEntries(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	newKr, legacyKr := withSeparateKeyrings(t)

	storeCredentialInKeyring(t, newKr, "new-key", "NEW123", "ISSNEW", "/tmp/new.p8")
	storeCredentialInKeyring(t, legacyKr, "legacy-key", "OLD123", "ISSOLD", "/tmp/old.p8")

	creds, err := ListCredentials()
	if err != nil {
		t.Fatalf("ListCredentials() error: %v", err)
	}
	if len(creds) != 2 {
		t.Fatalf("expected 2 credentials, got %d", len(creds))
	}

	if _, err := legacyKr.Get(keyringKey("legacy-key")); !errors.Is(err, keyring.ErrKeyNotFound) {
		t.Fatalf("expected legacy credential to be removed, got %v", err)
	}
	if _, err := newKr.Get(keyringKey("legacy-key")); err != nil {
		t.Fatalf("expected legacy credential to be migrated, got %v", err)
	}
}

func TestListCredentials_RemovesLegacyDuplicates(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	newKr, legacyKr := withSeparateKeyrings(t)

	storeCredentialInKeyring(t, newKr, "shared-key", "NEW123", "ISSNEW", "/tmp/new.p8")
	storeCredentialInKeyring(t, legacyKr, "shared-key", "OLD123", "ISSOLD", "/tmp/old.p8")

	creds, err := ListCredentials()
	if err != nil {
		t.Fatalf("ListCredentials() error: %v", err)
	}
	if len(creds) != 1 {
		t.Fatalf("expected 1 credential, got %d", len(creds))
	}

	if _, err := legacyKr.Get(keyringKey("shared-key")); !errors.Is(err, keyring.ErrKeyNotFound) {
		t.Fatalf("expected legacy duplicate to be removed, got %v", err)
	}
}

func TestRemoveCredentials_FallsBackToLegacy(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	_, legacyKr := withSeparateKeyrings(t)

	storeCredentialInKeyring(t, legacyKr, "legacy-only", "OLD123", "ISSOLD", "/tmp/old.p8")

	if err := RemoveCredentials("legacy-only"); err != nil {
		t.Fatalf("RemoveCredentials() error: %v", err)
	}
	if _, err := legacyKr.Get(keyringKey("legacy-only")); !errors.Is(err, keyring.ErrKeyNotFound) {
		t.Fatalf("expected legacy credential to be removed, got %v", err)
	}
}

func writeECDSAPEM(t *testing.T, path string, mode os.FileMode, pkcs8 bool) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() error: %v", err)
	}

	var der []byte
	if pkcs8 {
		der, err = x509.MarshalPKCS8PrivateKey(key)
	} else {
		der, err = x509.MarshalECPrivateKey(key)
	}
	if err != nil {
		t.Fatalf("marshal key error: %v", err)
	}

	var buf bytes.Buffer
	blockType := "PRIVATE KEY"
	if !pkcs8 {
		blockType = "EC PRIVATE KEY"
	}
	if err := pem.Encode(&buf, &pem.Block{Type: blockType, Bytes: der}); err != nil {
		t.Fatalf("pem encode error: %v", err)
	}

	if err := os.WriteFile(path, buf.Bytes(), mode); err != nil {
		t.Fatalf("write key file error: %v", err)
	}
}

func withArrayKeyring(t *testing.T) {
	t.Helper()
	previous := keyringOpener
	previousLegacy := legacyKeyringOpener
	kr := keyring.NewArrayKeyring([]keyring.Item{})
	keyringOpener = func() (keyring.Keyring, error) {
		return kr, nil
	}
	t.Cleanup(func() {
		keyringOpener = previous
		legacyKeyringOpener = previousLegacy
	})
	legacyKeyringOpener = func() (keyring.Keyring, error) {
		return kr, nil
	}
}

func withSeparateKeyrings(t *testing.T) (keyring.Keyring, keyring.Keyring) {
	t.Helper()
	previous := keyringOpener
	previousLegacy := legacyKeyringOpener
	kr := keyring.NewArrayKeyring([]keyring.Item{})
	legacyKr := keyring.NewArrayKeyring([]keyring.Item{})
	keyringOpener = func() (keyring.Keyring, error) {
		return kr, nil
	}
	legacyKeyringOpener = func() (keyring.Keyring, error) {
		return legacyKr, nil
	}
	t.Cleanup(func() {
		keyringOpener = previous
		legacyKeyringOpener = previousLegacy
	})
	return kr, legacyKr
}

func storeCredentialInKeyring(t *testing.T, kr keyring.Keyring, name, keyID, issuerID, keyPath string) {
	t.Helper()
	payload := credentialPayload{
		KeyID:          keyID,
		IssuerID:       issuerID,
		PrivateKeyPath: keyPath,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload error: %v", err)
	}
	if err := kr.Set(keyring.Item{Key: keyringKey(name), Data: data}); err != nil {
		t.Fatalf("store keyring item error: %v", err)
	}
}
