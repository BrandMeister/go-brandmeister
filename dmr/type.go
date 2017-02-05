package dmr

// DMR data types
const (
	TypePIHeader = 0x00 + iota
	TypeVoiceHeader
	TypeTerminatorLC
	TypeCSBK
	TypeMBCHeader
	TypeMBCContinuation
	TypeDataHeader
	TypeRate12Data
	TypeRate34Data
	TypeIdle
	/* pseudo types */
	TypeVoiceFrameA
	TypeVoiceFrameB
	TypeVoiceFrameC
	TypeVoiceFrameD
	TypeVoiceFrameE
	TypeVoiceFrameF
	TypeEmbeddedData = 0x11
)
