package printworker

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image/png"

	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
)

func (w *Worker) validateClaim(claim printing.Claim, requested printing.AttemptControl) error {
	if claim.Control.AttemptID != requested.AttemptID || claim.Control.SessionID != requested.SessionID || claim.Control.ClaimToken != requested.ClaimToken || claim.Control.Revision == 0 || claim.PrinterID != w.Config.PrinterID || claim.JobID == "" || claim.ContractVersion != 1 || claim.Copies < 1 || claim.Media != w.Config.Media {
		return errors.New("claimed print job does not match this worker")
	}
	if !claim.LeaseExpiresAt.After(w.Clock.Now().Add(w.Config.LeaseSafety)) {
		return errors.New("claimed print lease is not current")
	}
	a := claim.Artifact
	if a.ContentType != "image/png" || a.ByteLength <= 0 || a.ByteLength > w.Config.MaxArtifactBytes || a.Width != claim.Media.RasterWidth || a.Height != claim.Media.RasterHeight || a.Width <= 0 || a.Height <= 0 {
		return errors.New("unsupported print artifact")
	}
	digest, err := hex.DecodeString(a.SHA256)
	if err != nil || len(digest) != sha256.Size || hex.EncodeToString(digest) != a.SHA256 {
		return errors.New("invalid print artifact digest")
	}
	return nil
}
func validateArtifact(claim printing.Claim, body []byte, contentType string) error {
	if contentType != claim.Artifact.ContentType || int64(len(body)) != claim.Artifact.ByteLength {
		return errors.New("print artifact content type or length mismatch")
	}
	digest := sha256.Sum256(body)
	if hex.EncodeToString(digest[:]) != claim.Artifact.SHA256 {
		return errors.New("print artifact integrity check failed")
	}
	config, err := png.DecodeConfig(bytes.NewReader(body))
	if err != nil || config.Width != claim.Artifact.Width || config.Height != claim.Artifact.Height {
		return errors.New("print artifact dimensions do not match its profile")
	}
	return nil
}
