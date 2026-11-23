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
		{ct: "CREATE UNIQUE INDEX ON Foo(id_camp)", want: nil},
	}
	for _, tt := range tests {
		if got := isUniquesConstraint(tt.ct); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("isUniquesConstraint() = %v, want %v", got, tt.want)
		}
	}
}

func Test_isUniqueIndexColumns(t *testing.T) {
	tests := []struct {
		ct   string
		want []string
	}{
		{ct: "CREATE UNIQUE INDEX ON Dossier(KeyV1) WHERE length(KeyV1) > 0", want: []string{"KeyV1"}},
		{ct: "CREATE UNIQUE INDEX index_name ON QuestionTag (Tag)", want: []string{"Tag"}},
		{ct: "CREATE UNIQUE INDEX ON Foo (A, B) WHERE A IS NOT NULL", want: []string{"A", "B"}},
		{ct: "CREATE INDEX ON Foo (A)", want: nil},
		{ct: "CREATE UNIQUE INDEX ON Foo", want: nil},
	}
	for _, tt := range tests {
		if got := isUniqueIndexColumns(tt.ct); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("isUniqueIndexColumns() = %v, want %v", got, tt.want)
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
		{ct: "_SELECT KEY(Statut)", want: []string{"Statut"}},
	}
	for _, tt := range tests {
		if got := isSelectKey(tt.ct); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("isSelectKey() = %v, want %v", got, tt.want)
		}
	}
}
