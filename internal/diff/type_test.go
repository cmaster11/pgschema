package diff

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/pgplex/pgschema/ir"
	"github.com/pgplex/pgschema/testutil"
)

func TestDefaultMultirangeTypeName(t *testing.T) {
	cases := map[string]string{
		"timerange":  "timemultirange",
		"my_range_t": "my_multirange_t",
		"floatspan":  "floatspan_multirange",
		"rangerange": "multirangerange", // only the first occurrence
	}
	for in, want := range cases {
		if got := defaultMultirangeTypeName(in); got != want {
			t.Errorf("defaultMultirangeTypeName(%q) = %q, want %q", in, got, want)
		}
	}

	// Without "range" in the name the server clips the range name to
	// NAMEDATALEN-12 bytes and keeps the whole suffix.
	long := strings.Repeat("a", 60) + "span"
	if got, want := defaultMultirangeTypeName(long), strings.Repeat("a", 52)+"_multirange"; got != want {
		t.Errorf("long name: got %q, want %q", got, want)
	}
	// With "range" in the name the combined name is clipped at the end.
	longRange := strings.Repeat("b", 55) + "range"
	if got, want := defaultMultirangeTypeName(longRange), strings.Repeat("b", 55)+"multiran"; got != want {
		t.Errorf("long range name: got %q, want %q", got, want)
	}
}

func TestGenerateTypeSQLRange(t *testing.T) {
	cases := []struct {
		name string
		typ  *ir.Type
		want string
	}{
		{
			name: "default multirange name is omitted",
			typ: &ir.Type{
				Schema: "public", Name: "timerange", Kind: ir.TypeKindRange,
				Subtype: "time without time zone", MultirangeTypeName: "timemultirange",
				SubtypeDiff: "time_diff",
			},
			want: "CREATE TYPE timerange AS RANGE (\n    subtype = time without time zone,\n    subtype_diff = time_diff\n);",
		},
		{
			name: "explicit multirange name, opclass and collation",
			typ: &ir.Type{
				Schema: "public", Name: "textrange", Kind: ir.TypeKindRange,
				Subtype: "text", MultirangeTypeName: "textspans",
				SubtypeOpclass: "pg_catalog.text_pattern_ops", Collation: `"C"`,
			},
			want: "CREATE TYPE textrange AS RANGE (\n    subtype = text,\n    multirange_type_name = textspans,\n    subtype_opclass = pg_catalog.text_pattern_ops,\n    collation = \"C\"\n);",
		},
		{
			name: "multirange in another schema is printed qualified",
			typ: &ir.Type{
				Schema: "public", Name: "floatspan", Kind: ir.TypeKindRange,
				Subtype: "double precision", MultirangeTypeName: "other.floatspan_multirange",
			},
			want: "CREATE TYPE floatspan AS RANGE (\n    subtype = double precision,\n    multirange_type_name = other.floatspan_multirange\n);",
		},
		{
			name: "opclass and collation in the target schema lose the prefix",
			typ: &ir.Type{
				Schema: "public", Name: "textrange", Kind: ir.TypeKindRange,
				Subtype: "text", MultirangeTypeName: "textmultirange",
				SubtypeOpclass: "public.my_text_ops", Collation: "public.my_coll",
			},
			want: "CREATE TYPE textrange AS RANGE (\n    subtype = text,\n    subtype_opclass = my_text_ops,\n    collation = my_coll\n);",
		},
		{
			name: "same-schema subtype prefix is stripped",
			typ: &ir.Type{
				Schema: "public", Name: "band", Kind: ir.TypeKindRange,
				Subtype: "public.severity", MultirangeTypeName: "band_multirange",
			},
			want: "CREATE TYPE band AS RANGE (\n    subtype = severity\n);",
		},
	}
	for _, tc := range cases {
		if got := generateTypeSQL(tc.typ, "public", false); got != tc.want {
			t.Errorf("%s:\n got: %q\nwant: %q", tc.name, got, tc.want)
		}
	}
}

