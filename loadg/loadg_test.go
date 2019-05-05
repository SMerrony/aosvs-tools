package main

import "testing"

func TestGetKnownEntryTypes(t *testing.T) {
	fmtf := KnownFstatEntryTypes[2]
	if "FMTF" != fmtf.DgMnemonic {
		t.Errorf("Expected 'FMTF', got '%s'", fmtf.DgMnemonic)
	}
}
