package test

import (
	"fmt"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// Diff generates a diff text between two objects
func Diff(a, b interface{}) string {
	got := fmt.Sprintf("%#v", a)
	want := fmt.Sprintf("%#v", b)
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(got, want, false)
	return dmp.DiffPrettyText(diffs)
}