func TestTypesEqualRange(t *testing.T) {
	base := func() *ir.Type {
		return &ir.Type{
			Schema: "public", Name: "r", Kind: ir.TypeKindRange,
			Subtype: "integer", MultirangeTypeName: "rmulti", SubtypeDiff: "d",
		}
	}
	if !typesEqual(base(), base()) {
		t.Fatal("identical ranges compare unequal")
	}
	for name, mutate := range map[string]func(*ir.Type){
		"subtype":    func(r *ir.Type) { r.Subtype = "bigint" },
		"opclass":    func(r *ir.Type) { r.SubtypeOpclass = "pg_catalog.int4_ops" },
		"collation":  func(r *ir.Type) { r.Collation = `"C"` },
		"canonical":  func(r *ir.Type) { r.Canonical = "c" },
		"diff":       func(r *ir.Type) { r.SubtypeDiff = "" },
		"multirange": func(r *ir.Type) { r.MultirangeTypeName = "other" },
	} {
		changed := base()
		mutate(changed)
		if typesEqual(base(), changed) {
			t.Errorf("%s change not detected", name)
		}
	}
}

func TestTopologicallySortTypesRange(t *testing.T) {
	types := []*ir.Type{
		// Dependents sort before the range by name, so only the edges order them.
		{Schema: "public", Name: "availability", Kind: ir.TypeKindDomain, BaseType: "public.period_multirange"},
		{Schema: "public", Name: "booking", Kind: ir.TypeKindComposite, Columns: []*ir.TypeColumn{{Name: "slot", DataType: "public.period"}}},
		{Schema: "public", Name: "period", Kind: ir.TypeKindRange, Subtype: "public.severity", MultirangeTypeName: "period_multirange"},
		{Schema: "public", Name: "severity", Kind: ir.TypeKindEnum, EnumValues: []string{"low", "high"}},
	}

	sorted := topologicallySortTypes(types)
	order := make(map[string]int, len(sorted))
	for idx, typ := range sorted {
		order[typ.Name] = idx
	}
	assertBefore := func(first, second string) {
		if order[first] >= order[second] {
			t.Fatalf("expected %s before %s in %v", first, second, order)
		}
	}
	assertBefore("severity", "period")
	assertBefore("period", "booking")
	assertBefore("period", "availability")
}

func TestHoldBackFunctionsOfDroppedTypes(t *testing.T) {
	timeParam := &ir.Parameter{Mode: "IN", DataType: "time without time zone"}
	diffFn := &ir.Function{Schema: "public", Name: "time_diff", Parameters: []*ir.Parameter{timeParam, timeParam}}
	checkFn := &ir.Function{Schema: "public", Name: "is_valid"}
	other := &ir.Function{Schema: "public", Name: "unrelated"}
	types := []*ir.Type{
		{Schema: "public", Name: "timerange", Kind: ir.TypeKindRange, Subtype: "time without time zone", SubtypeDiff: "time_diff"},
		{Schema: "public", Name: "code", Kind: ir.TypeKindDomain, Constraints: []*ir.DomainConstraint{{Definition: "CHECK (is_valid(VALUE))"}}},
	}

	now, after := holdBackFunctionsOfDroppedTypes([]*ir.Function{diffFn, other, checkFn}, types)
	if len(now) != 1 || now[0] != other {
		t.Fatalf("expected only the unrelated function to drop now, got %v", now)
	}
	if len(after) != 2 || after[0] != diffFn || after[1] != checkFn {
		t.Fatalf("expected the referenced functions to wait in order, got %v", after)
	}

	// The subtype keeps the format_type spelling while parameter types are
	// normalized to the short one; they must still match.
	tsParam := &ir.Parameter{Mode: "IN", DataType: "timestamp"}
	tsDiff := &ir.Function{Schema: "public", Name: "ts_diff", Parameters: []*ir.Parameter{tsParam, tsParam}}
	tsRange := &ir.Type{Schema: "public", Name: "stamp_span", Kind: ir.TypeKindRange, Subtype: "timestamp without time zone", SubtypeDiff: "ts_diff"}
	if now, _ := holdBackFunctionsOfDroppedTypes([]*ir.Function{tsDiff}, []*ir.Type{tsRange}); len(now) != 0 {
		t.Fatalf("timestamp subtype must match the normalized parameter type")
	}

	// Same name, different signature: an overload taking the range is not
	// its subtype_diff and is not held.
	overload := &ir.Function{Schema: "public", Name: "time_diff", Parameters: []*ir.Parameter{{Mode: "IN", DataType: "timerange"}}}
	now, _ = holdBackFunctionsOfDroppedTypes([]*ir.Function{overload}, types[:1])
	if len(now) != 1 {
		t.Fatalf("overload taking the range must not be held as its subtype_diff")
	}

	now, after = holdBackFunctionsOfDroppedTypes([]*ir.Function{other}, nil)
	if len(now) != 1 || after != nil {
		t.Fatalf("no dropped types must leave the functions untouched")
	}
}

