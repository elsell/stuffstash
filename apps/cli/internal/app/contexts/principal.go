package contexts

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// Principal identifies an account using claims verified by the authentication
// adapter. It excludes rotating credentials and does not grant authority.
func Principal(session ports.Session) string {
	if session.Issuer == "" || session.Subject == "" {
		return ""
	}
	identity, _ := json.Marshal([2]string{session.Issuer, session.Subject})
	sum := sha256.Sum256(identity)
	return hex.EncodeToString(sum[:])
}
