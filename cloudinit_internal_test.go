package cloudinit

import (
	"testing"
	"time"
)

// TestDirEntryLen_OddPadding verifies that an even-length fileID (e.g. "ABCD")
// triggers the padding-to-even branch in dirEntryLen. The standard NoCloud
// payload only uses odd-length IDs ("\x00", "META-DATA;1") so the typical ISO
// build never exercises this branch.
func TestDirEntryLen_OddPadding(t *testing.T) {
	// len("ABCD")=4 → 33+4=37 (odd) → padded to 38.
	if got := dirEntryLen("ABCD"); got != 38 {
		t.Errorf("dirEntryLen(\"ABCD\") = %d, want 38", got)
	}
	// len("META-DATA;1")=11 → 33+11=44 (even) → no padding.
	if got := dirEntryLen("META-DATA;1"); got != 44 {
		t.Errorf("dirEntryLen(\"META-DATA;1\") = %d, want 44", got)
	}
}

// TestWriteDirRecord_OddPadding verifies the padding branch in writeDirRecord.
func TestWriteDirRecord_OddPadding(t *testing.T) {
	buf := make([]byte, 64)
	// len("AB")=2 → recLen=35 (odd) → padded to 36.
	n := writeDirRecord(buf, "AB", 1, 1, false, time.Unix(0, 0).UTC())
	if n != 36 {
		t.Errorf("writeDirRecord padded recLen = %d, want 36", n)
	}
	if buf[0] != 36 {
		t.Errorf("record length byte = %d, want 36", buf[0])
	}
}
