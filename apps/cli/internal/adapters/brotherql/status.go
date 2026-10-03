package brotherql

import (
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
)

const (
	statusReply        byte = 0
	statusComplete     byte = 1
	statusError        byte = 2
	statusOff          byte = 4
	statusNotification byte = 5
	statusPhase        byte = 6
)

type status struct {
	kind, phase byte
	reason      printing.Reason
}

func parseStatus(packet []byte) (status, error) {
	if len(packet) != 32 || packet[0] != 0x80 || packet[1] != 0x20 || packet[2] != 0x42 || packet[3] != 0x34 || packet[4] != 0x38 {
		return status{}, errors.New("invalid QL-800 status frame")
	}
	if packet[19] > 1 || packet[20] != 0 || packet[21] != 0 {
		return status{}, errors.New("invalid QL-800 phase")
	}
	s := status{kind: packet[18], phase: packet[19]}
	switch s.kind {
	case statusReply, statusComplete, statusError, statusOff, statusNotification, statusPhase:
	default:
		return status{}, errors.New("unrecognized QL-800 status type")
	}
	switch {
	case packet[8]&0x20 != 0 || s.kind == statusOff:
		s.reason = printing.Disconnected
	case packet[8]&0x03 != 0:
		s.reason = printing.NoMedia
	case packet[9]&0x10 != 0:
		s.reason = printing.CoverOpen
	case packet[8]&0x04 != 0:
		s.reason = printing.CutterJam
	case packet[9]&0x41 != 0:
		s.reason = printing.MediaError
	case packet[8] != 0 || packet[9] != 0 || s.kind == statusError:
		s.reason = printing.DeviceError
	}
	return s, nil
}
