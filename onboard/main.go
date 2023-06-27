package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"net"
	"time"
)

const (
	key = "0123456789abcdef" // 16-byte encryption key
)

type Message struct {
	Text string
}

func main() {

	// Register the Message type with gob
	gob.Register(Message{})

	// Start a server to listen for incoming UDP packets
	go startServer()

	for i := 0; i < 100000; i++ {
		// Connect to the server
		conn, err := net.Dial("udp", "localhost:8081")
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		// Create a message structure
		msg := Message{
			Text: "Hello, server!",
		}

		// Serialize the message structure to Gob
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(msg); err != nil {
			panic(err)
		}

		// Encrypt the Gob-encoded message
		encryptedData, err := encrypt([]byte(key), buf.Bytes())
		if err != nil {
			panic(err)
		}

		// Calculate MAC
		mac := generateMAC([]byte(key), encryptedData)

		// Append MAC to the encrypted message
		encryptedDataWithMAC := append(mac, encryptedData...)

		if _, err := conn.Write(encryptedDataWithMAC); err != nil {
			panic(err)
		}

		conn.Close()

	}

	for {
		time.Sleep(1 * time.Second)
	}

}

func startServer() {
	addr, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}

		receivedMsgWithMAC := buffer[:n]
		receivedMAC := receivedMsgWithMAC[:32]          // First 32 bytes are MAC
		receivedEncryptedMsg := receivedMsgWithMAC[32:] // Remaining bytes are encrypted message

		// Verify MAC
		if verifyMAC([]byte(key), receivedMAC, receivedEncryptedMsg) {

			// Decrypt the received data
			decryptedMsg, err := decrypt([]byte(key), receivedEncryptedMsg)
			if err != nil {
				panic(err)
			}

			// Deserialize the Gob-encoded data into the original structure
			var receivedMsg Message
			dec := gob.NewDecoder(bytes.NewReader(decryptedMsg))
			if err := dec.Decode(&receivedMsg); err != nil {
				panic(err)
			}

			fmt.Printf("Received message from %s: %+v\n", addr, receivedMsg)

			/*
				// Calculate MAC
				mac := generateMAC([]byte(key), receivedEncryptedMsg)

				// Append MAC to the encrypted message
				encryptedMsgWithMAC := append(mac, receivedEncryptedMsg...)

				if _, err := conn.WriteToUDP(encryptedMsgWithMAC, addr); err != nil {
					panic(err)
				}
			*/

		} else {
			fmt.Printf("Received message from %s: MAC verification failed.\n", addr)
		}
	}
}

func encrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a random initialization vector (IV)
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}

	// Encrypt the plaintext
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	copy(ciphertext[:aes.BlockSize], iv)
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

func decrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Extract the IV from the ciphertext
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// Decrypt the ciphertext
	plaintext := make([]byte, len(ciphertext))
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}

func generateMAC(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func verifyMAC(key, mac, data []byte) bool {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	expectedMAC := h.Sum(nil)
	return hmac.Equal(mac, expectedMAC)
}
