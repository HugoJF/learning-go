package valve

var (
	Header   = []byte{0x31}
	Response = []byte{0xff, 0xff, 0xff, 0xff, 0x66, 0x0a}
	Regions  = map[string]uint8{
		"UsEastCoast":  0x00,
		"UsWestCoast":  0x01,
		"SouthAmerica": 0x02,
		"Europe":       0x03,
		"Asia":         0x04,
		"Australia":    0x05,
		"MiddleEast":   0x06,
		"Africa":       0x07,
		"RestOfWorld":  0xff,
	}
)
