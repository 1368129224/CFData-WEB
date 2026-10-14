package main

import (
	"reflect"
	"testing"
)

func TestOfficialIPTypes(t *testing.T) {
	tests := []struct {
		input int
		want  []int
	}{
		{input: 4, want: []int{4}},
		{input: 6, want: []int{6}},
		{input: 46, want: []int{4, 6}},
		{input: 0, want: nil},
		{input: 64, want: nil},
	}
	for _, tt := range tests {
		if got := officialIPTypes(tt.input); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("officialIPTypes(%d) = %#v, want %#v", tt.input, got, tt.want)
		}
	}
}
