package communication

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

const (
	default_key = "0123456789abcdef"
)

type CommunicationHandler func(Communication_Packet)

type Manager struct {
	SystemID              string
	Address               *net.UDPAddr
	Connection            *net.UDPConn
	Key                   string
	Handler               func(Communication_Packet)
	GroundstationAddress  string
	OnboardAddress        string
	AntennaTrackerAddress string
	packetCounter         uint
	packetCounterMutex    sync.Mutex
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

	return &Manager{SystemID: in_systemid, Address: addr, Connection: conn, Key: key, packetCounter: 0, OnboardAddress: in_onboardAddress, GroundstationAddress: in_groundstationAddress, AntennaTrackerAddress: in_antennaTrackerAddress, Handler: in_handler}, nil
}

func (m *Manager) GetCounter() uint {
	return m.packetCounter
}

func (m *Manager) IncrementCounter() uint {
	m.packetCounterMutex.Lock()
	defer m.packetCounterMutex.Unlock()

	m.packetCounter = m.packetCounter + 1
	return m.packetCounter - 1

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

		n, addr, err := m.Connection.ReadFromUDP(buffer)
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

			// Decode
			decodeError = json.Unmarshal(decryptedPacket, &packet)
			if decodeError == nil {

				if packet.Type != 2 {
					go m.Send2Groundstation(&Communication_Message_ACK{ACKId: packet.Counter})
				}

				DefaultHandler(packet)

			} else {
				panic(decodeError)
			}

		}
	}

}

func (m *Manager) Send2Any(in_message Message, in_address string) {

	// Connect to the server
	conn, err := net.Dial("udp", in_address)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// encode packet
	var packet = Communication_Packet{Counter: m.IncrementCounter(), Type: in_message.GetType(), SubType: in_message.GetSubType(), Message: in_message.Encode()}
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

}

func (m *Manager) Send2Onboard(in_message Message) {

	// Connect to the server
	conn, err := net.Dial("udp", m.OnboardAddress)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// encode packet
	var packet = Communication_Packet{Counter: m.IncrementCounter(), Type: in_message.GetType(), SubType: in_message.GetSubType(), Message: in_message.Encode()}
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

}

func (m *Manager) Send2Groundstation(in_message Message) {

	// Connect to the server
	conn, err := net.Dial("udp", m.GroundstationAddress)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// encode packet
	var packet = Communication_Packet{Counter: m.IncrementCounter(), Type: in_message.GetType(), SubType: in_message.GetSubType(), Message: in_message.Encode()}
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

}

func (m *Manager) Send2AntennaTracker(in_message Message) {

	// Connect to the server
	conn, err := net.Dial("udp", m.AntennaTrackerAddress)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// encode packet
	var packet = Communication_Packet{Counter: m.IncrementCounter(), Type: in_message.GetType(), SubType: in_message.GetSubType(), Message: in_message.Encode()}
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

}

func DefaultHandler(in_packet Communication_Packet) {

	fmt.Printf("%+v\n", in_packet)

}

func InitializeProtocol() {

	gob.Register(Communication_Message_Ping{})
	gob.Register(Communication_Message_ACK{})
	gob.Register(Communication_Message_NACK{})

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
	GetType() uint
	GetSubType() uint
	Encode() string
}

// Message - Ping

type Communication_Message_Ping struct {
	Time uint
}

func (m *Communication_Message_Ping) GetType() uint {
	return 1
}
func (m *Communication_Message_Ping) GetSubType() uint {
	return 1
}
func (m *Communication_Message_Ping) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// Message - ACK

type Communication_Message_ACK struct {
	ACKId uint
}

func (m *Communication_Message_ACK) GetType() uint {
	return 2
}
func (m *Communication_Message_ACK) GetSubType() uint {
	return 1
}
func (m *Communication_Message_ACK) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// Message - NACK

type Communication_Message_NACK struct {
	NACKId uint
}

func (m *Communication_Message_NACK) GetType() uint {
	return 2
}
func (m *Communication_Message_NACK) GetSubType() uint {
	return 2
}
func (m *Communication_Message_NACK) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

type Communication_Message_ControlMode_Set struct {
	ControlMode uint
}

type Communication_Message_ControlMode_Report struct {
	ControlMode uint
}

type Communication_Message_TargetTracking_Set struct {
	CenterX float64
	CenterY float64
}

type Communication_Message_TargetTracking_Get struct {
	XMin float64
	XMax float64
	YMin float64
	YMax float64
}

type Communication_Message_TargetTrackingStatus struct {
	Status bool
}

type Communication_Message_GuidanceState struct {
	DistanceToNext float64
	HeadingToNext  float64
}

type Communication_Message_BatteryVoltage struct {
	Voltage float64
}

type Communication_Message_OnboardSystems struct {
	Video1In       uint
	Video2In       uint
	TargetTracking uint
	FCRX           uint
	FCTX           uint
	ControlLoop    uint
}
