package main

import "testing"

func TestGetKnownEntryTypes(t *testing.T) {
	kfet := KnownFstatEntryTypes()
	fmtf := kfet[2]
	if "FMTF" != fmtf.DgMnemonic {
		t.Errorf("Expected 'FMTF', got '%s'", fmtf.DgMnemonic)
	}
}
