package gbcarkhos

import "testing"

func TestModuleMetadata(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "module path", got: ModulePath, want: "goark.dev/gbc-arkhos"},
		{name: "repository", got: Repository, want: "goark-boot-contrib-arkhos"},
		{name: "starter id", got: StarterID, want: "goark.boot.contrib.arkhos"},
		{name: "server bean", got: BeanNameServer, want: "goark.boot.arkhos.server"},
		{name: "container bean", got: BeanNameContainer, want: "goark.boot.arkhos.container"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("metadata mismatch: got %q, want %q", tt.got, tt.want)
			}
		})
	}
}
