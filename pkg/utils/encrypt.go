package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

func EncryptBytes(passphrase, message []byte) ([]byte, error) { // skipcq: TCV-001
	aesKey := sha256.Sum256(passphrase)
	block, err := aes.NewCipher(aesKey[:])
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	// IV needs to be unique, but doesn't have to be secure.
	// It's common to put it at the beginning of the ciphertext.
	cipherText := make([]byte, aes.BlockSize+len(message))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil { // skipcq: TCV-001
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], message)

	return cipherText, nil
}

func DecryptBytes(passphrase, cipherText []byte) ([]byte, error) {
	aesKey := sha256.Sum256(passphrase)
	block, err := aes.NewCipher(aesKey[:])
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	if len(cipherText) < aes.BlockSize { // skipcq: TCV-001
		err = errors.New("ciphertext block size is too short")
		return nil, err
	}

	temp := make([]byte, len(cipherText))
	copy(temp, cipherText)

	// IV needs to be unique, but doesn't have to be secure.
	// It's common to put it at the beginning of the ciphertext.
	iv := temp[:aes.BlockSize]
	temp = temp[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(temp, temp)

	return temp, nil
}
