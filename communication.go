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

const (
	COMMUNICATION_DEFAULT_KEY = "0123456789abcdef"
)

type CommunicationHandler func(Communication_Packet)

type Manager struct {
	SystemID              string
	Address               *net.UDPAddr
	Connection            *net.UDPConn
	Initialized           bool
	Mutex                 sync.Mutex
	RXHandler             func(Communication_Packet)
	TXHandler             func(Communication_Packet)
	Key                   string
	GroundstationAddress  string
	OnboardAddress        string
	AntennaTrackerAddress string
	packetCounter         uint
}

func NewCommunicationManager(in_systemid, in_key, in_groundstationAddress, in_onboardAddress, in_antennaTrackerAddress string, in_RXhandler, in_TXHandler CommunicationHandler) *Manager {

	m := &Manager{
		SystemID:              in_systemid,
		GroundstationAddress:  in_groundstationAddress,
		OnboardAddress:        in_onboardAddress,
		AntennaTrackerAddress: in_antennaTrackerAddress,
		RXHandler:             in_RXhandler,
		TXHandler:             in_TXHandler,
	}

	if in_key == "" {
		m.Key = COMMUNICATION_DEFAULT_KEY
	} else {
		m.Key = in_key
	}

	return m

}

func (m *Manager) Initialze() {

	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	var addr *net.UDPAddr
	var conn *net.UDPConn
	var addrError, connError error

	switch m.SystemID {
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

func (m *Manager) isInitialized() bool {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	return m.Initialized
}

func (m *Manager) getCounter() uint {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	return m.packetCounter
}

func (m *Manager) getSystemID() string {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	return m.SystemID
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

func (m *Manager) Run() {

	buffer := make([]byte, 4096)

	var packet Communication_Packet
	var decodeError error

	// try to initialize until the connection is established

	if !m.isInitialized() {
		for {
			m.Initialze()

			if m.isInitialized() {
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
				m.Initialze()

				if m.isInitialized() {
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
	var packet = Communication_Packet{SenderID: m.SystemID, Counter: m.IncrementCounter(), RequestACK: in_requestACK, Type: in_message.GetType(), SubType: in_message.GetSubType(), Message: in_message.Encode()}
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
	var packet = Communication_Packet{SenderID: m.SystemID, Counter: m.IncrementCounter(), RequestACK: in_requestACK, Type: in_message.GetType(), SubType: in_message.GetSubType(), Message: in_message.Encode()}
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
	var packet = Communication_Packet{SenderID: m.SystemID, Counter: m.IncrementCounter(), RequestACK: in_requestACK, Type: in_message.GetType(), SubType: in_message.GetSubType(), Message: in_message.Encode()}
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
	var packet = Communication_Packet{SenderID: m.SystemID, Counter: m.IncrementCounter(), RequestACK: in_requestACK, Type: in_message.GetType(), SubType: in_message.GetSubType(), Message: in_message.Encode()}
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

		switch in_packet.SenderID {
		case "GS":
			m.SetGroundstationAddress(MessagePing.SenderAddress)
		case "OB":
			m.SetOnboardAddress(MessagePing.SenderAddress)
		case "AT":
			m.SetAntennaTrackerAddress(MessagePing.SenderAddress)
		default:
		}
	}

	MessageACK := Communication_Message_ACK{}
	MessageNACK := Communication_Message_ACK{}

	// if not ACK or NACK

	if (in_packet.Type != MessageACK.GetType() && in_packet.SubType != MessageACK.GetSubType()) || (in_packet.Type != MessageNACK.GetType() && in_packet.SubType != MessageNACK.GetSubType()) {

		if in_packet.RequestACK {

			switch in_packet.SenderID {
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

// visual tracking - stop

type Communication_Message_VisualTrackingData_Stop struct {
	RelX float64
	RelY float64
}

func (m *Communication_Message_VisualTrackingData_Stop) GetType() uint {
	return 4
}
func (m *Communication_Message_VisualTrackingData_Stop) GetSubType() uint {
	return 3
}
func (m *Communication_Message_VisualTrackingData_Stop) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// control mode - report

type Communication_Message_Control_Report struct {
	ControlMode         uint
	ControlManualMode   uint
	ManualInputValid    bool
	AutopilotInputValid bool
}

func (m *Communication_Message_Control_Report) GetType() uint {
	return 5
}
func (m *Communication_Message_Control_Report) GetSubType() uint {
	return 1
}
func (m *Communication_Message_Control_Report) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// control mode - set

type Communication_Message_ControlMode_Set struct {
	ControlMode uint
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

// control manual mode - set

type Communication_Message_ControlManualMode_Set struct {
	ControlManualMode uint
}

func (m *Communication_Message_ControlManualMode_Set) GetType() uint {
	return 5
}
func (m *Communication_Message_ControlManualMode_Set) GetSubType() uint {
	return 3
}
func (m *Communication_Message_ControlManualMode_Set) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// control manual input - set

type Communication_Message_ManualInput_Set struct {
	RollAxis  float64
	PitchAxis float64
	YawAxis   float64
	PowerAxis float64
}

func (m *Communication_Message_ManualInput_Set) GetType() uint {
	return 5
}
func (m *Communication_Message_ManualInput_Set) GetSubType() uint {
	return 4
}
func (m *Communication_Message_ManualInput_Set) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// control autopilot input - set

type Communication_Message_AutopilotInput_Set struct {
	Altitude float64
	Track    float64
}

func (m *Communication_Message_AutopilotInput_Set) GetType() uint {
	return 5
}
func (m *Communication_Message_AutopilotInput_Set) GetSubType() uint {
	return 5
}
func (m *Communication_Message_AutopilotInput_Set) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// target estimate - report

type Communication_Message_TargetEstimate_Report struct {
	Roll  float64
	Pitch float64
	Yaw   float64
}

func (m *Communication_Message_TargetEstimate_Report) GetType() uint {
	return 6
}
func (m *Communication_Message_TargetEstimate_Report) GetSubType() uint {
	return 1
}
func (m *Communication_Message_TargetEstimate_Report) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// GNSS performance - report

type Communication_Message_GNSSPerformance_Report struct {
	FixType           uint
	Latitude          float64
	Longitude         float64
	Altitude          float64
	Velocity          float64
	Heading           float64
	Track             float64
	HorDOP            float64
	VerDOP            float64
	HorACC            float64
	VerACC            float64
	VelACC            float64
	TrkACC            float64
	SatellitesVisible uint
}

func (m *Communication_Message_GNSSPerformance_Report) GetType() uint {
	return 7
}
func (m *Communication_Message_GNSSPerformance_Report) GetSubType() uint {
	return 1
}
func (m *Communication_Message_GNSSPerformance_Report) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// guidance state - report

type Communication_Message_GuidanceState_Report struct {
	GuidanceMode           string
	GuidanceDistanceToNext float64
	GuidanceTrackToNext    float64
}

func (m *Communication_Message_GuidanceState_Report) GetType() uint {
	return 8
}
func (m *Communication_Message_GuidanceState_Report) GetSubType() uint {
	return 1
}
func (m *Communication_Message_GuidanceState_Report) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// power - report

type Communication_Message_Power_Report struct {
	VoltageBattery float64
}

func (m *Communication_Message_Power_Report) GetType() uint {
	return 9
}
func (m *Communication_Message_Power_Report) GetSubType() uint {
	return 1
}
func (m *Communication_Message_Power_Report) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// flight plan - point

type Communication_FlightPlanPoint struct {
	Label           string
	Latitude_WGS84  float64
	Longitude_WGS84 float64
	Altitude_WGS84  float64
	X_ECEF          float64
	Y_ECEF          float64
	Z_ECEF          float64
}

// flight plan - upload

type Communication_Message_FlightPlan_Upload struct {
	Points []Communication_FlightPlanPoint
}

func (m *Communication_Message_FlightPlan_Upload) GetType() uint {
	return 10
}
func (m *Communication_Message_FlightPlan_Upload) GetSubType() uint {
	return 1
}
func (m *Communication_Message_FlightPlan_Upload) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// flight plan - clear

type Communication_Message_FlightPlan_Clear struct {
}

func (m *Communication_Message_FlightPlan_Clear) GetType() uint {
	return 10
}
func (m *Communication_Message_FlightPlan_Clear) GetSubType() uint {
	return 2
}
func (m *Communication_Message_FlightPlan_Clear) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// guidance - report

type Communication_Message_Guidance_Report struct {
	Initialized      bool
	Hash             string
	Points           []Communication_FlightPlanPoint
	ActivePointIndex int
	Autoproceed      bool
	LNAVTrack        float64
	LNAVOfftrack     float64
	LNAVDistance     float64
}

func (m *Communication_Message_Guidance_Report) GetType() uint {
	return 11
}
func (m *Communication_Message_Guidance_Report) GetSubType() uint {
	return 1
}
func (m *Communication_Message_Guidance_Report) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// guidance - next point

type Communication_Message_Guidance_NextPoint struct {
}

func (m *Communication_Message_Guidance_NextPoint) GetType() uint {
	return 11
}
func (m *Communication_Message_Guidance_NextPoint) GetSubType() uint {
	return 2
}
func (m *Communication_Message_Guidance_NextPoint) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// guidance - previous point

type Communication_Message_Guidance_PreviousPoint struct {
}

func (m *Communication_Message_Guidance_PreviousPoint) GetType() uint {
	return 11
}
func (m *Communication_Message_Guidance_PreviousPoint) GetSubType() uint {
	return 3
}
func (m *Communication_Message_Guidance_PreviousPoint) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// guidance - set autoproceed

type Communication_Message_Guidance_setAutoproceed struct {
	Autoproceed bool
}

func (m *Communication_Message_Guidance_setAutoproceed) GetType() uint {
	return 11
}
func (m *Communication_Message_Guidance_setAutoproceed) GetSubType() uint {
	return 4
}
func (m *Communication_Message_Guidance_setAutoproceed) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// flight controller - report

type Communication_Message_FlightController_Report struct {
	BaseMode   uint8
	CustomMode uint32
}

func (m *Communication_Message_FlightController_Report) GetType() uint {
	return 12
}
func (m *Communication_Message_FlightController_Report) GetSubType() uint {
	return 1
}
func (m *Communication_Message_FlightController_Report) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// onboard systems - report

type Communication_Message_OnboardSystems_Report struct {
	ControlLoop       float64
	TransmitToFC      float64
	ReceiveFromFC     float64
	FrontCameraVideo  float64
	BottomCameraVideo float64
	VisualTracking    float64
	ManualInput       float64
	AutopilotInput    float64
}

func (m *Communication_Message_OnboardSystems_Report) GetType() uint {
	return 13
}
func (m *Communication_Message_OnboardSystems_Report) GetSubType() uint {
	return 1
}
func (m *Communication_Message_OnboardSystems_Report) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

// camera - front - offsets

type Communication_Message_CameraParameters_Report struct {
	ID          uint
	OffsetRoll  float64
	OffsetPitch float64
	OffsetYaw   float64
}

func (m *Communication_Message_CameraParameters_Report) GetType() uint {
	return 14
}
func (m *Communication_Message_CameraParameters_Report) GetSubType() uint {
	return 1
}
func (m *Communication_Message_CameraParameters_Report) Encode() string {
	encoded, _ := json.Marshal(m)
	return string(encoded)
}
