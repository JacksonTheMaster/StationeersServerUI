package connectivity

import (
	"encoding/binary"
	"testing"
)

func TestProbeRoundTripParsing(t *testing.T) {
	nonce := "0123456789abcdef0123456789abcdef"
	packet := make([]byte, len(probeMagic)+len(nonce)+2+1200-len(probeMagic)-len(nonce)-2)
	copy(packet, probeMagic)
	copy(packet[len(probeMagic):], nonce)
	binary.BigEndian.PutUint16(packet[len(probeMagic)+len(nonce):], 1200)

	size, ok := parseProbe(packet, nonce)
	if !ok || size != 1200 {
		t.Fatalf("parseProbe() = (%d, %t), want (1200, true)", size, ok)
	}
	if _, ok := parseProbe(packet, "wrongnoncewrongnoncewrongnonce12"); ok {
		t.Fatal("parseProbe accepted a packet with the wrong nonce")
	}
}

func TestProbeSizeValidation(t *testing.T) {
	for _, size := range []int{1000, 1200, 1400} {
		if !isProbeSize(size) {
			t.Errorf("probe size %d was rejected", size)
		}
	}
	if isProbeSize(1440) {
		t.Fatal("unexpectedly accepted an unconfigured probe size")
	}
}

func TestNormalizeState(t *testing.T) {
	if got := normalizeState("", []ProbeResult{{Size: 1000, Received: true}, {Size: 1200, Received: true}, {Size: 1400, Received: true}}); got != stateReachable {
		t.Fatalf("normalizeState() = %q, want %q", got, stateReachable)
	}
	if got := normalizeState("", []ProbeResult{{Size: 1000, Received: true}, {Size: 1200, Received: false}}); got != statePartial {
		t.Fatalf("normalizeState() = %q, want %q", got, statePartial)
	}
}
