package diff

import (
	"testing"

	"github.com/pgplex/pgschema/ir"
)

func TestNormalizeBaseTypeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"integer", "integer"},
		{"INTEGER", "integer"},
		{"bigint", "bigint"},
		{"numeric(18,6)", "numeric"},
		{"numeric(20,6)", "numeric"},
		{"varchar(128)", "varchar"},
		{"character varying(255)", "character varying"},
		{"integer[]", "integer[]"},
		{"character varying(128)[]", "character varying[]"},
		{"timestamp(6) with time zone", "timestamp with time zone"},
		{"pg_catalog.int4", "int4"},
		{`"MyStatus"`, `"MyStatus"`},
		{`"mystatus"`, `"mystatus"`},
		{`"MyStatus"[]`, `"MyStatus"[]`},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeBaseTypeName(tt.input)
			if got != tt.want {
				t.Errorf("normalizeBaseTypeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNeedsUsingClause(t *testing.T) {
	tests := []struct {
		old  string
		new  string
		want bool
	}{
		{"text", "integer", true},
		{"integer", "bigint", true},
		{"numeric(18,6)", "numeric(20,6)", false},
		{"integer", "integer[]", true},
		{"text", "action_type", true},
		{"varchar(128)", "varchar(255)", false},
		{"timestamp(3) with time zone", "timestamp(6) with time zone", false},
		{"timestamp without time zone", "timestamp with time zone", true},
		{`"MyStatus"`, `"mystatus"`, true},
	}
	for _, tt := range tests {
		t.Run(tt.old+"→"+tt.new, func(t *testing.T) {
			got := needsUsingClause(tt.old, tt.new)
			if got != tt.want {
				t.Errorf("needsUsingClause(%q, %q) = %v, want %v", tt.old, tt.new, got, tt.want)
			}
		})
	}
}

// Only a change from an enum that the migration drops, or from a domain over
// it, to an enum converts through text. A change from a kept enum, or to any
// other type, keeps the plain cast, which can use a user-defined cast.
func TestColumnLeavingEnumConvertsThroughText(t *testing.T) {
	oldTypes := map[string]*ir.Type{
		"public.status_old": {Schema: "public", Name: "status_old", Kind: ir.TypeKindEnum},
		"public.code":       {Schema: "public", Name: "code", Kind: ir.TypeKindDomain, BaseType: "text"},
		"public.status_dom": {Schema: "public", Name: "status_dom", Kind: ir.TypeKindDomain, BaseType: "public.status_old"},
		"public.status_tag": {Schema: "public", Name: "status_tag", Kind: ir.TypeKindDomain, BaseType: "public.status_dom"},
		"public.pri":        {Schema: "public", Name: "pri", Kind: ir.TypeKindEnum},
		"public.pri_dom":    {Schema: "public", Name: "pri_dom", Kind: ir.TypeKindDomain, BaseType: "public.pri"},
	}
	newTypes := map[string]*ir.Type{
		"public.status_new":     {Schema: "public", Name: "status_new", Kind: ir.TypeKindEnum},
		"public.code":           {Schema: "public", Name: "code", Kind: ir.TypeKindDomain, BaseType: "text"},
		"public.status_new_dom": {Schema: "public", Name: "status_new_dom", Kind: ir.TypeKindDomain, BaseType: "public.status_new"},
		"public.pri":            {Schema: "public", Name: "pri", Kind: ir.TypeKindEnum},
		"public.pri_dom":        {Schema: "public", Name: "pri_dom", Kind: ir.TypeKindDomain, BaseType: "public.pri"},
		"public.lvl":            {Schema: "public", Name: "lvl", Kind: ir.TypeKindEnum},
	}
	goneTypes := map[string]struct{}{"public.status_old": {}, "public.status_dom": {}, "public.status_tag": {}}
	tests := []struct {
		old  string
		new  string
		want string
	}{
		{"public.status_old", "public.status_new", "ALTER TABLE t ALTER COLUMN s TYPE status_new USING s::text::status_new;"},
		{"public.status_old[]", "public.status_new[]", "ALTER TABLE t ALTER COLUMN s TYPE status_new[] USING s::text::status_new[];"},
		{"public.status_old", "public.status_new_dom", "ALTER TABLE t ALTER COLUMN s TYPE status_new_dom USING s::text::status_new_dom;"},
		{"public.status_old", "text", "ALTER TABLE t ALTER COLUMN s TYPE text USING s::text;"},
		{"public.status_old", "integer", "ALTER TABLE t ALTER COLUMN s TYPE integer USING s::integer;"},
		{"public.status_old", "public.code", "ALTER TABLE t ALTER COLUMN s TYPE code USING s::code;"},
		{"public.status_dom", "integer", "ALTER TABLE t ALTER COLUMN s TYPE integer USING s::integer;"},
		{"text", "public.status_new", "ALTER TABLE t ALTER COLUMN s TYPE status_new USING s::status_new;"},
		{"public.code", "public.status_new", "ALTER TABLE t ALTER COLUMN s TYPE status_new USING s::status_new;"},
		{"public.status_dom", "public.status_new", "ALTER TABLE t ALTER COLUMN s TYPE status_new USING s::text::status_new;"},
		{"public.status_tag", "public.status_new", "ALTER TABLE t ALTER COLUMN s TYPE status_new USING s::text::status_new;"},
		{"public.pri", "public.lvl", "ALTER TABLE t ALTER COLUMN s TYPE lvl USING s::lvl;"},
		{"public.pri[]", "public.lvl[]", "ALTER TABLE t ALTER COLUMN s TYPE lvl[] USING s::lvl[];"},
		{"public.pri_dom", "public.lvl", "ALTER TABLE t ALTER COLUMN s TYPE lvl USING s::lvl;"},
	}
	for _, tt := range tests {
		t.Run(tt.old+"→"+tt.new, func(t *testing.T) {
			td := &tableDiff{
				Table: &ir.Table{Schema: "public", Name: "t"},
				ModifiedColumns: []*ColumnDiff{{
					Old: &ir.Column{Name: "s", DataType: tt.old, IsNullable: true},
					New: &ir.Column{Name: "s", DataType: tt.new, IsNullable: true},
				}},
			}
			markColumnsLeavingEnums([]*tableDiff{td}, oldTypes, newTypes, goneTypes)
			got := td.ModifiedColumns[0].generateColumnSQL("public", "t", "public")
			if len(got) != 1 || got[0] != tt.want {
				t.Errorf("generateColumnSQL() = %q, want [%q]", got, tt.want)
			}
		})
	}
}
