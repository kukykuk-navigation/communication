package communication

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
