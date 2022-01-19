package ipmi

import "fmt"

// 28.2 Get Chassis Status Command
type GetChassisStatusRequest struct {
	// no data
}

type GetChassisStatusResponse struct {
	// Current Power State
	PowerRestorePolicy PowerRestorePolicy
	PowerControlFault  bool // Controller attempted to turn system power on or off, but system did not enter desired state.
	PowerFault         bool // fault detected in main power subsystem
	InterLock          bool // chassis is presently shut down because a chassis	panel interlock switch is active
	PowerOverload      bool // system shutdown because of power overload condition.
	PowerIsOn          bool

	// Last Power Event
	LastPowerOnByCommand                   bool
	LastPowerDownByPowerFault              bool
	LastPowerDownByPowerInterlockActivated bool
	LastPowerDownByPowerOverload           bool
	ACFailed                               bool

	// Last Power Event

	// Misc. Chassis State
	CollingFanFault         bool
	DriveFault              bool
	FrontPanelLockoutActive bool // (power off and reset via chassis push-buttons disabled. 前面板锁定)
	ChassisIntrusionActive  bool

	// Front Panel Button Capabilities and disable/enable status (Optional)
	SleepButtonDisableAllowed      bool
	DiagnosticButtonDisableAllowed bool
	ResetButtonDisableAllowed      bool
	PoweroffButtonDisableAllowed   bool
	SleepButtonDisabled            bool
	DiagnosticButtonDisabled       bool
	ResetButtonDisabled            bool
	PoweroffButtonDisabled         bool
}

// PowerRestorePolicy
// 通电开机策略
type PowerRestorePolicy uint8

const (
	PowerRestorePolicyAlwaysOff PowerRestorePolicy = 0 // 保持下电
	PowerRestorePolicyPrevious  PowerRestorePolicy = 1 // 与之前保持一致
	PowerRestorePolicyAlwaysOn  PowerRestorePolicy = 2 // 保持上电
)

var SupportedPowerRestorePolicies = []string{
	"always-off", "always-on", "previous",
}

func (p PowerRestorePolicy) String() string {
	switch p {
	case 0:
		// chassis stays powered off after AC/mains returns
		return "always-off"
	case 1:
		// after AC returns, power is restored to the state that was in effect when AC/mains was lost
		return "previous"
	case 2:
		// chassis always powers up after AC/mains returns
		return "always-on"
	case 3:
		return "unkown"
	}
	return "unkown"
}

func (req *GetChassisStatusRequest) Pack() []byte {
	return []byte{}
}

func (req *GetChassisStatusRequest) Command() Command {
	return CommandGetChassisStatus
}

func (res *GetChassisStatusResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetChassisStatusResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return ErrUnpackedDataTooShort
	}

	b1, _, _ := unpackUint8(msg, 0)
	// first clear bit 7, then shift right 5 bits
	b := (b1 & 0x7f) >> 5
	res.PowerRestorePolicy = PowerRestorePolicy(b)
	res.PowerControlFault = isBit4Set(b1)
	res.PowerFault = isBit3Set(b1)
	res.InterLock = isBit2Set(b1)
	res.PowerOverload = isBit1Set(b1)
	res.PowerIsOn = isBit0Set(b1)

	b2, _, _ := unpackUint8(msg, 1)
	res.LastPowerOnByCommand = isBit4Set(b2)
	res.LastPowerDownByPowerFault = isBit3Set(b2)
	res.LastPowerDownByPowerInterlockActivated = isBit2Set(b2)
	res.LastPowerDownByPowerOverload = isBit1Set(b2)
	res.ACFailed = isBit0Set(b2)

	b3, _, _ := unpackUint8(msg, 2)
	res.CollingFanFault = isBit3Set(b3)
	res.DriveFault = isBit2Set(b3)
	res.FrontPanelLockoutActive = isBit1Set(b3)
	res.ChassisIntrusionActive = isBit0Set(b3)

	if len(msg) == 4 {
		b4, _, _ := unpackUint8(msg, 3)
		res.SleepButtonDisableAllowed = isBit7Set(b4)
		res.DiagnosticButtonDisableAllowed = isBit6Set(b4)
		res.ResetButtonDisableAllowed = isBit5Set(b4)
		res.PoweroffButtonDisableAllowed = isBit4Set(b4)
		res.SleepButtonDisabled = isBit3Set(b4)
		res.DiagnosticButtonDisabled = isBit2Set(b4)
		res.ResetButtonDisabled = isBit1Set(b4)
		res.PoweroffButtonDisabled = isBit0Set(b4)
	}
	return nil
}

func (res *GetChassisStatusResponse) Format() string {
	return "System Power         : " + formatBool(res.PowerIsOn, "on", "off") + "\n" +
		"Power Overload       : " + fmt.Sprintf("%v", res.PowerOverload) + "\n" +
		"Power Interlock      : " + formatBool(res.InterLock, "active", "inactive") + "\n" +
		"Main Power Fault     : " + fmt.Sprintf("%v", res.PowerFault) + "\n" +
		"Power Control Fault  : " + fmt.Sprintf("%v", res.PowerControlFault) + "\n" +
		"Power Restore Policy : " + res.PowerRestorePolicy.String() + "\n" +
		"Last Power Event     : " + formatBool(res.ChassisIntrusionActive, "active", "inactive") + "\n" +
		"Chassis Intrusion    : " + formatBool(res.ChassisIntrusionActive, "active", "inactive") + "\n" +
		"Front-Panel Lockout  : " + formatBool(res.FrontPanelLockoutActive, "active", "inactive") + "\n" +
		"Drive Fault          : " + fmt.Sprintf("%v", res.DriveFault) + "\n" +
		"Cooling/Fan Fault    : " + fmt.Sprintf("%v", res.CollingFanFault) + "\n" +
		"Sleep Button Disable : " + formatBool(res.SleepButtonDisableAllowed, "allowed", "disallowed") + "\n" +
		"Diag Button Disable  : " + formatBool(res.DiagnosticButtonDisableAllowed, "allowed", "disallowed") + "\n" +
		"Reset Button Disable : " + formatBool(res.ResetButtonDisableAllowed, "allowed", "disallowed") + "\n" +
		"Power Button Disable : " + formatBool(res.PoweroffButtonDisableAllowed, "allowed", "disallowed") + "\n" +
		"Sleep Button Disabled: " + fmt.Sprintf("%v", res.SleepButtonDisabled) + "\n" +
		"Diag Button Disabled : " + fmt.Sprintf("%v", res.DiagnosticButtonDisabled) + "\n" +
		"Reset Button Disabled: " + fmt.Sprintf("%v", res.ResetButtonDisabled) + "\n" +
		"Power Button Disabled: " + fmt.Sprintf("%v", res.PoweroffButtonDisabled)
}

func (c *Client) GetChassisStatus() (response *GetChassisStatusResponse, err error) {
	request := &GetChassisStatusRequest{}
	response = &GetChassisStatusResponse{}
	err = c.Exchange(request, response)
	return
}