func TestDependentTypes(t *testing.T) {
	period := &ir.Type{Schema: "public", Name: "period", Kind: ir.TypeKindRange, Subtype: "date", MultirangeTypeName: "period_multirange"}
	booking := &ir.Type{Schema: "public", Name: "booking", Kind: ir.TypeKindComposite, Columns: []*ir.TypeColumn{{Name: "slot", DataType: "public.period"}}}
	availability := &ir.Type{Schema: "public", Name: "availability", Kind: ir.TypeKindDomain, BaseType: "public.period_multirange"}
	window := &ir.Type{Schema: "public", Name: "window", Kind: ir.TypeKindRange, Subtype: "public.availability", MultirangeTypeName: "window_multirange"}
	unrelated := &ir.Type{Schema: "public", Name: "color", Kind: ir.TypeKindEnum, EnumValues: []string{"red"}}
	all := []*ir.Type{availability, booking, unrelated, period, window}

	got := dependentTypes(all, []*ir.Type{period})
	want := []*ir.Type{availability, booking, window} // in the order of all, seeds excluded
	if len(got) != len(want) {
		t.Fatalf("got %d dependents, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("dependent %d: got %s, want %s", i, got[i].Name, want[i].Name)
		}
	}
	if dependentTypes(all, nil) != nil {
		t.Fatal("no seeds must give no dependents")
	}
}

