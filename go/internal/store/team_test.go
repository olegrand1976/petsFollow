package store

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func TestDefaultTeamPermissions(t *testing.T) {
	ref := DefaultTeamPermissions(TeamRoleReferenceVet)
	if !ref["team.manage"] || !ref["heartrate.validate"] {
		t.Fatalf("reference vet defaults: %#v", ref)
	}
	sec := DefaultTeamPermissions(TeamRoleSecretary)
	if sec["pets.write_clinical"] || sec["heartrate.validate"] || !sec["calendar.manage"] {
		t.Fatalf("secretary defaults: %#v", sec)
	}
	asst := DefaultTeamPermissions(TeamRoleAssistant)
	if asst["shares.manage"] || !asst["care.manage"] {
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
