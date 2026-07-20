package i18n

import "testing"

func TestAllPushCatalogKeys(t *testing.T) {
	keys := []string{
		"push.new_message_title",
		"push.new_message_body",
		"push.visit_confirmed_title",
		"push.visit_confirmed_body",
		"push.visit_proposed_title",
		"push.visit_proposed_body",
		"push.visit_reschedule_title",
		"push.visit_reschedule_body",
	}
	vars := map[string]string{"preview": "Hello", "petName": "Rex"}
	for _, loc := range Supported {
		for _, key := range keys {
			got := T(loc, key, vars)
			if got == "" || got == key {
				t.Errorf("%s missing/unresolved key %s", loc, key)
			}
		}
	}
}

func TestPushInterpolation(t *testing.T) {
	body := T("en", "push.visit_confirmed_body", map[string]string{"petName": "Bella"})
	if body != "The appointment for Bella is confirmed." {
		t.Fatalf("got %q", body)
	}
	msg := T("fr", "push.new_message_body", map[string]string{"preview": "Coucou"})
	if msg != "Coucou" {
		t.Fatalf("got %q", msg)
	}
}
