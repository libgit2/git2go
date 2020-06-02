package git

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTrailers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected []Trailer
	}{
		{
			"commit with zero trailers\n",
			[]Trailer{},
		},
		{
			"commit with one trailer\n\nCo-authored-by: Alice <alice@example.com>\n",
			[]Trailer{
				Trailer{Key: "Co-authored-by", Value: "Alice <alice@example.com>"},
			},
		},
		{
			"commit with two trailers\n\nCo-authored-by: Alice <alice@example.com>\nSigned-off-by: Bob <bob@example.com>\n",
			[]Trailer{
				Trailer{Key: "Co-authored-by", Value: "Alice <alice@example.com>"},
				Trailer{Key: "Signed-off-by", Value: "Bob <bob@example.com>"}},
		},
	}
	for _, test := range tests {
		fmt.Printf("%s", test.input)
		actual, err := MessageTrailers(test.input)
		if err != nil {
			t.Errorf("Trailers returned an unexpected error: %v", err)
		}
		if !reflect.DeepEqual(test.expected, actual) {
			t.Errorf("expecting %#v\ngot %#v", test.expected, actual)
		}
	}
}
