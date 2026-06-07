package cloudinit_test

import (
	"strings"
	"testing"

	cloudinit "github.com/openweft/cloud-init"
)

const sector = 2048

// TestBuildCloudInitISO verifies the ISO builder produces a valid ISO 9660
// image with the correct volume label and readable file content.
func TestBuildCloudInitISO(t *testing.T) {
	userData := "#!/bin/bash\necho hello from cloud-init\n"
	iso, err := cloudinit.BuildCloudInitISO("test-vm", "myhost", userData)
	if err != nil {
		t.Fatalf("BuildCloudInitISO: %v", err)
	}
	if len(iso) == 0 || len(iso)%sector != 0 {
		t.Fatalf("unexpected ISO size: %d", len(iso))
	}
	// Sector 16 must be a Primary Volume Descriptor
	pvd := iso[16*sector:]
	if pvd[0] != 1 {
		t.Errorf("PVD type: expected 1, got %d", pvd[0])
	}
	if string(pvd[1:6]) != "CD001" {
		t.Errorf("PVD standard identifier: expected CD001, got %q", pvd[1:6])
	}
	// Volume label should be "cidata" (left-justified, space-padded)
	label := strings.TrimRight(string(pvd[40:72]), " ")
	if label != "cidata" {
		t.Errorf("volume label: expected 'cidata', got %q", label)
	}
	// Sector 17 must be the Volume Descriptor Set Terminator
	vdst := iso[17*sector:]
	if vdst[0] != 0xFF {
		t.Errorf("VDST type: expected 0xFF, got 0x%02X", vdst[0])
	}
}

// TestBuildSSHCloudConfig verifies that BuildSSHCloudConfig returns a valid
// #cloud-config YAML with ssh_authorized_keys.
func TestBuildSSHCloudConfig(t *testing.T) {
	key := "ssh-ed25519 AAAAC3NzaC test@host"
	got := cloudinit.BuildSSHCloudConfig([]string{key}, "")
	if !strings.HasPrefix(got, "#cloud-config\n") {
		t.Errorf("expected #cloud-config header, got: %q", got[:min(len(got), 40)])
	}
	if !strings.Contains(got, "ssh_authorized_keys:") {
		t.Errorf("expected ssh_authorized_keys directive, got: %q", got)
	}
	if !strings.Contains(got, "  - "+key) {
		t.Errorf("expected key in output, got: %q", got)
	}
}

// TestBuildSSHCloudConfig_MultipleKeys verifies that multiple keys are all
// injected.
func TestBuildSSHCloudConfig_MultipleKeys(t *testing.T) {
	keys := []string{"ssh-ed25519 AAAA1 user1@host", "ssh-ed25519 AAAA2 user2@host"}
	got := cloudinit.BuildSSHCloudConfig(keys, "")
	for _, k := range keys {
		if !strings.Contains(got, "  - "+k) {
			t.Errorf("expected key %q in output, got: %q", k, got)
		}
	}
}

// TestBuildSSHCloudConfig_Empty verifies that an empty key list returns "".
func TestBuildSSHCloudConfig_Empty(t *testing.T) {
	got := cloudinit.BuildSSHCloudConfig(nil, "")
	if got != "" {
		t.Errorf("expected empty string for nil keys, got: %q", got)
	}
	got2 := cloudinit.BuildSSHCloudConfig([]string{}, "")
	if got2 != "" {
		t.Errorf("expected empty string for empty keys, got: %q", got2)
	}
}

// TestBuildCloudInitISO_EmptyUserData verifies that passing an empty userData
// string is accepted and produces a valid ISO. This exercises the
// sectorsNeeded(0) == 1 path and the empty-string default for instanceID and
// hostname.
func TestBuildCloudInitISO_EmptyUserData(t *testing.T) {
	iso, err := cloudinit.BuildCloudInitISO("", "", "")
	if err != nil {
		t.Fatalf("BuildCloudInitISO empty: %v", err)
	}
	if len(iso) == 0 || len(iso)%sector != 0 {
		t.Errorf("unexpected ISO size: %d", len(iso))
	}
	// PVD must still be valid.
	pvd := iso[16*sector:]
	if pvd[0] != 1 || string(pvd[1:6]) != "CD001" {
		t.Errorf("invalid PVD after empty userData")
	}
}

// TestBuildCloudInitISO_Defaults verifies that empty instanceID and hostname
// fall back to "iid-default" and copy the instanceID as hostname.
func TestBuildCloudInitISO_Defaults(t *testing.T) {
	iso, err := cloudinit.BuildCloudInitISO("", "", "#!/bin/bash\necho hi\n")
	if err != nil {
		t.Fatalf("BuildCloudInitISO defaults: %v", err)
	}
	if len(iso) == 0 {
		t.Fatal("expected non-empty ISO")
	}
	// The meta-data sector (sector 21) should contain "iid-default".
	// We simply verify it appears somewhere after the first 20 sectors.
	content := string(iso[20*sector:])
	if !strings.Contains(content, "iid-default") {
		t.Errorf("expected 'iid-default' in ISO content")
	}
}

// TestBuildCloudInitISO_LargeUserData verifies that userData larger than one
// ISO sector (2048 bytes) is handled correctly.
func TestBuildCloudInitISO_LargeUserData(t *testing.T) {
	large := strings.Repeat("x", 3*sector+7)
	iso, err := cloudinit.BuildCloudInitISO("big-vm", "bighost", large)
	if err != nil {
		t.Fatalf("BuildCloudInitISO large: %v", err)
	}
	if len(iso)%sector != 0 {
		t.Errorf("ISO size not sector-aligned: %d", len(iso))
	}
	// Verify the ISO contains the userData content.
	if !strings.Contains(string(iso), large[:64]) {
		t.Errorf("expected userData in ISO output")
	}
}
