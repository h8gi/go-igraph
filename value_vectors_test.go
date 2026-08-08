package igraph

import (
	"reflect"
	"strings"
	"testing"
)

func TestBoolVectorRoundTrip(t *testing.T) {
	for _, values := range [][]bool{nil, {}, {true, false, true}} {
		vector, err := newBoolVector(values)
		if err != nil {
			t.Fatalf("newBoolVector(%v) error = %v", values, err)
		}
		got, err := vector.slice()
		if err != nil {
			t.Fatalf("boolean slice error = %v", err)
		}
		vector.close()

		want := append([]bool{}, values...)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("boolean round trip = %#v, want %#v", got, want)
		}
		if got == nil {
			t.Error("boolean round trip returned a nil slice")
		}
	}
}

func TestStringVectorRoundTrip(t *testing.T) {
	for _, values := range [][]string{nil, {}, {"", "ordinary", "日本語", "🙂"}} {
		vector, err := newStringVector(values)
		if err != nil {
			t.Fatalf("newStringVector(%q) error = %v", values, err)
		}
		got, err := vector.slice()
		if err != nil {
			t.Fatalf("string slice error = %v", err)
		}
		vector.close()

		want := append([]string{}, values...)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("string round trip = %#v, want %#v", got, want)
		}
		if got == nil {
			t.Error("string round trip returned a nil slice")
		}
	}
}

func TestStringVectorRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		value string
		match string
	}{
		{name: "embedded NUL", value: "a\x00b", match: "embedded NUL"},
		{name: "invalid UTF-8", value: string([]byte{0xff}), match: "valid UTF-8"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vector, err := newStringVector([]string{"valid", tt.value})
			if vector != nil || err == nil || !strings.Contains(err.Error(), tt.match) {
				t.Errorf("newStringVector() = %v, %v, want nil error containing %q", vector, err, tt.match)
			}
		})
	}
}

func TestValueVectorResultsOwnTheirStorage(t *testing.T) {
	booleans, err := newBoolVector([]bool{true})
	if err != nil {
		t.Fatal(err)
	}
	boolResult, err := booleans.slice()
	if err != nil {
		t.Fatal(err)
	}
	booleans.close()
	if !reflect.DeepEqual(boolResult, []bool{true}) {
		t.Fatalf("boolean result after close = %v", boolResult)
	}

	strings, err := newStringVector([]string{"owned"})
	if err != nil {
		t.Fatal(err)
	}
	stringResult, err := strings.slice()
	if err != nil {
		t.Fatal(err)
	}
	strings.close()
	if !reflect.DeepEqual(stringResult, []string{"owned"}) {
		t.Fatalf("string result after close = %v", stringResult)
	}
}
