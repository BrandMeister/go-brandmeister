package homebrew

import "time"

// Defaults
const (
	DefaultPort              = 62030
	DefaultKeepAliveInterval = time.Second * 5
	DefaultTimeout           = time.Second * 15
	DefaultFirebaseTokenID   = 393838907279
	DefaultFirebaseToken     = "bk3RNwTe3H0:CI2k_HHwgIpoDKCIZvvDMExUdFQ3P1"
)

// Message types
const (
	SignDMRData           = "DMRD"
	SignRepeaterConfig    = "RPTC"
	SignRepeaterLogin     = "RPTL"
	SignRepeaterKey       = "RPTK"
	SignRepeaterPong      = "RPTPONG"
	SignRepeaterClose     = "RPTCL"
	SignRepeaterRSSI      = "RPTRSSI"
	SignRepeaterInterrupt = "RPTINTR"
	SignRepeaterIdle      = "RPTIDLE"
	SignRepeaterWake      = "RPTWAKE"
	SignMaster            = "MST"
	SignMasterACK         = "MSTACK"
	SignMasterNAK         = "MSTNAK"
	SignMasterPing        = "MSTPING"
	SignMasterClose       = "MSTCL"
)
