package communication

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
)

const (
	default_key = "0123456789abcdef"
)

type CommunicationHandler func(interface{})

type Manager struct {
	SystemID              string
	Address               *net.UDPAddr
	Connection            *net.UDPConn
	Key                   string
	Handler               func(Message interface{})
	GroundstationAddress  string
	OnboardAddress        string
	AntennaTrackerAddress string
	PacketCounter         uint
}

func InitializeManager(in_systemid, in_port, in_key, in_onboardAddress, in_groundstationAddress, in_antennaTrackerAddress string, in_handler CommunicationHandler) (*Manager, error) {

	var addr *net.UDPAddr
	var conn *net.UDPConn
	var addrError, connError error

	addr, addrError = net.ResolveUDPAddr("udp", ":"+in_port)
	if addrError != nil {
		return &Manager{}, addrError
	}

	conn, connError = net.ListenUDP("udp", addr)
	if connError != nil {
		return &Manager{}, connError
	}

	var key string

	if in_key == "" {
		key = default_key
	} else {
		key = in_key
	}

	return &Manager{SystemID: in_systemid, Address: addr, Connection: conn, Key: key, PacketCounter: 0, OnboardAddress: in_onboardAddress, GroundstationAddress: in_groundstationAddress, AntennaTrackerAddress: in_antennaTrackerAddress, Handler: in_handler}, nil
}

func (m *Manager) GetCounter() uint {
	m.PacketCounter = m.PacketCounter + 1
	return m.PacketCounter - 1
}

func (m *Manager) GetKey() string {
	return m.Key
}

func (m *Manager) Run() {

	defer m.Connection.Close()

	buffer := make([]byte, 1024)

	var packet Communication_Packet

	var decodeError error

	for {
		n, _, err := m.Connection.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}

		receivedPacketWithMAC := buffer[:n]
		receivedMAC := receivedPacketWithMAC[:32]
		receivedEncryptedPacket := receivedPacketWithMAC[32:]

		// Verify MAC
		if verifyMAC([]byte(m.Key), receivedMAC, receivedEncryptedPacket) {

			// Decrypt the received data
			decryptedPacket, decryptErr := decrypt([]byte(m.Key), receivedEncryptedPacket)
			if decryptErr != nil {
				panic(decryptErr)
			}

			fmt.Printf("%s\n", decryptedPacket)

			// Decode
			decodeError = json.Unmarshal(decryptedPacket, &packet)
			if decodeError == nil {

				fmt.Printf("received: %v %v %v %v\n", packet.Counter, packet.Type, packet.SubType, packet.Message)

				continue
			}

		}
	}

}

func (m *Manager) Send2Onboard(in_message interface{}) {

	// Connect to the server

	conn, err := net.Dial("udp", m.OnboardAddress)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Serialize the message structure to Gob
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(in_message); err != nil {
		panic(err)
	}

	// Encrypt the Gob-encoded message
	encryptedData, err := encrypt([]byte(m.Key), buf.Bytes())
	if err != nil {
		panic(err)
	}

	// Calculate MAC
	mac := generateMAC([]byte(m.Key), encryptedData)

	// Append MAC to the encrypted message
	encryptedDataWithMAC := append(mac, encryptedData...)

	if _, err := conn.Write(encryptedDataWithMAC); err != nil {
		panic(err)
	}

	m.PacketCounter = m.PacketCounter + 1

}

func (m *Manager) Send2Groundstation(in_message interface{}) {

	if msg, ok := in_message.(Message); ok {

		// Connect to the server
		conn, err := net.Dial("udp", m.GroundstationAddress)
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		// encode message
		encodedMessage, err := json.Marshal(in_message)
		if err != nil {
			panic(err)
		}

		// encode packet
		var packet = Communication_Packet{Counter: m.GetCounter(), Type: msg.getType(), SubType: msg.getSubType(), Message: string(encodedMessage)}
		encodedPacket, err := json.Marshal(packet)
		if err != nil {
			panic(err)
		}

		// Encrypt the Gob-encoded message
		encryptedData, err := encrypt([]byte(m.Key), encodedPacket)
		if err != nil {
			panic(err)
		}

		// Calculate MAC
		mac := generateMAC([]byte(m.Key), encryptedData)

		// Append MAC to the encrypted message
		encryptedDataWithMAC := append(mac, encryptedData...)

		if _, err := conn.Write(encryptedDataWithMAC); err != nil {
			panic(err)
		}

		m.PacketCounter = m.PacketCounter + 1
	}

}