func TestOrderTypeAndFunctionDrops(t *testing.T) {
	isPositive := &ir.Function{Schema: "public", Name: "is_positive", Parameters: []*ir.Parameter{{Mode: "IN", DataType: "integer"}}}
	positive := &ir.Type{Schema: "public", Name: "positive", Kind: ir.TypeKindDomain, BaseType: "integer", Constraints: []*ir.DomainConstraint{{Definition: "CHECK (is_positive(VALUE))"}}}
	positiveDiff := &ir.Function{Schema: "public", Name: "positive_diff", Parameters: []*ir.Parameter{{Mode: "IN", DataType: "positive"}, {Mode: "IN", DataType: "positive"}}}
	span := &ir.Type{Schema: "public", Name: "positive_span", Kind: ir.TypeKindRange, Subtype: "public.positive", SubtypeDiff: "positive_diff", MultirangeTypeName: "positive_span_multirange"}
	spanSet := &ir.Type{Schema: "public", Name: "busy", Kind: ir.TypeKindDomain, BaseType: "public.positive_span_multirange"}

	var names []string
	for _, node := range orderTypeAndFunctionDrops([]*ir.Type{positive, span, spanSet}, []*ir.Function{isPositive, positiveDiff}) {
		if node.function != nil {
			names = append(names, node.function.Name)
		} else {
			names = append(names, node.typeObj.Name)
		}
	}
	want := []string{"busy", "positive_span", "positive_diff", "positive", "is_positive"}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Fatalf("got %v, want %v", names, want)
	}

	// A held function that the catalog records as depending on a type (a
	// SQL-standard body cast) drops before that type even though no type
	// references it through a signature.
	status := &ir.Type{Schema: "public", Name: "z_status", Kind: ir.TypeKindEnum}
	bodyFn := &ir.Function{Schema: "public", Name: "span_diff", TypeDependencies: []string{"public.z_status"}, Parameters: []*ir.Parameter{{Mode: "IN", DataType: "integer"}, {Mode: "IN", DataType: "integer"}}}
	intSpan := &ir.Type{Schema: "public", Name: "span", Kind: ir.TypeKindRange, Subtype: "integer", SubtypeDiff: "span_diff", MultirangeTypeName: "span_multirange"}
	names = nil
	for _, node := range orderTypeAndFunctionDrops([]*ir.Type{status, intSpan}, []*ir.Function{bodyFn}) {
		if node.function != nil {
			names = append(names, node.function.Name)
		} else {
			names = append(names, node.typeObj.Name)
		}
	}
	if strings.Join(names, ",") != "span,span_diff,z_status" {
		t.Fatalf("got %v", names)
	}

	// Two ranges whose difference functions cross-reference: a_diff's body
	// only mentions z_span in a literal (no recorded dependency), z_diff's
	// body casts to a_span (recorded). z_diff must drop before a_span.
	aDiff := &ir.Function{Schema: "public", Name: "a_diff", Definition: "SELECT 'z_span'", Parameters: []*ir.Parameter{{Mode: "IN", DataType: "integer"}, {Mode: "IN", DataType: "integer"}}}
	zDiff := &ir.Function{Schema: "public", Name: "z_diff", TypeDependencies: []string{"public.a_span"}, Parameters: []*ir.Parameter{{Mode: "IN", DataType: "integer"}, {Mode: "IN", DataType: "integer"}}}
	aSpan := &ir.Type{Schema: "public", Name: "a_span", Kind: ir.TypeKindRange, Subtype: "integer", SubtypeDiff: "a_diff", MultirangeTypeName: "a_span_multirange"}
	zSpan := &ir.Type{Schema: "public", Name: "z_span", Kind: ir.TypeKindRange, Subtype: "integer", SubtypeDiff: "z_diff", MultirangeTypeName: "z_span_multirange"}
	names = nil
	for _, node := range orderTypeAndFunctionDrops([]*ir.Type{aSpan, zSpan}, []*ir.Function{aDiff, zDiff}) {
		if node.function != nil {
			names = append(names, node.function.Name)
		} else {
			names = append(names, node.typeObj.Name)
		}
	}
	if slices.Index(names, "z_diff") > slices.Index(names, "a_span") || slices.Index(names, "a_span") > slices.Index(names, "a_diff") || slices.Index(names, "z_span") > slices.Index(names, "z_diff") {
		t.Fatalf("got %v", names)
	}

	// No held functions: plain reverse dependency order of the types.
	names = nil
	for _, node := range orderTypeAndFunctionDrops([]*ir.Type{positive, span}, nil) {
		names = append(names, node.typeObj.Name)
	}
	if strings.Join(names, ",") != "positive_span,positive" {
		t.Fatalf("got %v", names)
	}
}

