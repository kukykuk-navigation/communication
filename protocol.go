package communication

import "encoding/json"

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
