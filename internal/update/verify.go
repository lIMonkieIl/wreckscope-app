package update

import (
	"crypto/subtle"
	"encoding/hex"
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/updater"
)

// requireSHA256 fails closed when SHA256SUMS / Verification is missing or mismatches.
func requireSHA256(rel *updater.Release, sum []byte, name string) error {
	if rel == nil || rel.Verification == nil || len(rel.Verification.Digest) == 0 {
		return fmt.Errorf("updater: missing checksum for %s (SHA256SUMS required)", name)
	}
	if subtle.ConstantTimeCompare(sum, rel.Verification.Digest) != 1 {
		return fmt.Errorf("updater: checksum mismatch for %s (got %s)", name, hex.EncodeToString(sum))
	}
	return nil
}
