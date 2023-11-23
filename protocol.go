package communication

import "encoding/json"

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
	Initialized             bool
	Hash                    string
	Points                  []Communication_FlightPlanPoint
	ActivePointIndex        int
	Autoproceed             bool
	LNAVNavigationTrack     float64
	LNAVNavigationtDistance float64
	LNAVDirectTrack         float64
	LNAVDirectDistance      float64
	LNAVApproachingTrack    float64
	LNAVETASeconds          float64
	LNAVError               float64
	LNAVMode                uint
	VNAVTargetAltitude      float64
	VNAVError               float64
	VNAVMode                uint
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

type Communication_Message_Guidance_FirstPoint struct {
}

func (m *Communication_Message_Guidance_FirstPoint) GetType() uint {
	return 11
}
func (m *Communication_Message_Guidance_FirstPoint) GetSubType() uint {
	return 4
}
func (m *Communication_Message_Guidance_FirstPoint) Encode() string {
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
	HRES        float64
	VRES        float64
	HFOV        float64
	VFOV        float64
	OffsetRoll  float64
	OffsetPitch float64
	OffsetYaw   float64
	Fx          float64
	Fy          float64
	Cx          float64
	Cy          float64
	K1          float64
	K2          float64
	K3          float64
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
