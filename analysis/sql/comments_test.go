package sql

import (
	"testing"
)

func Test_isUniqueConstraint(t *testing.T) {
	tests := []struct {
		ct   string
		want string
	}{
		{ct: "ADD UNIQUE(id_camp, id_personne,  id_groupe)", want: ""},
		{ct: "ADD UNIQUE(id_camp)", want: "id_camp"},
		{ct: "ADD UNIQUE(id_camp )", want: "id_camp"},
	}
	for _, tt := range tests {
		if got := isUniqueConstraint(tt.ct); got != tt.want {
			t.Errorf("isUniqueConstraint() = %v, want %v", got, tt.want)
		}
	}
}
