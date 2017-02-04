package rewind

func marshalBytes(dst, src []byte, offset *int) {
	*offset += copy(dst, src)
}

func marshalUint16(dst []byte, u uint16, offset *int) {
	dst[0] = byte(u)
	dst[1] = byte(u >> 8)
	*offset += 2
}

func marshalUint32(dst []byte, u uint32, offset *int) {
	dst[0] = byte(u)
	dst[1] = byte(u >> 8)
	dst[2] = byte(u >> 16)
	dst[3] = byte(u >> 24)
	*offset += 4
}
