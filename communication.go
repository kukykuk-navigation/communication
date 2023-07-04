package communication

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"net"
	"sync"
)

const (
	default_key = "0123456789abcdef"
)

type CommunicationHandler func(Communication_Packet)

type Manager struct {
	SystemID                   string
	Address                    *net.UDPAddr
	Connection                 *net.UDPConn
	Handler                    func(Communication_Packet)
	Key                        string
	KeyMutex                   sync.Mutex
	GroundstationAddress       string
	GroundstationAddressMutex  sync.Mutex
	OnboardAddress             string
	OnboardAddressMutex        sync.Mutex
	AntennaTrackerAddress      string
	AntennaTrackerAddressMutex sync.Mutex
	packetCounter              uint
	packetCounterMutex         sync.Mutex
}

func InitializeManager(in_systemid, in_port, in_key, in_groundstationAddress, in_onboardAddress, in_antennaTrackerAddress string, in_handler CommunicationHandler) (*Manager, error) {

	var addr *net.UDPAddr
	var conn *net.UDPConn
	var addrError, connError error

	switch in_systemid {
	case "GS":
		addr, addrError = net.ResolveUDPAddr("udp", in_groundstationAddress)
		if addrError != nil {
			return &Manager{}, addrError
		}
	case "OB":
		addr, addrError = net.ResolveUDPAddr("udp", in_onboardAddress)
		if addrError != nil {
			return &Manager{}, addrError
		}
	case "AT":
		addr, addrError = net.ResolveUDPAddr("udp", in_antennaTrackerAddress)
		if addrError != nil {
			return &Manager{}, addrError
		}
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
	m.KeyMutex.Lock()
	defer m.KeyMutex.Unlock()
	return m.Key
}

func (m *Manager) SetKey(in_key string) {
	m.KeyMutex.Lock()
	defer m.KeyMutex.Unlock()
	m.Key = in_key
}

func (m *Manager) GetGroundstationAddress() string {
	m.GroundstationAddressMutex.Lock()
	defer m.GroundstationAddressMutex.Unlock()
	return m.GroundstationAddress
}

func (m *Manager) SetGroundstationAddress(in_address string) {
	m.GroundstationAddressMutex.Lock()
	defer m.GroundstationAddressMutex.Unlock()
	m.GroundstationAddress = in_address
}

func (m *Manager) GetOnboardAddress() string {
	m.OnboardAddressMutex.Lock()
	defer m.OnboardAddressMutex.Unlock()
	return m.OnboardAddress
}

func (m *Manager) SetOnboardAddress(in_address string) {
	m.OnboardAddressMutex.Lock()
	defer m.OnboardAddressMutex.Unlock()
	m.OnboardAddress = in_address
}

func (m *Manager) GetAntennaTrackerAddress() string {
	m.AntennaTrackerAddressMutex.Lock()
	defer m.AntennaTrackerAddressMutex.Unlock()
	return m.AntennaTrackerAddress
}

func (m *Manager) SetAntennaTrackerAddress(in_address string) {
	m.AntennaTrackerAddressMutex.Lock()
	defer m.AntennaTrackerAddressMutex.Unlock()
	m.AntennaTrackerAddress = in_address
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

			// Decode
			decodeError = json.Unmarshal(decryptedPacket, &packet)
			if decodeError != nil {
				panic(decodeError)
			}

			// Handlers
			go m.MinimalHandler(packet)
			go m.Handler(packet)

		}
	}

}

