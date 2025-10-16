package main

import "testing"

func TestSanitizeFilename(t *testing.T) {
	var testCases = []struct {
		name string
		in   string

		want string
	}{
		{
			name: "simple name",
			in:   "HelloWorld",
			want: "helloworld",
		},
		{
			name: "name with spaces",
			in:   "Hello World",
			want: "hello_world",
		},
		{
			name: "name with special characters",
			in:   "Hello@World!",
			want: "helloworld",
		},
		{
			name: "name with mixed characters",
			in:   "My File-Name_123",
			want: "my_file-name_123",
		},
		{
			name: "name with multiple spaces and special characters",
			in:   "  My  File @ Name!  ",
			want: "my__file_name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := SanitizeFilename(tc.in)
			if got != tc.want {
				t.Errorf("got '%s', want '%s'", got, tc.want)
			}
		})
	}
}
