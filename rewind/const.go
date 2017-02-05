package rewind

import "time"

// Defaults
const (
	Sign                     = "REWIND01"
	SignLength               = len(Sign)
	CallLength               = 10
	DefaultKeepAliveInterval = time.Second * 5
	DefaultTimeout           = time.Second * 15
	DefaultPort              = 54005
)

// Packet classes
const (
	ClassRewindControl = 0x0000
	ClassSystemConsole = 0x0100
	ClassServiceNotice = 0x0200
	ClassDeviceData    = 0x0800
	ClassApplication   = 0x0900
)

// Control types
const (
	TypeKeepAlive = ClassRewindControl + iota
	TypeClose
	TypeChallenge
	TypeAuthentication
)

// System Console types
const (
	TypeReport = ClassSystemConsole + iota
)

// Service Notice types
const (
	TypeBusyNotice = ClassServiceNotice + iota
	TypeAddressNotice
	TypeBindingNotice
)

// Device Data types for various vendors
const (
	// Kairos
	ClassKairosData = ClassDeviceData + 0x00
	ClassHyteraData = ClassDeviceData + 0x10
)

// Kairos Data types
const (
	TypeKairosExternalServer = ClassKairosData + iota
	TypeKairosRemoteControl
	TypeKairosSNMPTrap
)

// Hytera Data types
const (
	TypeHyteraPeerData = ClassHyteraData + iota
	TypeHyteraRDACData
	TypeHyteraMediaData
)

// Application types
const (
	TypeConfiguration = ClassApplication + iota
	TypeSubscription
	TypeDMRDataBase     = ClassApplication + 0x10
	TypeDMRAudioBase    = ClassApplication + 0x20
	TypeDMREmbeddedData = ClassApplication + 0x27
	TypeSuperHeader     = ClassApplication + 0x28
	TypeFailureCode     = ClassApplication + 0x29
)

// Flags
const (
	FlagNone      = 0
	FlagRealTime1 = 1 << iota
	FlagRealTime2
	FlagDefaultSet = FlagNone
)

// Roles
const (
	RoleRepeaterAgent = 0x10
	RoleApplication   = 0x20
)

// Services
const (
	ServiceCronosAgent       = RoleRepeaterAgent + 0
	ServiceTellusAgent       = RoleRepeaterAgent + 1
	ServiceSimpleApplication = RoleApplication + 0
)

// Options
const (
	// OptionSuperHeader enables sending of super header metadata
	OptionSuperHeader = 1 << iota

	// OptionLinearFrame enables sending of linear coded AMBE without FEC
	OptionLinearFrame
)

// Session types
const (
	PrivateVoice SessionType = 5
	GroupVoice   SessionType = 7
)
