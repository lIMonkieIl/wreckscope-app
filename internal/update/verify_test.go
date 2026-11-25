package update

import (
	"crypto/sha256"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/updater"
)

func TestRequireSHA256FailClosed(t *testing.T) {
	sum := sha256.Sum256([]byte("payload"))
	if err := requireSHA256(nil, sum[:], "setup.exe"); err == nil {
		t.Fatal("nil release must fail")
	}
	if err := requireSHA256(&updater.Release{}, sum[:], "setup.exe"); err == nil {
		t.Fatal("missing Verification must fail")
	}
	if err := requireSHA256(&updater.Release{Verification: &updater.Verification{}}, sum[:], "setup.exe"); err == nil {
		t.Fatal("empty digest must fail")
	}
	bad := append([]byte(nil), sum[:]...)
	bad[0] ^= 0xff
	rel := &updater.Release{Verification: &updater.Verification{Digest: bad}}
	if err := requireSHA256(rel, sum[:], "setup.exe"); err == nil {
		t.Fatal("mismatch must fail")
	}
	rel.Verification.Digest = sum[:]
	if err := requireSHA256(rel, sum[:], "setup.exe"); err != nil {
		t.Fatal(err)
	}
}