func TestSplitTypeAndFunctionDrops(t *testing.T) {
	isPositive := &ir.Function{Schema: "public", Name: "is_positive", Parameters: []*ir.Parameter{{Mode: "IN", DataType: "integer"}}}
	positive := &ir.Type{Schema: "public", Name: "positive", Kind: ir.TypeKindDomain, BaseType: "integer", Constraints: []*ir.DomainConstraint{{Definition: "CHECK (is_positive(VALUE))"}}}
	positiveDiff := &ir.Function{Schema: "public", Name: "positive_diff", Parameters: []*ir.Parameter{{Mode: "IN", DataType: "positive"}, {Mode: "IN", DataType: "positive"}}}
	span := &ir.Type{Schema: "public", Name: "positive_span", Kind: ir.TypeKindRange, Subtype: "public.positive", SubtypeDiff: "positive_diff", MultirangeTypeName: "positive_span_multirange"}
	spanSet := &ir.Type{Schema: "public", Name: "busy", Kind: ir.TypeKindDomain, BaseType: "public.positive_span_multirange"}
	types := []*ir.Type{positive, span, spanSet}
	functions := []*ir.Function{isPositive, positiveDiff}

	nodeNames := func(nodes []droppedTypeOrFunction) string {
		var names []string
		for _, node := range nodes {
			if node.function != nil {
				names = append(names, node.function.Name)
			} else {
				names = append(names, node.typeObj.Name)
			}
		}
		return strings.Join(names, ",")
	}

	// A waiting type takes along everything that must be dropped after it,
	// and both parts keep the joint order. A type that must stay keeps every
	// type it would be taken along by in the first part.
	cases := []struct {
		wait      string
		stay      string
		now       string
		afterward string
	}{
		{"positive_span", "", "busy", "positive_span,positive_diff,positive,is_positive"},
		{"positive", "", "busy,positive_span,positive_diff", "positive,is_positive"},
		{"busy", "", "", "busy,positive_span,positive_diff,positive,is_positive"},
		{"", "", "busy,positive_span,positive_diff,positive,is_positive", ""},
		{"busy", "positive_span", "busy,positive_span,positive_diff,positive,is_positive", ""},
		{"positive_span", "positive_span", "busy,positive_span,positive_diff,positive,is_positive", ""},
		{"positive_span", "busy", "busy", "positive_span,positive_diff,positive,is_positive"},
	}
	for _, c := range cases {
		now, later := splitTypeAndFunctionDrops(types, functions,
			func(t *ir.Type) bool { return t.Name == c.wait },
			func(t *ir.Type) bool { return t.Name == c.stay })
		if nodeNames(now) != c.now || nodeNames(later) != c.afterward {
			t.Errorf("wait %q, stay %q: got now %q and later %q, want %q and %q", c.wait, c.stay, nodeNames(now), nodeNames(later), c.now, c.afterward)
		}
	}
}

func TestTableUsesType(t *testing.T) {
	types := map[string]struct{}{"public.tone_old": {}}
	const cast = "((note)::tone_old = 'soft'::tone_old)"
	cases := map[string]*ir.Table{
		"check constraint": {Constraints: map[string]*ir.Constraint{"c": {Type: ir.ConstraintTypeCheck, CheckClause: "CHECK " + cast}}},
		"exclusion constraint": {Constraints: map[string]*ir.Constraint{"x": {Type: ir.ConstraintTypeExclusion,
			ExclusionDefinition: "EXCLUDE USING btree (id WITH =) WHERE " + cast}}},
		"index expression":  {Indexes: map[string]*ir.Index{"i": {Columns: []*ir.IndexColumn{{Name: "((note)::tone_old)"}}, IsExpression: true}}},
		"index predicate":   {Indexes: map[string]*ir.Index{"i": {Columns: []*ir.IndexColumn{{Name: "id"}}, IsPartial: true, Where: cast}}},
		"policy using":      {Policies: map[string]*ir.RLSPolicy{"p": {Using: cast}}},
		"policy with check": {Policies: map[string]*ir.RLSPolicy{"p": {WithCheck: cast}}},
		"trigger condition": {Triggers: map[string]*ir.Trigger{"tr": {Condition: "((new.note)::tone_old = 'soft'::tone_old)"}}},
	}
	for name, table := range cases {
		table.Schema = "public"
		if !tableUsesType(table, types) {
			t.Errorf("%s: tableUsesType() = false, want true", name)
		}
	}
	unrelated := &ir.Table{
		Schema: "public",
		Constraints: map[string]*ir.Constraint{"x": {Type: ir.ConstraintTypeExclusion,
			ExclusionDefinition: "EXCLUDE USING btree (id WITH =) WHERE ((note = 'soft'::text))"}},
		Indexes:  map[string]*ir.Index{"i": {Columns: []*ir.IndexColumn{{Name: "note"}}, IsPartial: true, Where: "(note = 'soft'::text)"}},
		Policies: map[string]*ir.RLSPolicy{"p": {Using: "(note = 'soft'::text)", WithCheck: "true"}},
		Triggers: map[string]*ir.Trigger{"tr": {Condition: "(new.note IS NOT NULL)"}},
	}
	if tableUsesType(unrelated, types) {
		t.Errorf("tableUsesType() = true for a table without the type, want false")
	}
}

