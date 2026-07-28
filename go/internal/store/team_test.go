package store

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func TestDefaultTeamPermissions(t *testing.T) {
	ref := DefaultTeamPermissions(TeamRoleReferenceVet)
	if !ref["team.manage"] || !ref["heartrate.validate"] || !ref["shares.read"] || !ref["pharmacy.write"] {
		t.Fatalf("reference vet defaults: %#v", ref)
	}
	sec := DefaultTeamPermissions(TeamRoleSecretary)
	if sec["pets.write_clinical"] || sec["pharmacy.write"] || sec["shares.manage"] || !sec["shares.read"] || !sec["pharmacy.read"] || !sec["calendar.manage"] {
		t.Fatalf("secretary defaults: %#v", sec)
	}
	asst := DefaultTeamPermissions(TeamRoleAssistant)
	if asst["shares.manage"] || !asst["care.manage"] || !asst["pharmacy.write"] || !asst["shares.read"] {
		t.Fatalf("assistant defaults: %#v", asst)
	}
}

func TestMergeTeamPermissionsOverride(t *testing.T) {
	m := mergeTeamPermissions(TeamRoleSecretary, map[string]bool{
		"care.manage":        true,
		"heartrate.validate": true,
	})
	if !m["care.manage"] {
		t.Fatal("override care.manage")
	}
	if m["heartrate.validate"] {
		t.Fatal("secretary cannot elevate heartrate.validate even if override true")
	}
}

func TestMergeTeamPermissionsImpliesRead(t *testing.T) {
	m := mergeTeamPermissions(TeamRoleVet, map[string]bool{
		"shares.manage":  true,
		"shares.read":    false,
		"pharmacy.write": true,
		"pharmacy.read":  false,
	})
	if !m["shares.read"] {
		t.Fatal("shares.manage must imply shares.read")
	}
	if !m["pharmacy.read"] {
		t.Fatal("pharmacy.write must imply pharmacy.read")
	}
}

func TestMergeTeamPermissionsLegacyClinicalMirrorsPharmacy(t *testing.T) {
	// Override pré-pharmacy.* : seule la clé clinique est présente.
	off := mergeTeamPermissions(TeamRoleAssistant, map[string]bool{
		"pets.write_clinical": false,
	})
	if off["pharmacy.write"] {
		t.Fatalf("legacy clinical=false should mirror pharmacy.write off: %#v", off)
	}
	on := mergeTeamPermissions(TeamRoleAssistant, map[string]bool{
		"pets.write_clinical": true,
	})
	if !on["pharmacy.write"] {
		t.Fatalf("legacy clinical=true should mirror pharmacy.write on: %#v", on)
	}
	// Explicite pharmacy.write gagne sur le miroir legacy.
	indep := mergeTeamPermissions(TeamRoleAssistant, map[string]bool{
		"pets.write_clinical": false,
		"pharmacy.write":      true,
	})
	if !indep["pharmacy.write"] || indep["pets.write_clinical"] {
		t.Fatalf("explicit pharmacy.write must stay independent: %#v", indep)
	}
	sec := mergeTeamPermissions(TeamRoleSecretary, map[string]bool{
		"pets.write_clinical": true,
	})
	if sec["pharmacy.write"] || sec["pets.write_clinical"] {
		t.Fatalf("secretary hard-deny must block pharmacy.write: %#v", sec)
	}
}

func TestHardDeniedCapabilitiesClamp(t *testing.T) {
	asst := mergeTeamPermissions(TeamRoleAssistant, map[string]bool{
		"heartrate.validate": true,
		"team.manage":        true,
		"care.manage":        false,
	})
	if asst["heartrate.validate"] || asst["team.manage"] {
		t.Fatalf("assistant hard-denied elevated: %#v", asst)
	}
	if asst["care.manage"] {
		t.Fatal("override false should clamp care.manage off")
	}
	sec := mergeTeamPermissions(TeamRoleSecretary, map[string]bool{
		"pets.write_clinical": true,
		"shares.manage":       true,
		"commissions.view":    true,
	})
	if sec["pets.write_clinical"] || sec["shares.manage"] || sec["commissions.view"] {
		t.Fatalf("secretary hard-denied elevated: %#v", sec)
	}
	vet := mergeTeamPermissions(TeamRoleVet, map[string]bool{
		"practice.settings": true,
		"team.manage":       true,
	})
	if vet["practice.settings"] || vet["team.manage"] {
		t.Fatalf("vet hard-denied elevated: %#v", vet)
	}
}

func TestTeamRoleToKernel(t *testing.T) {
	cases := []struct {
		in   TeamRole
		want kernel.Role
	}{
		{TeamRoleReferenceVet, kernel.RoleVet},
		{TeamRoleVet, kernel.RoleVet},
		{TeamRoleAssistant, kernel.RoleVetAssistant},
		{TeamRoleSecretary, kernel.RoleSecretary},
		{TeamRole("unknown"), kernel.RoleVet},
	}
	for _, c := range cases {
		if got := teamRoleToKernel(c.in); got != c.want {
			t.Fatalf("teamRoleToKernel(%q)=%q want %q", c.in, got, c.want)
		}
	}
}
