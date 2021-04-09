module example.com/packet-counter

go 1.16

replace github.com/lukaslueg/dumpcap => ../../dumpcap

require (
	github.com/google/gopacket v1.1.19
	github.com/lukaslueg/dumpcap v0.0.0-00010101000000-000000000000
	github.com/manifoldco/promptui v0.8.0
)
