package tcpip

// Checksum will calculate the checksum of a byte slice.
func Checksum(b []byte) uint16 {
	sum := uint32(0)
	for ; len(b) >= 2; b = b[2:] {
		b0 := uint32(b[0])
		b1 := uint32(b[1])
		sum += (b0 << 8) | b1
	}
	if len(b) > 0 {
		sum += uint32(b[0]) << 8
	}
	for sum > 0xFFFF {
		sum = (sum >> 16) + (sum & 0xFFFF)
	}
	csum := ^uint16(sum)
	if csum == 0 {
		csum = 0xFFFF
	}
	return csum
}