func (m *Manager) Send2AntennaTracker(in_message interface{}) {

	// Connect to the server

	conn, err := net.Dial("udp", m.AntennaTrackerAddress)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Serialize the message structure to Gob
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(in_message); err != nil {
		panic(err)
	}

	// Encrypt the Gob-encoded message
	encryptedData, err := encrypt([]byte(m.Key), buf.Bytes())
	if err != nil {
		panic(err)
	}

	// Calculate MAC
	mac := generateMAC([]byte(m.Key), encryptedData)

	// Append MAC to the encrypted message
	encryptedDataWithMAC := append(mac, encryptedData...)

	if _, err := conn.Write(encryptedDataWithMAC); err != nil {
		panic(err)
	}

	m.PacketCounter = m.PacketCounter + 1

}

func DefaultHandler(in_message interface{}) {

	// Decoding based on the type of message
	switch message := in_message.(type) {

	case Communication_Message_Ping:
		// Handle decoding for message type 1
		fmt.Printf("Received PING: %+v\n", message)

	case Communication_Message_ACK:
		// Handle decoding for message type 2
		fmt.Printf("Received ACK: %+v\n", message)

	}
}

func InitializeProtocol() {

	gob.Register(Communication_Message_Ping{})
	gob.Register(Communication_Message_ACK{})
	gob.Register(Communication_Message_NACK{})
	gob.Register(Communication_Message_ControlMode_Set{})
	gob.Register(Communication_Message_ControlMode_Report{})

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

type Communication_Packet struct {
	Counter uint
	Type    uint
	SubType uint
	Message string
}

type Message interface {
	getType() uint
	getSubType() uint
}

// Message - Ping

type Communication_Message_Ping struct {
	Counter  uint
	SenderID string
}

func (m *Communication_Message_Ping) getType() uint {
	return 1
}
func (m *Communication_Message_Ping) getSubType() uint {
	return 1
}

// Message - ACK

type Communication_Message_ACK struct {
	Counter  uint
	SenderID string
	ACKId    uint
}

func (m *Communication_Message_ACK) getType() uint {
	return 2
}
func (m *Communication_Message_ACK) getSubType() uint {
	return 1
}

// Message - NACK

type Communication_Message_NACK struct {
	Counter  uint
	SenderID string
	NACKId   uint
}

func (m *Communication_Message_NACK) getType() uint {
	return 2
}
func (m *Communication_Message_NACK) getSubType() uint {
	return 2
}

type Communication_Message_ControlMode_Set struct {
	Counter     uint
	SenderID    string
	ControlMode uint
}

type Communication_Message_ControlMode_Report struct {
	Counter     uint
	SenderID    string
	ControlMode uint
}

type Communication_Message_TargetTracking_Set struct {
	Counter  uint
	SenderID string
	CenterX  float64
	CenterY  float64
}

type Communication_Message_TargetTracking_Get struct {
	Counter  uint
	SenderID string
	XMin     float64
	XMax     float64
	YMin     float64
	YMax     float64
}

type Communication_Message_TargetTrackingStatus struct {
	Counter  uint
	SenderID string
	Status   bool
}

type Communication_Message_GuidanceState struct {
	Counter        uint
	SenderID       string
	DistanceToNext float64
	HeadingToNext  float64
}

type Communication_Message_BatteryVoltage struct {
	Counter  uint
	SenderID string
	Voltage  float64
}

type Communication_Message_OnboardSystems struct {
	Counter        uint
	SenderID       string
	Video1In       uint
	Video2In       uint
	TargetTracking uint
	FCRX           uint
	FCTX           uint
	ControlLoop    uint
}
