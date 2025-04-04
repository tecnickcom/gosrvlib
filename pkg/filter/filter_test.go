package filter

import (
	"errors"
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type stringAlias string

func strPtr(v string) *string {
	return &v
}

func getSliceLen(slice any) uint {
	rSlice := reflect.ValueOf(slice)
	rSlice = reflect.Indirect(rSlice)

	return uint(rSlice.Len())
}

func TestParseJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		json    string
		want    [][]Rule
		wantErr bool
	}{
		{
			name: "success",
			json: `[
			  [
				{ "field": "name", "type": "==", "value": "doe" },
				{ "field": "age", "type": "<=", "value": 42 }
			  ],
			  [
				{ "field": "address.country", "type": "regexp", "value": "^EN$|^FR$" }
			  ]
			]`,
			want: [][]Rule{
				{
					{Field: "name", Type: TypeEqual, Value: "doe"},
					{Field: "age", Type: TypeLTE, Value: 42.0},
				},
				{
					{Field: "address.country", Type: TypeRegexp, Value: "^EN$|^FR$"},
				},
			},
			wantErr: false,
		},
		{
			name:    "error - invalid json",
			json:    `[`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := ParseJSON(tt.json)

			if tt.wantErr {
				require.Error(t, err, "ParseRules() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, r, "Filtered = %v, want %v", r, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name: "success",
			opts: []Option{
				func(_ *Processor) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "error - option error",
			opts: []Option{
				func(_ *Processor) error {
					return errors.New("test error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p, err := New(tt.opts...)

			if tt.wantErr {
				require.Error(t, err, "New() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, p)
			}
		})
	}
}

func TestFilter_ParseURLQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rawQuery string
		opts     []Option
		want     [][]Rule
		wantErr  bool
	}{
		{
			// [[{"field":"Age","type":"==","value":42}]]
			name:     "success - default key",
			rawQuery: "filter=%5B%5B%7B%22field%22%3A%22Age%22%2C%22type%22%3A%22%3D%3D%22%2C%22value%22%3A42%7D%5D%5D",
			want: [][]Rule{{{
				Field: "Age",
				Type:  TypeEqual,
				Value: 42.0,
			}}},
			wantErr: false,
		},
		{
			// [[{"field":"Age","type":"==","value":42}]]
			name:     "success - custom key",
			rawQuery: "myCustomFilter=%5B%5B%7B%22field%22%3A%22Age%22%2C%22type%22%3A%22%3D%3D%22%2C%22value%22%3A42%7D%5D%5D",
			opts:     []Option{WithQueryFilterKey("myCustomFilter")},
			want: [][]Rule{{{
				Field: "Age",
				Type:  TypeEqual,
				Value: 42.0,
			}}},
			wantErr: false,
		},
		{
			name:     "success - empty value",
			rawQuery: "filter=",
			want:     nil,
			wantErr:  false,
		},
		{
			name:     "success - missing value",
			rawQuery: "",
			want:     nil,
			wantErr:  false,
		},
		{
			name:     "error - invalid json",
			rawQuery: "filter=%5B",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p, err := New(tt.opts...)
			require.NoError(t, err)

			u := &url.URL{
				RawQuery: tt.rawQuery,
			}
			rules, err := p.ParseURLQuery(u.Query())

			if tt.wantErr {
				require.Error(t, err, "ParseURLQuery() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, rules, "ParseURLQuery rules = %v, want %v", rules, tt.want)
			}
		})
	}
}

//nolint:maintidx
func TestFilter_Apply(t *testing.T) {
	t.Parallel()

	type simpleStruct struct {
		StringField    string `json:"string_field"`
		IntField       int    `json:"int_field,omitempty"`
		Float64Field   float64
		StringPtrField *string
		unexported     string
	}

	type complexStruct struct {
		Internal simpleStruct `json:"internal"`
	}

	type complexStructWithPtr struct {
		Internal *simpleStruct
	}

	type embeddingStruct struct {
		simpleStruct
	}

	trueRegex := Rule{
		Field: "",
		Type:  TypeRegexp,
		Value: ".*",
	}
	falseRegex := Rule{
		Field: "",
		Type:  TypeRegexp,
		Value: "$a",
	}

	tests := []struct {
		name             string
		rules            [][]Rule
		opts             []Option
		elements         any
		want             any
		wantTotalMatches uint
		wantErr          bool
	}{
		{
			name: "success - nested string equal",
			elements: &[]complexStruct{
				{
					Internal: simpleStruct{
						StringField: "value 1",
					},
				},
				{
					Internal: simpleStruct{
						StringField: "value 2",
					},
				},
			},
			rules: [][]Rule{{{
				Field: "Internal.StringField",
				Type:  TypeEqual,
				Value: "value 1",
			}}},
			want: &[]complexStruct{
				{
					Internal: simpleStruct{
						StringField: "value 1",
					},
				},
			},
			wantTotalMatches: 1,
		},
		{
			name: "success - nested string not equal",
			elements: &[]complexStruct{
				{
					Internal: simpleStruct{
						StringField: "value 1",
					},
				},
				{
					Internal: simpleStruct{
						StringField: "value 2",
					},
				},
			},
			rules: [][]Rule{{{
				Field: "Internal.StringField",
				Type:  TypePrefixNot + TypeEqual,
				Value: "value 1",
			}}},
			want: &[]complexStruct{
				{
					Internal: simpleStruct{
						StringField: "value 2",
					},
				},
			},
			wantTotalMatches: 1,
		},
		{
			name: "success - nested regex",
			elements: &[]complexStruct{
				{
					Internal: simpleStruct{
						StringField: "value 1",
					},
				},
				{
					Internal: simpleStruct{
						StringField: "value 2",
					},
				},
			},
			rules: [][]Rule{{{
				Field: "Internal.StringField",
				Type:  TypeRegexp,
				Value: ".* 1",
			}}},
			want: &[]complexStruct{
				{
					Internal: simpleStruct{
						StringField: "value 1",
					},
				},
			},
			wantTotalMatches: 1,
		},
		{
			name: "success - int equal",
			elements: &[]simpleStruct{
				{
					IntField: 42,
				},
				{
					IntField: 43,
				},
			},
			rules: [][]Rule{{{
				Field: "IntField",
				Type:  TypeEqual,
				Value: 42,
			}}},
			want: &[]simpleStruct{
				{
					IntField: 42,
				},
			},
			wantTotalMatches: 1,
		},
		{
			name: "success - float64 equal",
			elements: &[]simpleStruct{
				{
					Float64Field: 42,
				},
				{
					Float64Field: 43,
				},
			},
			rules: [][]Rule{{{
				Field: "Float64Field",
				Type:  TypeEqual,
				Value: 42,
			}}},
			want: &[]simpleStruct{
				{
					Float64Field: 42,
				},
			},
			wantTotalMatches: 1,
		},
		{
			name: "success - nil equal",
			elements: &[]simpleStruct{
				{
					StringPtrField: strPtr("value 1"),
				},
				{
					StringPtrField: nil,
				},
			},
			rules: [][]Rule{{{
				Field: "StringPtrField",
				Type:  TypeEqual,
				Value: nil,
			}}},
			want: &[]simpleStruct{
				{
					StringPtrField: nil,
				},
			},
			wantTotalMatches: 1,
		},
		{
			name: "success - int lt",
			elements: &[]simpleStruct{
				{
					IntField: 41,
				},
				{
					IntField: 42,
				},
				{
					IntField: 43,
				},
			},
			rules: [][]Rule{{{
				Field: "IntField",
				Type:  TypeLT,
				Value: 42,
			}}},
			want: &[]simpleStruct{
				{
					IntField: 41,
				},
			},
			wantTotalMatches: 1,
		},
		{
			name: "success - int lte",
			elements: &[]simpleStruct{
				{
					IntField: 41,
				},
				{
					IntField: 42,
				},
				{
					IntField: 43,
				},
			},
			rules: [][]Rule{{{
				Field: "IntField",
				Type:  TypeLTE,
				Value: 42,
			}}},
			want: &[]simpleStruct{
				{
					IntField: 41,
				},
				{
					IntField: 42,
				},
			},
			wantTotalMatches: 2,
		},
		{
			name: "success - int gt",
			elements: &[]simpleStruct{
				{
					IntField: 41,
				},
				{
					IntField: 42,
				},
				{
					IntField: 43,
				},
			},
			rules: [][]Rule{{{
				Field: "IntField",
				Type:  TypeGT,
				Value: 42,
			}}},
			want: &[]simpleStruct{
				{
					IntField: 43,
				},
			},
			wantTotalMatches: 1,
		},
		{
			name: "success - int gte",
			elements: &[]simpleStruct{
				{
					IntField: 41,
				},
				{
					IntField: 42,
				},
				{
					IntField: 43,
				},
			},
			rules: [][]Rule{{{
				Field: "IntField",
				Type:  TypeGTE,
				Value: 42,
			}}},
			want: &[]simpleStruct{
				{
					IntField: 42,
				},
				{
					IntField: 43,
				},
			},
			wantTotalMatches: 2,
		},
		{
			name: "error - non numeric type for gt",
			elements: &[]simpleStruct{
				{
					IntField: 41,
				},
				{
					IntField: 42,
				},
				{
					IntField: 43,
				},
			},
			rules: [][]Rule{{{
				Field: "IntField",
				Type:  TypeGT,
				Value: "error",
			}}},
			wantErr: true,
		},
		{
			name: "error - non numeric type for gte",
			elements: &[]simpleStruct{
				{
					IntField: 41,
				},
				{
					IntField: 42,
				},
				{
					IntField: 43,
				},
			},
			rules: [][]Rule{{{
				Field: "IntField",
				Type:  TypeGTE,
				Value: "error",
			}}},
			wantErr: true,
		},
		{
			name: "error - non string type for not-contains",
			elements: &[]simpleStruct{
				{
					StringField: "Alpha",
				},
			},
			rules: [][]Rule{{{
				Field: "StringField",
				Type:  TypePrefixNot + TypeContains,
				Value: 5,
			}}},
			wantErr: true,
		},
		{
			name: "success - equal fold",
			elements: &[]simpleStruct{
				{
					StringField: "Alpha",
				},
				{
					StringField: "Beta",
				},
			},
			rules: [][]Rule{{{
				Field: "StringField",
				Type:  TypeEqualFold,
				Value: "beta",
			}}},
			want:             &[]simpleStruct{{StringField: "Beta"}},
			wantTotalMatches: 1,
		},
		{
			name: "success - has prefix",
			elements: &[]simpleStruct{
				{
					StringField: "Alpha",
				},
				{
					StringField: "Beta",
				},
			},
			rules: [][]Rule{{{
				Field: "StringField",
				Type:  TypeHasPrefix,
				Value: "Be",
			}}},
			want:             &[]simpleStruct{{StringField: "Beta"}},
			wantTotalMatches: 1,
		},
		{
			name: "success - has suffix",
			elements: &[]simpleStruct{
				{
					StringField: "Alpha",
				},
				{
					StringField: "Beta",
				},
			},
			rules: [][]Rule{{{
				Field: "StringField",
				Type:  TypeHasSuffix,
				Value: "ta",
			}}},
			want:             &[]simpleStruct{{StringField: "Beta"}},
			wantTotalMatches: 1,
		},
		{
			name: "success - invalid filter value type",
			elements: &[]simpleStruct{
				{
					StringField: "value 1",
				},
			},
			rules: [][]Rule{{{
				Field: "StringField",
				Type:  TypeEqual,
				Value: 42,
			}}},
			want:             &[]simpleStruct{},
			wantTotalMatches: 0,
		},
		{
			name: "success - regexp with an int",
			elements: &[]simpleStruct{
				{
					IntField: 42,
				},
			},
			rules: [][]Rule{{{
				Field: "IntField",
				Type:  TypeRegexp,
				Value: "42",
			}}},
			want:             &[]simpleStruct{},
			wantTotalMatches: 0,
		},
		{
			name: "success - mismatched array",
			elements: &[]any{
				complexStructWithPtr{
					Internal: &simpleStruct{
						StringField: "value 1",
					},
				},
				complexStruct{
					Internal: simpleStruct{
						StringField: "value 2",
					},
				},
			},
			rules: [][]Rule{{{
				Field: "Internal.StringField",
				Type:  TypeEqual,
				Value: "value 1",
			}}},
			want: &[]any{
				complexStructWithPtr{
					Internal: &simpleStruct{
						StringField: "value 1",
					},
				},
			},
			wantTotalMatches: 1,
		},
		{
			name: "success - with field tags",
			elements: &[]complexStruct{
				{
					Internal: simpleStruct{
						StringField: "value 1",
					},
				},
				{
					Internal: simpleStruct{
						StringField: "value 2",
					},
				},
			},
			opts: []Option{WithFieldNameTag("json")},
			rules: [][]Rule{{{
				Field: "internal.string_field",
				Type:  TypeEqual,
				Value: "value 1",
			}}},
			want: &[]complexStruct{
				{
					Internal: simpleStruct{
						StringField: "value 1",
					},
				},
			},
			wantTotalMatches: 1,
		},
		{
			name: "success - with field tags with commas",
			elements: &[]complexStruct{
				{
					Internal: simpleStruct{
						IntField: 1,
					},
				},
				{
					Internal: simpleStruct{
						IntField: 2,
					},
				},
			},
			opts: []Option{WithFieldNameTag("json")},
			rules: [][]Rule{{{
				Field: "internal.int_field",
				Type:  TypeEqual,
				Value: 1,
			}}},
			want: &[]complexStruct{
				{
					Internal: simpleStruct{
						IntField: 1,
					},
				},
			},
			wantTotalMatches: 1,
		},
		{
			name: "success - with embedding struct",
			elements: &[]embeddingStruct{
				{
					simpleStruct: simpleStruct{
						StringField: "value 1",
					},
				},
				{
					simpleStruct: simpleStruct{
						StringField: "value 2",
					},
				},
			},
			rules: [][]Rule{{{
				Field: "StringField",
				Type:  TypeEqual,
				Value: "value 1",
			}}},
			want: &[]embeddingStruct{
				{
					simpleStruct: simpleStruct{
						StringField: "value 1",
					},
				},
			},
			wantTotalMatches: 1,
		},
		{
			name: "success - with embedding struct and field tags",
			elements: &[]embeddingStruct{
				{
					simpleStruct: simpleStruct{
						StringField: "value 1",
					},
				},
				{
					simpleStruct: simpleStruct{
						StringField: "value 2",
					},
				},
			},
			opts: []Option{
				WithFieldNameTag("json"),
			},
			rules: [][]Rule{{{
				Field: "string_field",
				Type:  TypeEqual,
				Value: "value 1",
			}}},
			want: &[]embeddingStruct{
				{
					simpleStruct: simpleStruct{
						StringField: "value 1",
					},
				},
			},
			wantTotalMatches: 1,
		},
		{
			name:     "success - with root field selector",
			elements: &[]int{41, 42, 43},
			rules: [][]Rule{{{
				Field: "",
				Type:  TypeEqual,
				Value: 42,
			}}},
			want:             &[]int{42},
			wantTotalMatches: 1,
		},
		{
			name:             "success - with empty AND filter",
			elements:         &[]int{41, 42, 43},
			rules:            [][]Rule{},
			want:             &[]int{41, 42, 43},
			wantTotalMatches: 3,
		},
		{
			name:             "success - with empty OR filter",
			elements:         &[]int{41, 42, 43},
			rules:            [][]Rule{{}},
			want:             &[]int{},
			wantTotalMatches: 0,
		},
		{
			name: "success - nested path not found",
			elements: &[]any{
				complexStruct{
					Internal: simpleStruct{
						StringField: "value 1",
					},
				},
			},
			rules: [][]Rule{{{
				Field: "Internal.InvalidField",
				Type:  TypeEqual,
				Value: "value 1",
			}}},
			want:             &[]any{},
			wantTotalMatches: 0,
		},
		{
			name: "success - with field tag not found",
			elements: &[]any{
				complexStruct{
					Internal: simpleStruct{
						StringField: "value 1",
					},
				},
			},
			opts: []Option{WithFieldNameTag("json")},
			rules: [][]Rule{{{
				Field: "internal.invalid_field",
				Type:  TypeEqual,
				Value: "value 1",
			}}},
			want:             &[]any{},
			wantTotalMatches: 0,
		},
		{
			name:             "success - with max results option",
			elements:         &[]string{"1", "2", "3", "4", "5"},
			opts:             []Option{WithMaxResults(3)},
			rules:            [][]Rule{{trueRegex}},
			want:             &[]string{"1", "2", "3"},
			wantTotalMatches: 5,
		},
		{
			name:             "combination - true AND true",
			elements:         &[]string{"a"},
			rules:            [][]Rule{{trueRegex}, {trueRegex}},
			want:             &[]string{"a"},
			wantTotalMatches: 1,
		},
		{
			name:             "combination - true AND false",
			elements:         &[]string{"a"},
			rules:            [][]Rule{{trueRegex}, {falseRegex}},
			want:             &[]string{},
			wantTotalMatches: 0,
		},
		{
			name:             "combination - false AND true",
			elements:         &[]string{"a"},
			rules:            [][]Rule{{falseRegex}, {trueRegex}},
			want:             &[]string{},
			wantTotalMatches: 0,
		},
		{
			name:     "combination - false AND false",
			elements: &[]string{"a"},
			rules:    [][]Rule{{falseRegex}, {falseRegex}},
			want:     &[]string{},
		},
		{
			name:             "combination - true OR false",
			elements:         &[]string{"a"},
			rules:            [][]Rule{{trueRegex, falseRegex}},
			want:             &[]string{"a"},
			wantTotalMatches: 1,
		},
		{
			name:             "combination - true OR true",
			elements:         &[]string{"a"},
			rules:            [][]Rule{{trueRegex, trueRegex}},
			want:             &[]string{"a"},
			wantTotalMatches: 1,
		},
		{
			name:             "combination - false OR true",
			elements:         &[]string{"a"},
			rules:            [][]Rule{{falseRegex, trueRegex}},
			want:             &[]string{"a"},
			wantTotalMatches: 1,
		},
		{
			name:             "combination - false OR false",
			elements:         &[]string{"a"},
			rules:            [][]Rule{{falseRegex, falseRegex}},
			want:             &[]string{},
			wantTotalMatches: 0,
		},
		{
			name:             "combination - (false OR true) AND (true OR false)",
			elements:         &[]string{"a"},
			rules:            [][]Rule{{falseRegex, trueRegex}, {trueRegex, falseRegex}},
			opts:             []Option{WithMaxRules(4)},
			want:             &[]string{"a"},
			wantTotalMatches: 1,
		},
		{
			name:             "combination - (false OR true) AND (false OR false)",
			elements:         &[]string{"a"},
			rules:            [][]Rule{{falseRegex, trueRegex}, {falseRegex, falseRegex}},
			opts:             []Option{WithMaxRules(4)},
			want:             &[]string{},
			wantTotalMatches: 0,
		},
		{
			name:     "error - not a pointer",
			elements: 42,
			rules: [][]Rule{{{
				Type: TypeEqual,
			}}},
			wantErr: true,
		},
		{
			name:     "error - not a slice",
			elements: &simpleStruct{},
			rules: [][]Rule{{{
				Type: TypeEqual,
			}}},
			wantErr: true,
		},
		{
			name: "error - unexported field",
			elements: &[]any{
				complexStruct{
					Internal: simpleStruct{
						unexported: "value 1",
					},
				},
			},
			rules: [][]Rule{{{
				Field: "Internal.unexported",
				Type:  TypeEqual,
				Value: "value 1",
			}}},
			wantErr: true,
		},
		{
			name: "error - with nil item",
			elements: &[]any{
				nil,
			},
			rules: [][]Rule{{{
				Field: "Somefield",
				Type:  TypeEqual,
				Value: "value 1",
			}}},
			wantErr: true,
		},
		{
			name: "error - nested path inside a basic type",
			elements: &[]any{
				simpleStruct{
					StringField: "value 1",
				},
			},
			rules: [][]Rule{{{
				Field: "StringField.InvalidField",
				Type:  TypeEqual,
				Value: "value 1",
			}}},
			wantErr: true,
		},
		{
			name: "error - invalid regex",
			elements: &[]simpleStruct{
				{
					StringField: "value 1",
				},
			},
			rules: [][]Rule{{{
				Field: "StringField",
				Type:  TypeRegexp,
				Value: "(",
			}}},
			wantErr: true,
		},
		{
			name: "error - not a string and regexp",
			elements: &[]simpleStruct{
				{
					StringField: "value 1",
				},
			},
			rules: [][]Rule{{{
				Field: "StringField",
				Type:  TypeRegexp,
				Value: 1,
			}}},
			wantErr: true,
		},
		{
			name: "error - invalid filter type",
			elements: &[]simpleStruct{
				{
					StringField: "value 1",
				},
			},
			rules: [][]Rule{{{
				Field: "StringField",
				Type:  "invalid filter type",
				Value: "value 1",
			}}},
			wantErr: true,
		},
		{
			name:     "error - too many rules",
			elements: &[]int{1, 2, 3},
			rules: [][]Rule{{
				{
					Field: "",
					Type:  TypeEqual,
					Value: 1,
				},
				{
					Field: "",
					Type:  TypeEqual,
					Value: 3,
				},
			}},
			opts:    []Option{WithMaxRules(1)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p, err := New(tt.opts...)
			require.NoError(t, err)

			sliceLen, totalMatches, err := p.Apply(tt.rules, tt.elements)

			if tt.wantErr {
				require.Error(t, err, "Apply() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, tt.elements, "Filtered = %v, want %v", tt.elements, tt.want)
				wantSliceLen := getSliceLen(tt.elements)
				require.Equal(t, wantSliceLen, sliceLen, "Apply() returned sliceLen=%d, want %d", sliceLen, wantSliceLen)
				require.Equal(t, tt.wantTotalMatches, totalMatches, "Apply() returned totalMatches=%d, want %d", totalMatches, tt.wantTotalMatches)
			}
		})
	}
}

func TestFilter_ApplySubset(t *testing.T) {
	t.Parallel()

	trueRegex := Rule{
		Field: "",
		Type:  TypeRegexp,
		Value: ".*",
	}

	tests := []struct {
		name             string
		rules            [][]Rule
		opts             []Option
		elements         any
		offset           uint
		length           uint
		wantTotalMatches uint
		want             any
		wantErr          bool
	}{
		{
			name:             "success - whole slice",
			elements:         &[]string{"1", "2", "3", "4", "5"},
			rules:            [][]Rule{{trueRegex}},
			offset:           0,
			length:           5,
			want:             &[]string{"1", "2", "3", "4", "5"},
			wantTotalMatches: 5,
		},
		{
			name:             "success - contained subset",
			elements:         &[]string{"1", "2", "3", "4", "5"},
			rules:            [][]Rule{{trueRegex}},
			offset:           1,
			length:           3,
			want:             &[]string{"2", "3", "4"},
			wantTotalMatches: 5,
		},
		{
			name:             "success - offset > len(input)",
			elements:         &[]string{"1", "2", "3", "4", "5"},
			rules:            [][]Rule{{trueRegex}},
			offset:           5,
			length:           10,
			want:             &[]string{},
			wantTotalMatches: 5,
		},
		{
			name:             "success - offset in but length out of bounds",
			elements:         &[]string{"1", "2", "3", "4", "5"},
			rules:            [][]Rule{{trueRegex}},
			offset:           3,
			length:           10,
			want:             &[]string{"4", "5"},
			wantTotalMatches: 5,
		},
		{
			name:             "success - no rules with length and offset",
			elements:         &[]string{"1", "2", "3", "4", "5"},
			rules:            [][]Rule{{trueRegex}},
			offset:           2,
			length:           2,
			want:             &[]string{"3", "4"},
			wantTotalMatches: 5,
		},
		{
			name:     "error - length < 1",
			elements: &[]string{"1", "2", "3", "4", "5"},
			rules:    [][]Rule{{trueRegex}},
			offset:   0,
			length:   0,
			wantErr:  true,
		},
		{
			name:     "error - length > p.maxResults",
			elements: &[]string{"1", "2", "3", "4", "5"},
			rules:    [][]Rule{{trueRegex}},
			offset:   0,
			length:   MaxResults + 1,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p, err := New(tt.opts...)
			require.NoError(t, err)

			sliceLen, totalMatches, err := p.ApplySubset(tt.rules, tt.elements, tt.offset, tt.length)

			if tt.wantErr {
				require.Error(t, err, "ApplySubset() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, tt.elements, "Filtered = %v, want %v", tt.elements, tt.want)
				wantSliceLen := getSliceLen(tt.elements)
				require.Equal(t, wantSliceLen, sliceLen, "ApplySubset() returned sliceLen=%d, want %d", sliceLen, wantSliceLen)
				require.Equal(t, tt.wantTotalMatches, totalMatches, "ApplySubset() returned totalMatches=%d, want %d", totalMatches, tt.wantTotalMatches)
			}
		})
	}
}

func benchmarkFilterApply(b *testing.B, n int, json string, opts ...Option) {
	b.Helper()

	type simpleStruct struct {
		IntField     int
		Float64Field float64
		SomeField1   any
		SomeField2   any
		SomeField3   any
		StringField  string `json:"string_field"`
	}

	filter, err := New(opts...)
	require.NoError(b, err)

	data := make([]simpleStruct, n)
	for i := range n {
		data[i] = simpleStruct{
			StringField: "hello world",
		}
	}

	rules, err := ParseJSON(json)
	require.NoError(b, err)

	b.ResetTimer()

	for range b.N {
		b.StopTimer()

		dataCopy := make([]simpleStruct, len(data))
		copy(dataCopy, data)

		b.StartTimer()

		_, _, err := filter.Apply(rules, &dataCopy)
		require.NoError(b, err)
	}
}

func BenchmarkFilter_Apply_Equal_100(b *testing.B) {
	benchmarkFilterApply(
		b,
		100,
		`[[{"field": "StringField", "type": "==", "value": "hello world"}]]`,
	)
}

func BenchmarkFilter_Apply_Equal_1000(b *testing.B) {
	benchmarkFilterApply(
		b,
		1000,
		`[[{"field": "StringField", "type": "==", "value": "hello world"}]]`,
	)
}

func BenchmarkFilter_Apply_Equal_10000(b *testing.B) {
	benchmarkFilterApply(
		b,
		10000,
		`[[{"field": "StringField", "type": "==", "value": "hello world"}]]`,
	)
}

func BenchmarkFilter_Apply_Regexp_1000(b *testing.B) {
	benchmarkFilterApply(
		b,
		1000,
		`[[{"field": "StringField", "type": "regexp", "value": "hello.*"}]]`,
	)
}

func BenchmarkFilter_Apply_WithTagField_1000(b *testing.B) {
	benchmarkFilterApply(
		b,
		1000,
		`[[{"field": "string_field", "type": "==", "value": "hello world"}]]`,
		WithFieldNameTag("json"),
	)
}