// A replaced type (a range whose definition changes) must be dropped before
// the create phase creates it again, also when a dropped domain built on it
// is still used by a column that the modify phase changes.
func TestReplacedTypeDroppedBeforeItsRecreate(t *testing.T) {
	oldIR := testutil.ParseSQLToIR(t, sharedTestPostgres, `
CREATE TYPE r AS RANGE (subtype = int4);
CREATE DOMAIN w AS r;
CREATE TABLE t (id integer PRIMARY KEY, v w);`, "public")
	newIR := testutil.ParseSQLToIR(t, sharedTestPostgres, `
CREATE TYPE r AS RANGE (subtype = int8);
CREATE TABLE t (id integer PRIMARY KEY, v int4range);`, "public")

	sqlText := buildSQLFromSteps(GenerateMigration(oldIR, newIR, "public"))
	drop := strings.Index(sqlText, "DROP TYPE IF EXISTS r RESTRICT;")
	create := strings.Index(sqlText, "CREATE TYPE r AS RANGE")
	if drop < 0 || create < 0 || drop > create {
		t.Errorf("want DROP TYPE r before CREATE TYPE r, got:\n%s", sqlText)
	}
}

// A dropped type whose name a new object takes must be dropped before that
// object is created, also when a column the modify phase changes still uses
// it. A new relation takes the name with its row type. A new index, or a
// constraint that creates one, takes the pg_class name of a composite type,
// but not the name of an enum. Without such an object, the drop waits for the
// column change.
func TestDroppedTypeWaitsUnlessNewObjectTakesItsName(t *testing.T) {
	const (
		enum      = "CREATE TYPE status AS ENUM ('a', 'b');"
		composite = "CREATE TYPE status AS (a integer, b integer);"
	)
	cases := []struct {
		name    string
		oldType string
		newSQL  string
		create  string
		stays   bool
	}{
		{"table", enum, "CREATE TABLE status (id integer PRIMARY KEY);", "CREATE TABLE IF NOT EXISTS status", true},
		{"view", enum, "CREATE VIEW status AS SELECT 1 AS id;", "CREATE OR REPLACE VIEW status", true},
		{"materialized view", enum, "CREATE MATERIALIZED VIEW status AS SELECT 1 AS id;", "CREATE MATERIALIZED VIEW IF NOT EXISTS status", true},
		{"sequence", enum, "CREATE SEQUENCE status;", "CREATE SEQUENCE IF NOT EXISTS status", true},
		{"no relation", enum, "CREATE TABLE state (id integer PRIMARY KEY);", "CREATE TABLE IF NOT EXISTS state", false},
		{"index and enum", enum, "CREATE INDEX status ON orders (s);", "CREATE INDEX IF NOT EXISTS status ON orders", false},
		{"index and composite", composite, "CREATE INDEX status ON orders (s);", "CREATE INDEX IF NOT EXISTS status ON orders", true},
		{"primary key and composite", composite, "CREATE TABLE things (id integer CONSTRAINT status PRIMARY KEY);", "CREATE TABLE IF NOT EXISTS things", true},
		{"unique and composite", composite, "ALTER TABLE orders ADD CONSTRAINT status UNIQUE (s);", "ADD CONSTRAINT status UNIQUE", true},
		{"exclusion and composite", composite, "ALTER TABLE orders ADD CONSTRAINT status EXCLUDE USING btree (s WITH =);", "ADD CONSTRAINT status EXCLUDE", true},
		{"materialized view index and composite", composite, "CREATE MATERIALIZED VIEW things AS SELECT 1 AS id; CREATE INDEX status ON things (id);", "CREATE INDEX IF NOT EXISTS status ON things", true},
		{"no index and composite", composite, "CREATE INDEX state ON orders (s);", "CREATE INDEX IF NOT EXISTS state ON orders", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			oldIR := testutil.ParseSQLToIR(t, sharedTestPostgres, c.oldType+`
CREATE TABLE orders (id integer PRIMARY KEY, s status);`, "public")
			newIR := testutil.ParseSQLToIR(t, sharedTestPostgres, `
CREATE TABLE orders (id integer PRIMARY KEY, s text);
`+c.newSQL, "public")

			sqlText := buildSQLFromSteps(GenerateMigration(oldIR, newIR, "public"))
			drop := strings.Index(sqlText, "DROP TYPE IF EXISTS status RESTRICT;")
			create := strings.Index(sqlText, c.create)
			alter := strings.Index(sqlText, "ALTER TABLE orders ALTER COLUMN s TYPE text")
			if drop < 0 || create < 0 || alter < 0 {
				t.Fatalf("missing DROP TYPE status, %q or ALTER COLUMN s in:\n%s", c.create, sqlText)
			}
			if c.stays && drop > create {
				t.Errorf("want DROP TYPE status before %q, got:\n%s", c.create, sqlText)
			}
			if !c.stays && drop < alter {
				t.Errorf("want DROP TYPE status after ALTER COLUMN s, got:\n%s", sqlText)
			}
		})
	}
}

