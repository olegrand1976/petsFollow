package handlers

import "testing"

func TestValidIBAN(t *testing.T) {
	// Well-known valid BE IBAN (mod-97).
	if !validIBAN(normalizeIBAN("BE68 5390 0754 7034")) {
		t.Fatal("expected valid BE IBAN")
	}
	if validIBAN(normalizeIBAN("BE68 5390 0754 7035")) {
		t.Fatal("expected invalid checksum")
	}
	if validIBAN("SHORT") {
		t.Fatal("expected too short")
	}
	if validIBAN(normalizeIBAN("BE68 5390 0754 7034!")) {
		t.Fatal("expected invalid charset")
	}
}

func TestValidBIC(t *testing.T) {
	if !validBIC("") {
		t.Fatal("empty BIC allowed")
	}
	if !validBIC("GEBABEBB") {
		t.Fatal("8-char BIC")
	}
	if !validBIC("GEBABEBBXXX") {
		t.Fatal("11-char BIC")
	}
	if validBIC("SHORT") || validBIC("TOOLONGCODE12") {
		t.Fatal("invalid BIC lengths")
	}
}
