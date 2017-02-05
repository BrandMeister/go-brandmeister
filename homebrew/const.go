package homebrew

import "time"

// Defaults
const (
	DefaultPort              = 62030
	DefaultKeepAliveInterval = time.Second * 5
	DefaultTimeout           = time.Second * 15
)

// Message types
const (
	SignDMRData        = "DMRD"
	SignRepeaterConfig = "RPTC"
	SignRepeaterLogin  = "RPTL"
	SignRepeaterKey    = "RPTK"
	SignRepeaterPong   = "RPTPONG"
	SignRepeaterClose  = "RPTCL"
	SignMaster         = "MST"
	SignMasterACK      = "MSTACK"
	SignMasterNAK      = "MSTNAK"
	SignMasterPing     = "MSTPING"
	SignMasterClose    = "MSTCL"
)