// A dropped type that the current definition of a modified function,
// procedure or aggregate uses waits until that routine is changed.
func TestDroppedTypeWaitsForModifiedRoutine(t *testing.T) {
	cases := []struct {
		name    string
		routine string
		changed string
	}{
		{"function", `
CREATE FUNCTION pick() RETURNS text LANGUAGE sql
BEGIN ATOMIC SELECT ('a'::%s)::text; END;`, "CREATE OR REPLACE FUNCTION pick"},
		{"procedure", `
CREATE TABLE audit (id integer PRIMARY KEY, status text);
CREATE PROCEDURE record_status(p_id integer) LANGUAGE sql
BEGIN ATOMIC INSERT INTO audit (id, status) VALUES (p_id, ('a'::%s)::text); END;`, "CREATE OR REPLACE PROCEDURE record_status"},
		{"aggregate", `
CREATE FUNCTION keep_state(s anyenum, v integer) RETURNS anyenum LANGUAGE sql IMMUTABLE AS $$ SELECT s $$;
CREATE AGGREGATE tally(integer) (SFUNC = keep_state, STYPE = %s, INITCOND = 'a');`, "CREATE AGGREGATE tally"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			oldIR := testutil.ParseSQLToIR(t, sharedTestPostgres, "CREATE TYPE status_old AS ENUM ('a', 'b');"+
				fmt.Sprintf(c.routine, "status_old"), "public")
			newIR := testutil.ParseSQLToIR(t, sharedTestPostgres, "CREATE TYPE status_new AS ENUM ('a', 'b');"+
				fmt.Sprintf(c.routine, "status_new"), "public")

			sqlText := buildSQLFromSteps(GenerateMigration(oldIR, newIR, "public"))
			drop := strings.Index(sqlText, "DROP TYPE IF EXISTS status_old RESTRICT;")
			changed := strings.Index(sqlText, c.changed)
			if drop < 0 || changed < 0 || drop < changed {
				t.Errorf("want DROP TYPE status_old after %q, got:\n%s", c.changed, sqlText)
			}
		})
	}
}

