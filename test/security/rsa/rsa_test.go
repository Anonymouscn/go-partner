package rsa

import (
	"fmt"
	"github.com/Anonymouscn/go-partner/security/rsa"
	"testing"
)

func TestRSAWorkflow(t *testing.T) {
	privateKey, publicKey, err := rsa.GenerateRSAKey(4096)
	if err != nil {
		fmt.Printf("Generate key error: %v\n", err.Error())
		return
	}
	fmt.Printf("Generate key:\nPrivate key:\n%v\n\nPublic key:\n%v\n\n", privateKey, publicKey)

	// ================================================ 加密/解密测试 ================================================ //

	plainText := "This is a test plain text"
	fmt.Printf("Set a plain text: %v\n", plainText)
	cipherText, err := rsa.EncryptWithBase64(plainText, publicKey)
	if err != nil {
		fmt.Printf("encrypt text error: %v", err.Error())
		return
	}
	fmt.Printf("Get a cipher text: %v\n", cipherText)

	decryptText, err := rsa.DecryptWithBase64(cipherText, privateKey)
	if err != nil {
		fmt.Printf("decrypt text error: %v", err.Error())
		return
	}
	fmt.Printf("Get a decrypt text: %v\n", decryptText)
	assume := "decrypt text is the same as plain text"
	if decryptText != plainText {
		assume = "decrypt text is not the same as plain text"
	}
	fmt.Printf("conclusion: %v\n", assume)

	// ================================================ 签名/验签测试 ================================================ //

	signText, err := rsa.SignWithBase64(plainText, privateKey)
	if err != nil {
		fmt.Printf("sign error: %v", err.Error())
		return
	}
	fmt.Printf("Get sign: %v\n", signText)
	verify, err := rsa.VerifySignWithBase64(plainText, signText, publicKey)
	if err != nil {
		fmt.Printf("verify error: %v\n", err.Error())
		return
	}
	fmt.Printf("verify result: %v\n", verify)
}

func TestRSASign(t *testing.T) {
	privateKey := "$privateKey"
	plaintext := "$plaintext"
	signText, _ := rsa.SignWithBase64(plaintext, privateKey)
	fmt.Println(signText)
}
