package store

import "testing"

func TestLooksLikeEmail(t *testing.T) {
	cases := map[string]bool{
		"a@b.co":   true,
		"bad":      false,
		"@x.com":   false,
		"a@b":      false,
		"a@b.c":    true,
		"":         false,
	}
	for in, want := range cases {
		if got := looksLikeEmail(in); got != want {
			t.Errorf("%q: got %v want %v", in, got, want)
		}
	}
}

func TestMappedValue(t *testing.T) {
	raw := map[string]string{"Courriel": " a@b.co ", "Nom": "Alice"}
	h := "Courriel"
	if got := mappedValue(raw, &h); got != "a@b.co" {
		t.Fatalf("got %q", got)
	}
	if mappedValue(raw, nil) != "" {
		t.Fatal("nil header")
	}
}

func TestEncryptDecryptCredentials(t *testing.T) {
	plain := []byte("email;password\na@b.co;secret\n")
	cipher, err := encryptBytes("dev-key", plain)
	if err != nil {
		t.Fatal(err)
	}
	out, err := decryptBytes("dev-key", cipher)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != string(plain) {
		t.Fatalf("got %q", out)
	}
}