func (m *Manager) Send2Any(in_message Communication_Message, in_address string) {

	// Connect to the server
	conn, err := net.Dial("udp", in_address)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// encode packet
	var packet = Communication_Packet{SenderID: m.SystemID, Counter: m.IncrementCounter(), Type: in_message.GetType(), SubType: in_message.GetSubType(), Message: in_message.Encode()}
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

func (m *Manager) Send2Onboard(in_message Communication_Message) {

	// Connect to the server
	conn, err := net.Dial("udp", m.GetOnboardAddress())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// encode packet
	var packet = Communication_Packet{SenderID: m.SystemID, Counter: m.IncrementCounter(), Type: in_message.GetType(), SubType: in_message.GetSubType(), Message: in_message.Encode()}
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

func (m *Manager) Send2Groundstation(in_message Communication_Message) {

	// Connect to the server
	conn, err := net.Dial("udp", m.GetGroundstationAddress())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// encode packet
	var packet = Communication_Packet{SenderID: m.SystemID, Counter: m.IncrementCounter(), Type: in_message.GetType(), SubType: in_message.GetSubType(), Message: in_message.Encode()}
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

func (m *Manager) Send2AntennaTracker(in_message Communication_Message) {

	// Connect to the server
	conn, err := net.Dial("udp", m.GetAntennaTrackerAddress())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// encode packet
	var packet = Communication_Packet{SenderID: m.SystemID, Counter: m.IncrementCounter(), Type: in_message.GetType(), SubType: in_message.GetSubType(), Message: in_message.Encode()}
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

func (m *Manager) MinimalHandler(in_packet Communication_Packet) {

	// if PING perform the linking
	if in_packet.Type == 1 {

		var message_ping Communication_Message_Ping
		if err := json.Unmarshal([]byte(in_packet.Message), &message_ping); err != nil {
			panic(err)
		}

		switch in_packet.SenderID {
		case "GS":
			m.SetGroundstationAddress(message_ping.SenderAddress)
		case "OB":
			m.SetGroundstationAddress(message_ping.SenderAddress)
		case "AT":
			m.SetGroundstationAddress(message_ping.SenderAddress)
		default:
		}
	}

	// if not ACK or NACK
	if in_packet.Type != 2 {

		switch in_packet.SenderID {
		case "GS":
			go m.Send2Groundstation(&Communication_Message_ACK{ACKId: in_packet.Counter})
		case "OB":
			go m.Send2Onboard(&Communication_Message_ACK{ACKId: in_packet.Counter})
		case "AT":
			go m.Send2AntennaTracker(&Communication_Message_ACK{ACKId: in_packet.Counter})
		default:
		}

	}
}

func DefaultHandler(in_packet Communication_Packet) {

	//fmt.Printf("%+v\n", in_packet)

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
	SenderID string
	Counter  uint
	Type     uint
	SubType  uint
	Message  string
}

type Communication_Message interface {
	GetType() uint
	GetSubType() uint
	Encode() string
}

// Message - Ping

type Communication_Message_Ping struct {
	Time          uint
	SenderAddress string
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

// CUSTOM PROTOCOL

// navigation data - report

type Communication_Message_NavigationData_Report struct {
	Latitude  float64
	Longitude float64
	Altitude  float64
	VN        float64
	VE        float64
	VD        float64
	IAS       float64
	Heading   float64
	Track     float64
	Roll      float64
	Pitch     float64
	RollRate  float64
	PitchRate float64
}

func (m *Communication_Message_NavigationData_Report) GetType() uint {
	return 3
}
func (m *Communication_Message_NavigationData_Report) GetSubType() uint {
	return 1
}
func (m *Communication_Message_NavigationData_Report) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// visual tracking - report

type Communication_Message_VisualTrackingData_Report struct {
	RelXMin float64
	RelXMax float64
	RelYMin float64
	RelYMax float64
	Active  bool
}

func (m *Communication_Message_VisualTrackingData_Report) GetType() uint {
	return 4
}
func (m *Communication_Message_VisualTrackingData_Report) GetSubType() uint {
	return 1
}
func (m *Communication_Message_VisualTrackingData_Report) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// visual tracking - init

type Communication_Message_VisualTrackingData_Init struct {
	RelX float64
	RelY float64
}

func (m *Communication_Message_VisualTrackingData_Init) GetType() uint {
	return 4
}
func (m *Communication_Message_VisualTrackingData_Init) GetSubType() uint {
	return 2
}
func (m *Communication_Message_VisualTrackingData_Init) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// control mode - report

type Communication_Message_ControlMode_Report struct {
	ControlMode string
}

func (m *Communication_Message_ControlMode_Report) GetType() uint {
	return 5
}
func (m *Communication_Message_ControlMode_Report) GetSubType() uint {
	return 1
}
func (m *Communication_Message_ControlMode_Report) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// control mode - set

type Communication_Message_ControlMode_Set struct {
	ControlMode string
}

func (m *Communication_Message_ControlMode_Set) GetType() uint {
	return 5
}
func (m *Communication_Message_ControlMode_Set) GetSubType() uint {
	return 2
}
func (m *Communication_Message_ControlMode_Set) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}
