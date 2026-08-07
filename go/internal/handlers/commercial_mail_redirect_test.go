package handlers

import "testing"

// La cible d'un lien tracké est renvoyée en redirection depuis un mail
// commercial : un host ambigu y est un open redirect. Go 1.26 durcit
// net/url.Parse, mais GODEBUG=urlstrictcolons=0 peut le relâcher — le garde-fou
// doit tenir dans les deux cas.
func TestIsSafeRedirectURL(t *testing.T) {
	safe := []string{
		"https://petsfollow.app/produits",
		"http://localhost:3002/login",
		"https://[::1]:8443/labo",
	}
	for _, raw := range safe {
		if !isSafeRedirectURL(raw) {
			t.Errorf("isSafeRedirectURL(%q) = false, want true", raw)
		}
	}

	unsafe := []string{
		"",
		"https://host:80:80/",
		"https://::1/",
		"javascript:alert(1)",
		"data:text/html;base64,PHNjcmlwdD4=",
		"https:///no-host",
		"/relative/path",
	}
	for _, raw := range unsafe {
		if isSafeRedirectURL(raw) {
			t.Errorf("isSafeRedirectURL(%q) = true, want false", raw)
		}
	}
}
