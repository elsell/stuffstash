package presentation

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func (o Output) pairingReview(v ports.PairingReview) error {
	fields := [][2]string{{"Pairing", v.ID}, {"Name", v.Name}, {"Fingerprint", v.PublicKeyFingerprint}, {"Rotation", strconv.FormatBool(v.Rotation)}}
	for _, c := range v.Candidates {
		b, err := json.Marshal(c)
		if err != nil {
			return err
		}
		fields = append(fields, [2]string{"Candidate", string(b)})
	}
	return o.details(fields)
}
func (o Output) approvedConnector(v ports.ApprovedConnector) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return o.details([][2]string{{"Connector", v.ID}, {"Name", v.Name}, {"Generation", strconv.FormatUint(v.Generation, 10)}, {"Details", string(b)}})
}
