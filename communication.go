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
	"time"
)

const CONNECTION_THRESHOLD int64 = 1000

type CommunicationHandler func(Communication_Packet)

type Manager struct {
	SystemID                    string
	SystemType                  string
	Address                     *net.UDPAddr
	Connection                  *net.UDPConn
	Initialized                 bool
	RXHandler                   func(Communication_Packet)
	TXHandler                   func(Communication_Packet)
	Key                         string
	GroundstationID             string
	GroundstationAddress        string
	GroundstationLastTimestamp  int64
	OnboardID                   string
	OnboardAddress              string
	OnboardLastTimestamp        int64
	AntennaTrackerID            string
	AntennaTrackerAddress       string
	AntennaTrackerLastTimestamp int64
	packetCounter               uint
	Mutex                       sync.Mutex
}

func NewCommunicationManager(in_systemID, in_systemType, in_key, in_groundstationAddress, in_onboardAddress, in_antennaTrackerAddress string, in_RXhandler, in_TXHandler CommunicationHandler) *Manager {

	m := &Manager{
		SystemID:              in_systemID,
		SystemType:            in_systemType,
		GroundstationAddress:  in_groundstationAddress,
		OnboardAddress:        in_onboardAddress,
		AntennaTrackerAddress: in_antennaTrackerAddress,
		RXHandler:             in_RXhandler,
		TXHandler:             in_TXHandler,
	}

	switch m.SystemType {
	case "GS":
		m.SetGroundstationID(in_systemID)
	case "OB":
		m.SetOnboardID(in_systemID)
	case "AT":
		m.SetAntennaTrackerID(in_systemID)
	}

	if in_key == "" {
		m.Key = COMMUNICATION_DEFAULT_KEY
	} else {
		m.Key = in_key
	}

	return m

}

func (m *Manager) Initialize() {

	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	var addr *net.UDPAddr
	var conn *net.UDPConn
	var addrError, connError error

	switch m.SystemType {
	case "GS":
		addr, addrError = net.ResolveUDPAddr("udp", m.GroundstationAddress)
		if addrError != nil {
			if connError != nil {
				m.Initialized = false
				return
			} else {
				m.Address = addr
			}
		}
	case "OB":
		addr, addrError = net.ResolveUDPAddr("udp", m.OnboardAddress)
		if addrError != nil {
			if connError != nil {
				m.Initialized = false
				return
			} else {
				m.Address = addr
			}
		}
	case "AT":
		addr, addrError = net.ResolveUDPAddr("udp", m.AntennaTrackerAddress)
		if addrError != nil {
			if connError != nil {
				m.Initialized = false
				return
			} else {
				m.Address = addr
			}
		}
	}

	conn, connError = net.ListenUDP("udp", addr)
	if connError != nil {
		m.Initialized = false
		return
	} else {
		m.Initialized = true
		m.Connection = conn
	}

}

func (m *Manager) IsInitialized() bool {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	return m.Initialized
}

func (m *Manager) GetCounter() uint {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	return m.packetCounter
}

func (m *Manager) GetSystemID() string {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	return m.SystemID
}

func (m *Manager) GetSystemType() string {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	return m.SystemType
}

func (m *Manager) IncrementCounter() uint {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	m.packetCounter = m.packetCounter + 1
	return m.packetCounter - 1

}

func (m *Manager) GetKey() string {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	return m.Key
}

func (m *Manager) SetKey(in_key string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.Key = in_key
}

func (m *Manager) GetGroundstationID() string {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	return m.GroundstationID
}

func (m *Manager) SetGroundstationID(in_ID string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.GroundstationID = in_ID
}

func (m *Manager) GetGroundstationAddress() string {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	return m.GroundstationAddress
}

func (m *Manager) SetGroundstationAddress(in_address string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.GroundstationAddress = in_address
}

func (m *Manager) UpdateGroundstationTimestamp() {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.GroundstationLastTimestamp = time.Now().UnixMilli()
}

func (m *Manager) GroundstationConnected() bool {

	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	now := time.Now().UnixMilli()

	if now-m.GroundstationLastTimestamp > CONNECTION_THRESHOLD {
		return false
	} else {
		return true
	}

}

func (m *Manager) GetOnboardID() string {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	return m.OnboardID
}

func (m *Manager) SetOnboardID(in_ID string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.OnboardID = in_ID
}

func (m *Manager) GetOnboardAddress() string {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	return m.OnboardAddress
}

func (m *Manager) SetOnboardAddress(in_address string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.OnboardAddress = in_address
}

func (m *Manager) UpdateOnboardTimestamp() {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.OnboardLastTimestamp = time.Now().UnixMilli()
}

