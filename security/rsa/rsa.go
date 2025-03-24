package rsa

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

// GenerateRSAKey generates an RSA key pair with the specified bit size
// Returns the private and public key in PEM format
func GenerateRSAKey(bits int) (privateKey, publicKey string, err error) {
	// Generate the RSA key pair
	priKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate RSA key: %v", err)
	}

	// Encode private key to PKCS8 format for consistency
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(priKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal private key: %v", err)
	}

	privateKeyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	prvKey := pem.EncodeToMemory(privateKeyBlock)

	// Encode public key
	pubKey := &priKey.PublicKey
	pubBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal public key: %v", err)
	}

	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	pubKeyPEM := pem.EncodeToMemory(pubBlock)

	return string(prvKey), string(pubKeyPEM), nil
}

// ================================================ Encryption and Decryption ================================================ //

// EncryptWithBase64 encrypts data with RSA public key and returns Base64 encoded string
// Handles data larger than the maximum block size by chunking
func EncryptWithBase64(plaintext, publicKey string) (string, error) {
	// Decode the PEM-encoded public key
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return "", fmt.Errorf("failed to decode public key")
	}

	// Parse the public key
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %v", err)
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("not an RSA public key")
	}

	// Calculate maximum encryption block size
	keySize := rsaPubKey.Size()
	maxEncryptSize := keySize - 11 // PKCS1v15 padding requires 11 bytes

	plaintextBytes := []byte(plaintext)
	var result bytes.Buffer

	// Process the plaintext in chunks
	for i := 0; i < len(plaintextBytes); i += maxEncryptSize {
		end := i + maxEncryptSize
		if end > len(plaintextBytes) {
			end = len(plaintextBytes)
		}

		chunk := plaintextBytes[i:end]
		encryptedChunk, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPubKey, chunk)
		if err != nil {
			return "", fmt.Errorf("encryption failed: %v", err)
		}

		result.Write(encryptedChunk)
	}

	// Encode the encrypted data to Base64
	return base64.StdEncoding.EncodeToString(result.Bytes()), nil
}

// DecryptWithBase64 decrypts Base64 encoded data with RSA private key
// Handles data larger than the maximum block size by chunking
func DecryptWithBase64(ciphertext, privateKey string) (string, error) {
	// Decode the Base64 encoded ciphertext
	encryptedBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode Base64 data: %v", err)
	}

	// Decode the PEM-encoded private key
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return "", fmt.Errorf("failed to decode private key")
	}

	// Parse the private key
	priKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %v", err)
	}

	rsaPrivateKey, ok := priKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("not an RSA private key")
	}

	// Calculate block size for decryption
	keySize := rsaPrivateKey.Size()
	var result bytes.Buffer

	// Process the ciphertext in chunks
	for i := 0; i < len(encryptedBytes); i += keySize {
		end := i + keySize
		if end > len(encryptedBytes) {
			end = len(encryptedBytes)
		}

		// RSA decryption blocks must be exactly keySize bytes
		// If we have a partial block at the end, it's an error
		if end-i != keySize {
			return "", fmt.Errorf("invalid ciphertext block size")
		}

		chunk := encryptedBytes[i:end]
		decryptedChunk, err := rsa.DecryptPKCS1v15(rand.Reader, rsaPrivateKey, chunk)
		if err != nil {
			return "", fmt.Errorf("decryption failed: %v", err)
		}

		result.Write(decryptedChunk)
	}

	return result.String(), nil
}

// ================================================ Signing and Verification ================================================ //

// SignWithBase64 signs data with RSA private key and returns Base64 encoded signature
func SignWithBase64(plaintext, privateKey string) (string, error) {
	// Decode the PEM-encoded private key
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return "", fmt.Errorf("failed to decode private key")
	}

	// Parse the private key
	priKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %v", err)
	}

	rsaPrivateKey, ok := priKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("not an RSA private key")
	}

	// Calculate SHA-256 hash of the plaintext
	hash := sha256.Sum256([]byte(plaintext))

	// Sign the hash with the private key
	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("signing failed: %v", err)
	}

	// Encode the signature to Base64
	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifySignWithBase64 verifies a Base64 encoded signature with RSA public key
func VerifySignWithBase64(plaintext, sign, publicKey string) (bool, error) {
	// Decode the Base64 encoded signature
	signature, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %v", err)
	}

	// Decode the PEM-encoded public key
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return false, fmt.Errorf("failed to decode public key")
	}

	// Parse the public key
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse public key: %v", err)
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("not an RSA public key")
	}

	// Calculate SHA-256 hash of the plaintext
	hash := sha256.Sum256([]byte(plaintext))

	// Verify the signature
	err = rsa.VerifyPKCS1v15(rsaPubKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		return false, nil // Signature verification failed, but not an error
	}

	return true, nil
}
