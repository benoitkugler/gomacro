package sql

import (
	"reflect"
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
		{ct: "ADD PRIMARY KEY(id_camp )", want: "id_camp"},
	}
	for _, tt := range tests {
		if got := isUniqueConstraint(tt.ct); got != tt.want {
			t.Errorf("isUniqueConstraint() = %v, want %v", got, tt.want)
		}
	}
}

func Test_isUniquesConstraint(t *testing.T) {
	tests := []struct {
		ct   string
		want []string
	}{
		{ct: "ADD UNIQUE(id_camp, id_personne,  id_groupe)", want: []string{"id_camp", "id_personne", "id_groupe"}},
		{ct: "ADD PRIMARY KEY (id_camp, id_personne,  id_groupe)", want: []string{"id_camp", "id_personne", "id_groupe"}},
		{ct: "ADD UNIQUE(id_camp)", want: []string{"id_camp"}},
		{ct: "ADD UNIQUE ", want: nil},
	}
	for _, tt := range tests {
		if got := isUniquesConstraint(tt.ct); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("isUniquesConstraint() = %v, want %v", got, tt.want)
		}
	}
}

func Test_isSelectKey(t *testing.T) {
	tests := []struct {
		ct   string
		want []string
	}{
		{ct: "_SELECT KEY(IdTeacher, IdTrivial)", want: []string{"IdTeacher", "IdTrivial"}},
		{ct: "_SELECT KEY (IdTeacher , IdTrivial)", want: []string{"IdTeacher", "IdTrivial"}},
		{ct: "_SELECT (IdTeacher, IdTrivial)", want: nil},
	}
	for _, tt := range tests {
		if got := isSelectKey(tt.ct); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("isSelectKey() = %v, want %v", got, tt.want)
		}
	}
}
