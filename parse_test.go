package main

import (
	"fmt"
	"reflect"
	"testing"

	"toddgaunt.com/grimoire/test"
)

func TestSpellSegmentsSubstitute(t *testing.T) {
	var testCases = []struct {
		name          string
		spellSegments *SpellSegments
		paramValues   map[string]string

		want string
		err  error
	}{
		{
			name: "no parameters",
			spellSegments: &SpellSegments{
				Segments:     []string{"echo Hello World"},
				ParamIndices: []int{},
				Params:       []Param{},
			},
			paramValues: map[string]string{},

			want: "echo Hello World",
		},
		{
			name: "single parameter",
			spellSegments: &SpellSegments{
				Segments:     []string{"echo ", "name"},
				ParamIndices: []int{1},
				Params: []Param{
					{Name: "name", DefaultValues: []string{"World"}},
				},
			},
			paramValues: map[string]string{"name": "Alice"},

			want: "echo Alice",
		},
		{
			name: "multiple parameters",
			spellSegments: &SpellSegments{
				Segments:     []string{"cp ", "source", " ", "destination"},
				ParamIndices: []int{1, 3},
				Params: []Param{
					{Name: "source", DefaultValues: []string{"file.txt"}},
					{Name: "destination", DefaultValues: []string{"backup.txt"}},
				},
			},
			paramValues: map[string]string{"source": "data.csv", "destination": "data_backup.csv"},

			want: "cp data.csv data_backup.csv",
		},
		{
			name: "missing parameter value",
			spellSegments: &SpellSegments{
				Segments:     []string{"mv ", "oldname", " ", "newname"},
				ParamIndices: []int{1, 3},
				Params: []Param{
					{Name: "oldname", DefaultValues: []string{"file1.txt"}},
					{Name: "newname", DefaultValues: []string{"file2.txt"}},
				},
			},
			paramValues: map[string]string{"oldname": "document.txt"},

			want: "", // Expect error due to missing 'newname'
			err:  fmt.Errorf("no value provided for parameter 'newname'"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.spellSegments.Reconstruct(tc.paramValues)

			if !test.ErrorTextEqual(err, tc.err) {
				t.Fatalf("got error %q, want error %q", err, tc.err)
			}

			if result != tc.want {
				t.Errorf("got '%s', want '%s'", result, tc.want)
			}
		})
	}
}

func TestParseSpell(t *testing.T) {
	var testCases = []struct {
		name  string
		spell string

		want *SpellSegments
		err  error
	}{
		{
			name:  "ok - no parameters",
			spell: "echo Hello World",

			want: &SpellSegments{
				Segments:     []string{"echo Hello World"},
				ParamIndices: []int{},
				Params:       []Param{},
			},
		},
		{
			name:  "ok - single parameter",
			spell: "echo <name>",

			want: &SpellSegments{
				Segments:     []string{"echo ", "name"},
				ParamIndices: []int{1},
				Params: []Param{
					{Name: "name", DefaultValues: nil},
				},
			},
		},
		{
			name:  "ok - parameter with default",
			spell: "echo <name=World>",

			want: &SpellSegments{
				Segments:     []string{"echo ", "name"},
				ParamIndices: []int{1},
				Params: []Param{
					{Name: "name", DefaultValues: []string{"World"}},
				},
			},
		},
		{
			name:  "ok - multiple parameters with defaults",
			spell: "cp <source=file.txt> <destination=backup.txt>",

			want: &SpellSegments{
				Segments:     []string{"cp ", "source", " ", "destination"},
				ParamIndices: []int{1, 3},
				Params: []Param{
					{Name: "source", DefaultValues: []string{"file.txt"}},
					{Name: "destination", DefaultValues: []string{"backup.txt"}},
				},
			},
		},
		{
			name:  "ok - parameter with multiple defaults",
			spell: "mv <oldname=file1.txt;file_old.txt> <newname=file2.txt;file_new.txt>",

			want: &SpellSegments{
				Segments:     []string{"mv ", "oldname", " ", "newname"},
				ParamIndices: []int{1, 3},
				Params: []Param{
					{Name: "oldname", DefaultValues: []string{"file1.txt", "file_old.txt"}},
					{Name: "newname", DefaultValues: []string{"file2.txt", "file_new.txt"}},
				},
			},
		},
		{
			name:  "ok - repeated parameters",
			spell: "echo <name> and again <name>",

			want: &SpellSegments{
				Segments:     []string{"echo ", "name", " and again ", "name"},
				ParamIndices: []int{1, 3},
				Params: []Param{
					{Name: "name", DefaultValues: nil},
				},
			},
		},
		{
			name:  "ok - repeated parameter with default only on first occurrence",
			spell: "echo <name=World> and again <name>",

			want: &SpellSegments{
				Segments:     []string{"echo ", "name", " and again ", "name"},
				ParamIndices: []int{1, 3},
				Params: []Param{
					{Name: "name", DefaultValues: []string{"World"}},
				},
			},
		},
		{
			name:  "ok - trailing text following last parameter",
			spell: "echo <name> trailing segment test",

			want: &SpellSegments{
				Segments:     []string{"echo ", "name", " trailing segment test"},
				ParamIndices: []int{1},
				Params: []Param{
					{Name: "name", DefaultValues: nil},
				},
			},
		},
		{
			name:  "error - on repeated parameter with defaults",
			spell: "echo <name=World> and again <name=Everyone>",

			err: fmt.Errorf("parameter 'name' appears multiple times with default values - defaults only allowed on first occurrence"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseSpell(tc.spell)

			if !test.ErrorTextEqual(err, tc.err) {
				t.Fatalf("got error %q, want error %q", err, tc.err)
			}
			if err != nil {
				return
			}

			if !reflect.DeepEqual(result, tc.want) {
				t.Errorf("unexpected result\ngot: %#v\nwant:%#v\ndiff: %s", result, tc.want, test.Diff(result, tc.want))
			}
		})
	}
}