func TestExprCastsToType(t *testing.T) {
	types := map[string]struct{}{"public.status_old": {}}
	cases := map[string]bool{
		"'a'::status_old":                     true,
		"('a'::public.status_old)::text":      true,
		"'{a}'::status_old[]":                 true,
		"(s <> 'b'::\"status_old\")":          true,
		"'a'::status_older":                   false,
		"'x::status_old'::text":               false,
		"status_old":                          false,
		"'a'::other.status_old":               false,
		"nextval('status_old_seq'::regclass)": false,
		"":                                    false,
	}
	for expr, want := range cases {
		if got := exprCastsToType(expr, "public", types); got != want {
			t.Errorf("exprCastsToType(%q) = %v, want %v", expr, got, want)
		}
	}
}

func TestGenerateAlterCompositeStatements(t *testing.T) {
	oldType := &ir.Type{
		Schema: "public", Name: "address", Kind: ir.TypeKindComposite,
		Columns: []*ir.TypeColumn{
			{Name: "street", DataType: "text"},
			{Name: "zip", DataType: "integer"},
			{Name: "country", DataType: "text"},
		},
	}
	newType := &ir.Type{
		Schema: "public", Name: "address", Kind: ir.TypeKindComposite,
		Columns: []*ir.TypeColumn{
			{Name: "zip", DataType: "bigint"},
			{Name: "street", DataType: "text"},
			{Name: "city", DataType: "text"},
			{Name: "region", DataType: "public.region_kind"},
		},
	}
	got := generateDropCompositeAttributeStatements(oldType, newType, "public", false)
	got = append(got, generateAlterCompositeStatements(oldType, newType, "public", false, nil)...)
	// A filter picks the attributes of a batch
	got = append(got, generateAlterCompositeStatements(oldType, newType, "public", false, func(_ *ir.Type, attr *ir.TypeColumn) bool {
		return attr.Name == "region"
	})...)
	want := []string{
		"ALTER TYPE address DROP ATTRIBUTE country;",
		"ALTER TYPE address ALTER ATTRIBUTE zip TYPE bigint;",
		"ALTER TYPE address ADD ATTRIBUTE city text;",
		"ALTER TYPE address ADD ATTRIBUTE region region_kind;",
		"ALTER TYPE address ADD ATTRIBUTE region region_kind;",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d statements, want %d:\n%s", len(got), len(want), strings.Join(got, "\n"))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("statement %d:\n got: %q\nwant: %q", i, got[i], want[i])
		}
	}
}

func TestTypesEqualComposite(t *testing.T) {
	composite := func(cols ...*ir.TypeColumn) *ir.Type {
		return &ir.Type{Schema: "public", Name: "c", Kind: ir.TypeKindComposite, Columns: cols}
	}
	a, b := &ir.TypeColumn{Name: "a", DataType: "integer"}, &ir.TypeColumn{Name: "b", DataType: "text"}
	if !typesEqual(composite(a, b), composite(a, b)) {
		t.Error("identical composites compare unequal")
	}
	// ADD ATTRIBUTE can only append, so attribute order alone is not a change.
	if !typesEqual(composite(a, b), composite(b, a)) {
		t.Error("reordered attributes reported as a change")
	}
	if typesEqual(composite(a), composite(a, b)) {
		t.Error("added attribute not detected")
	}
	if typesEqual(composite(a, b), composite(a, &ir.TypeColumn{Name: "b", DataType: "varchar(10)"})) {
		t.Error("changed attribute type not detected")
	}
	if typesEqual(composite(a, b), composite(a, &ir.TypeColumn{Name: "c", DataType: "text"})) {
		t.Error("renamed attribute not detected")
	}
}
