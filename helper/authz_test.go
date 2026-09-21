package helper

import (
	"reflect"
	"testing"
)

func TestPermissionSet(t *testing.T) {
	permissions := NewPermissionSet(map[string][]string{
		"admin": {"user:delete", "user:list", "user:list"},
		"user":  {"profile:read"},
	})

	if !permissions.Can("admin", "user:delete") {
		t.Fatal("admin seharusnya memiliki user:delete")
	}
	if permissions.Can("user", "user:delete") {
		t.Fatal("user tidak seharusnya memiliki user:delete")
	}
	if permissions.Can("unknown", "user:list") {
		t.Fatal("role yang tidak dikenal harus ditolak")
	}
	if got, want := permissions.PermissionsOf("admin"), []string{"user:delete", "user:list"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("PermissionsOf(admin) = %v, want %v", got, want)
	}
	if got, want := permissions.KnownRoles(), []string{"admin", "user"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("KnownRoles() = %v, want %v", got, want)
	}
	if !permissions.IsKnownRole("admin") || permissions.IsKnownRole("staff") {
		t.Fatal("IsKnownRole mengembalikan hasil yang salah")
	}
}

func TestPermissionSetNilFailsClosed(t *testing.T) {
	var permissions *PermissionSet

	if permissions.Can("admin", "user:list") {
		t.Fatal("PermissionSet nil harus menolak permission")
	}
	if permissions.IsKnownRole("admin") {
		t.Fatal("PermissionSet nil tidak boleh mengenal role")
	}
	if got := permissions.PermissionsOf("admin"); got == nil || len(got) != 0 {
		t.Fatalf("PermissionsOf pada PermissionSet nil = %v, want empty slice", got)
	}
	if got := permissions.KnownRoles(); got == nil || len(got) != 0 {
		t.Fatalf("KnownRoles pada PermissionSet nil = %v, want empty slice", got)
	}
}
