package communication

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
)

const (
	key = "0123456789abcdef"
)

type Manager struct {
	Address    *net.UDPAddr
	Connection *net.UDPConn
}

func InitializeManager() (Manager, error) {

	var addr *net.UDPAddr
	var conn *net.UDPConn
	var addrError, connError error

	addr, addrError = net.ResolveUDPAddr("udp", ":8080")
	if addrError != nil {
		return Manager{}, addrError
	}

	conn, connError = net.ListenUDP("udp", addr)
	if connError != nil {
		return Manager{}, connError
	}

	return Manager{Address: addr, Connection: conn}, nil
}

func (cm *Manager) Run() {

	defer cm.Connection.Close()

	buffer := make([]byte, 1024)

	for {
		n, addr, err := cm.Connection.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}

		receivedMsgWithMAC := buffer[:n]
		receivedMAC := receivedMsgWithMAC[:32]
		receivedEncryptedMsg := receivedMsgWithMAC[32:]

		// Verify MAC
		if verifyMAC([]byte(key), receivedMAC, receivedEncryptedMsg) {

			// Decrypt the received data
			decryptedMsg, err := decrypt([]byte(key), receivedEncryptedMsg)
			if err != nil {
				panic(err)
			}

			// Deserialize the Gob-encoded data into the original structure
			var receivedMsg Message1
			dec := gob.NewDecoder(bytes.NewReader(decryptedMsg))
			if err := dec.Decode(&receivedMsg); err != nil {
				panic(err)
			}

			fmt.Printf("Received message from %s: %+v\n", addr, receivedMsg)

		} else {
			fmt.Printf("Received message from %s: MAC verification failed.\n", addr)
		}
	}

}

func (m *Manager) Send(in_sAddress string) {

	// Connect to the server
	conn, err := net.Dial("udp", in_sAddress)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Create a message structure
	msg := Message1{
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

}

func InitializeProtocol() {

	gob.Register(Message1{})
	gob.Register(Message2{})

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

type Message1 struct {
	Text string
}

type Message2 struct {
	Text   string
	Number int
}