func (m *Manager) OnboardConnected() bool {

	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	now := time.Now().UnixMilli()

	if now-m.OnboardLastTimestamp > CONNECTION_THRESHOLD {
		return false
	} else {
		return true
	}

}

func (m *Manager) GetAntennaTrackerID() string {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	return m.AntennaTrackerID
}

func (m *Manager) SetAntennaTrackerID(in_ID string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.AntennaTrackerID = in_ID
}

func (m *Manager) GetAntennaTrackerAddress() string {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	return m.AntennaTrackerAddress
}

func (m *Manager) SetAntennaTrackerAddress(in_address string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.AntennaTrackerAddress = in_address
}

func (m *Manager) UpdateAntennaTrackerTimestamp() {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.AntennaTrackerLastTimestamp = time.Now().UnixMilli()
}

func (m *Manager) AntennaTrackerConnected() bool {

	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	now := time.Now().UnixMilli()

	if now-m.AntennaTrackerLastTimestamp > CONNECTION_THRESHOLD {
		return false
	} else {
		return true
	}

}

func (m *Manager) Run() {

	buffer := make([]byte, 4096)

	var packet Communication_Packet
	var decodeError error

	// try to initialize until the connection is established

	if !m.IsInitialized() {
		for {
			m.Initialize()

			if m.IsInitialized() {
				break
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	}

	// receive loop

	for {

		n, _, err := m.Connection.ReadFromUDP(buffer)
		if err != nil {

			// if connection is not established, try to reinitialzie until it is established

			m.Connection.Close()

			for {
				m.Initialize()

				if m.IsInitialized() {
					break
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}

		receivedPacketWithMAC := buffer[:n]
		receivedMAC := receivedPacketWithMAC[:32]
		receivedEncryptedPacket := receivedPacketWithMAC[32:]

		// Verify MAC
		if verifyMAC([]byte(m.Key), receivedMAC, receivedEncryptedPacket) {

			// Decrypt the received data
			decryptedPacket, decryptErr := decrypt([]byte(m.Key), receivedEncryptedPacket)
			if decryptErr != nil {

				// if could not decrypt, skip the received data and wait for next

				continue
			}

			// Decode
			decodeError = json.Unmarshal(decryptedPacket, &packet)
			if decodeError != nil {

				// if could not decrypt, skip the received data and wait for next

				continue
			}

			// Handlers
			go m.MinimalRXHandler(packet)
			go m.RXHandler(packet)

		}
	}

}

func (m *Manager) Send2Any(in_message Communication_Message, in_address string, in_requestACK bool) {

	// Connect to the server
	conn, err := net.Dial("udp", in_address)
	if err != nil {

		// if connection is not established, return

		return

	}
	defer conn.Close()

	// encode packet
	var packet = Communication_Packet{SenderID: m.SystemID, SenderType: m.SystemType, Counter: m.IncrementCounter(), RequestACK: in_requestACK, Type: in_message.GetType(), SubType: in_message.GetSubType(), Message: in_message.Encode()}
	encodedPacket, err := json.Marshal(packet)
	if err != nil {

		// if encoding fails, return

		return

	}

	// handler
	m.TXHandler(packet)

	// Encrypt the Gob-encoded message
	encryptedData, err := encrypt([]byte(m.Key), encodedPacket)
	if err != nil {

		// if encryption fails, return

		return

	}

	// Calculate MAC
	mac := generateMAC([]byte(m.Key), encryptedData)

	// Append MAC to the encrypted message
	encryptedDataWithMAC := append(mac, encryptedData...)

	if _, err := conn.Write(encryptedDataWithMAC); err != nil {

		// if sending fails, return

		return
	}

}

func (m *Manager) Send2Onboard(in_message Communication_Message, in_requestACK bool) {

	// Connect to the server
	conn, err := net.Dial("udp", m.GetOnboardAddress())
	if err != nil {

		// if connection is not established, return

		return

	}
	defer conn.Close()

	// encode packet
	var packet = Communication_Packet{SenderID: m.SystemID, SenderType: m.SystemType, Counter: m.IncrementCounter(), RequestACK: in_requestACK, Type: in_message.GetType(), SubType: in_message.GetSubType(), Message: in_message.Encode()}
	encodedPacket, err := json.Marshal(packet)
	if err != nil {

		// if encoding fails, return

		return

	}

	// handler
	m.TXHandler(packet)

	// Encrypt the Gob-encoded message
	encryptedData, err := encrypt([]byte(m.Key), encodedPacket)
	if err != nil {

		// if encryption fails, return

		return

	}

	// Calculate MAC
	mac := generateMAC([]byte(m.Key), encryptedData)

	// Append MAC to the encrypted message
	encryptedDataWithMAC := append(mac, encryptedData...)

	if _, err := conn.Write(encryptedDataWithMAC); err != nil {

		// if sending fails, return

		return
	}

}

func (m *Manager) Send2Groundstation(in_message Communication_Message, in_requestACK bool) {

	// Connect to the server
	conn, err := net.Dial("udp", m.GetGroundstationAddress())
	if err != nil {

		// if connection is not established, return

		return

	}
	defer conn.Close()

	// encode packet
	var packet = Communication_Packet{SenderID: m.SystemID, SenderType: m.SystemType, Counter: m.IncrementCounter(), RequestACK: in_requestACK, Type: in_message.GetType(), SubType: in_message.GetSubType(), Message: in_message.Encode()}
	encodedPacket, err := json.Marshal(packet)
	if err != nil {

		// if encoding fails, return

		return

	}

	// handler
	m.TXHandler(packet)

	// Encrypt the Gob-encoded message
	encryptedData, err := encrypt([]byte(m.Key), encodedPacket)
	if err != nil {

		// if encryption fails, return

		return

	}
	// Calculate MAC
	mac := generateMAC([]byte(m.Key), encryptedData)

	// Append MAC to the encrypted message
	encryptedDataWithMAC := append(mac, encryptedData...)

	if _, err := conn.Write(encryptedDataWithMAC); err != nil {

		// if sending fails, return

		return
	}

}

func (m *Manager) Send2AntennaTracker(in_message Communication_Message, in_requestACK bool) {

	// Connect to the server
	conn, err := net.Dial("udp", m.GetAntennaTrackerAddress())
	if err != nil {

		// if connection is not established, return

		return

	}
	defer conn.Close()

	// encode packet
	var packet = Communication_Packet{SenderID: m.SystemID, SenderType: m.SystemType, Counter: m.IncrementCounter(), RequestACK: in_requestACK, Type: in_message.GetType(), SubType: in_message.GetSubType(), Message: in_message.Encode()}
	encodedPacket, err := json.Marshal(packet)
	if err != nil {

		// if encoding fails, return

		return

	}

	// handler
	m.TXHandler(packet)

	// Encrypt the Gob-encoded message
	encryptedData, err := encrypt([]byte(m.Key), encodedPacket)
	if err != nil {

		// if encryption fails, return

		return

	}

	// Calculate MAC
	mac := generateMAC([]byte(m.Key), encryptedData)

	// Append MAC to the encrypted message
	encryptedDataWithMAC := append(mac, encryptedData...)

	if _, err := conn.Write(encryptedDataWithMAC); err != nil {

		// if sending fails, return

		return
	}

}

func (m *Manager) MinimalRXHandler(in_packet Communication_Packet) {

	MessagePing := Communication_Message_Ping{}

	// if PING perform the linking

	if in_packet.Type == MessagePing.GetType() && in_packet.SubType == MessagePing.GetSubType() {

		if err := json.Unmarshal([]byte(in_packet.Message), &MessagePing); err != nil {
			return
		}

		switch in_packet.SenderType {
		case "GS":
			m.SetGroundstationID(in_packet.SenderID)
			m.SetGroundstationAddress(MessagePing.SenderAddress)
			m.UpdateGroundstationTimestamp()
		case "OB":
			m.SetOnboardID(in_packet.SenderID)
			m.SetOnboardAddress(MessagePing.SenderAddress)
			m.UpdateOnboardTimestamp()
		case "AT":
			m.SetAntennaTrackerID(in_packet.SenderID)
			m.SetAntennaTrackerAddress(MessagePing.SenderAddress)
			m.UpdateAntennaTrackerTimestamp()
		default:
		}
	}

	MessageACK := Communication_Message_ACK{}
	MessageNACK := Communication_Message_NACK{}

	// if not ACK or NACK

	if (in_packet.Type != MessageACK.GetType() && in_packet.SubType != MessageACK.GetSubType()) || (in_packet.Type != MessageNACK.GetType() && in_packet.SubType != MessageNACK.GetSubType()) {

		if in_packet.RequestACK {

			switch in_packet.SenderType {
			case "GS":
				go m.Send2Groundstation(&Communication_Message_ACK{ACKId: in_packet.Counter}, false)
			case "OB":
				go m.Send2Onboard(&Communication_Message_ACK{ACKId: in_packet.Counter}, false)
			case "AT":
				go m.Send2AntennaTracker(&Communication_Message_ACK{ACKId: in_packet.Counter}, false)
			default:
			}
		}

	}
}

func EmptyHandler(in_packet Communication_Packet) {

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
	SenderID   string
	SenderType string
	Counter    uint
	RequestACK bool
	Type       uint
	SubType    uint
	Message    string
}

type Communication_Message interface {
	GetType() uint
	GetSubType() uint
	Encode() string
}
