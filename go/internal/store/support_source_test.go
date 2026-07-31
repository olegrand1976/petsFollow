package store

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func TestResolveSupportSource(t *testing.T) {
	cases := []struct {
		role kernel.Role
		req  string
		want string
	}{
		{kernel.RoleClient, "nuxt_pro", SupportSourceFlutterClient},
		{kernel.RoleClient, "flutter_client", SupportSourceFlutterClient},
		{kernel.RoleCarePro, "nuxt_pro", SupportSourceFlutterProLight},
		{kernel.RoleCarePro, "", SupportSourceFlutterProLight},
		{kernel.RoleVet, "flutter_pro_light", SupportSourceFlutterProLight},
		{kernel.RoleVet, "nuxt_pro", SupportSourceNuxtPro},
		{kernel.RoleVet, "flutter_client", SupportSourceNuxtPro},
		{kernel.RoleVet, "", SupportSourceNuxtPro},
		{kernel.RoleAdmin, "flutter_client", SupportSourceNuxtPro},
		{kernel.RoleCommercial, "flutter_pro_light", SupportSourceNuxtPro},
		{kernel.RoleVetAssistant, "", SupportSourceNuxtPro},
		{kernel.RoleSecretary, "flutter_client", SupportSourceNuxtPro},
		{kernel.RoleCommercialManager, "flutter_pro_light", SupportSourceNuxtPro},
	}
	for _, tc := range cases {
		got := ResolveSupportSource(tc.role, tc.req)
		if got != tc.want {
			t.Errorf("role=%s req=%q: got %s want %s", tc.role, tc.req, got, tc.want)
		}
	}
}
