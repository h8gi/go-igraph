package igraph

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestAttributeScopeAndTypeValidation(t *testing.T) {
	for _, scope := range []AttributeScope{AttributeGraph, AttributeVertex, AttributeEdge} {
		if _, err := scope.cValue(); err != nil {
			t.Errorf("scope %d: %v", scope, err)
		}
	}
	if _, err := AttributeScope(255).cValue(); err == nil {
		t.Error("invalid attribute scope accepted")
	}

	for _, attributeType := range []AttributeType{AttributeNumeric, AttributeBoolean, AttributeString} {
		cType, err := attributeType.cValue()
		if err != nil {
			t.Errorf("type %d: %v", attributeType, err)
			continue
		}
		converted, err := attributeTypeFromC(int(cType))
		if err != nil || converted != attributeType {
			t.Errorf("type %d round trip = %d, %v", attributeType, converted, err)
		}
	}
	if _, err := AttributeType(0).cValue(); err == nil {
		t.Error("zero attribute type accepted")
	}
	if _, err := attributeTypeFromC(127); err == nil {
		t.Error("object attribute type accepted")
	}
}

func TestAttributeStringValidation(t *testing.T) {
	if err := validateAttributeName("name"); err != nil {
		t.Fatalf("valid name: %v", err)
	}
	if err := validateAttributeString("value", ""); err != nil {
		t.Fatalf("empty value: %v", err)
	}

	for name, value := range map[string]string{
		"empty": "",
		"NUL":   "a\x00b",
		"UTF-8": string([]byte{0xff}),
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateAttributeName(value); err == nil {
				t.Fatalf("validateAttributeName(%q) error = nil", value)
			}
		})
	}
	if err := validateAttributeString("value", "a\x00b"); err == nil {
		t.Error("embedded NUL value accepted")
	}
	if err := validateAttributeString("value", string([]byte{0xff})); err == nil {
		t.Error("invalid UTF-8 value accepted")
	}
}

func TestAttributeMetadataConversion(t *testing.T) {
	types := make([]int, 3)
	for i, attributeType := range []AttributeType{AttributeBoolean, AttributeNumeric, AttributeString} {
		cType, err := attributeType.cValue()
		if err != nil {
			t.Fatal(err)
		}
		types[i] = int(cType)
	}
	metadata, err := attributeMetadataFromSlices(
		AttributeVertex,
		[]string{"active", "score", "label"},
		types,
	)
	if err != nil {
		t.Fatalf("attributeMetadataFromSlices() error = %v", err)
	}
	if len(metadata) != 3 || metadata[0].Scope != AttributeVertex || metadata[2].Type != AttributeString {
		t.Fatalf("metadata = %#v", metadata)
	}

	empty, err := attributeMetadataFromSlices(AttributeEdge, nil, nil)
	if err != nil {
		t.Fatalf("empty metadata error = %v", err)
	}
	if empty == nil || len(empty) != 0 {
		t.Fatalf("empty metadata = %#v, want non-nil empty", empty)
	}
}

func TestAttributeMetadataConversionFailures(t *testing.T) {
	numeric, err := AttributeNumeric.cValue()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name  string
		scope AttributeScope
		names []string
		types []int
	}{
		{name: "scope", scope: 99},
		{name: "length", names: []string{"one"}},
		{name: "empty name", names: []string{""}, types: []int{int(numeric)}},
		{name: "duplicate", names: []string{"same", "same"}, types: []int{int(numeric), int(numeric)}},
		{name: "unsupported type", names: []string{"object"}, types: []int{127}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metadata, err := attributeMetadataFromSlices(test.scope, test.names, test.types)
			if err == nil || metadata != nil {
				t.Fatalf("metadata = %#v, error = %v", metadata, err)
			}
		})
	}
}

func TestAttributeRuntimeIsInstalledOnce(t *testing.T) {
	if !attributeRuntimeInstalled() {
		t.Fatal("C attribute runtime is not installed")
	}
	for range 4 {
		if err := ensureAttributeRuntime(); err != nil {
			t.Fatalf("repeated ensureAttributeRuntime() error = %v", err)
		}
	}

	var state attributeRuntimeState
	var calls atomic.Int32
	start := make(chan struct{})
	var wait sync.WaitGroup
	for range 32 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			if err := ensureAttributeRuntimeWithInstaller(&state, func() int {
				calls.Add(1)
				return 0
			}); err != nil {
				t.Errorf("concurrent setup error = %v", err)
			}
		}()
	}
	close(start)
	wait.Wait()
	if calls.Load() != 1 {
		t.Fatalf("installer calls = %d, want 1", calls.Load())
	}
}

func TestAttributeRuntimeSetupFailureIsSticky(t *testing.T) {
	var state attributeRuntimeState
	var calls int
	install := func() int {
		calls++
		return 1
	}
	first := ensureAttributeRuntimeWithInstaller(&state, install)
	second := ensureAttributeRuntimeWithInstaller(&state, install)
	if first == nil || second == nil || first.Error() != second.Error() {
		t.Fatalf("setup errors = %v, %v", first, second)
	}
	if calls != 1 {
		t.Fatalf("installer calls = %d, want 1", calls)
	}

	var nilState attributeRuntimeState
	if err := ensureAttributeRuntimeWithInstaller(&nilState, nil); err == nil {
		t.Error("nil installer error = nil")
	}
}

func TestAttributeRecordLifecycle(t *testing.T) {
	for _, attributeType := range []AttributeType{AttributeNumeric, AttributeBoolean, AttributeString} {
		t.Run(string(rune('0'+attributeType)), func(t *testing.T) {
			record, err := newAttributeRecord("value", attributeType)
			if err != nil {
				t.Fatalf("newAttributeRecord() error = %v", err)
			}
			if err := record.checkType(attributeType); err != nil {
				t.Fatalf("checkType() error = %v", err)
			}
			if size, err := record.size(); err != nil || size != 0 {
				t.Fatalf("initial size = %d, %v", size, err)
			}
			if err := record.resize(3); err != nil {
				t.Fatalf("resize() error = %v", err)
			}
			if size, err := record.size(); err != nil || size != 3 {
				t.Fatalf("resized size = %d, %v", size, err)
			}
			if err := record.resize(-1); err == nil {
				t.Error("negative resize accepted")
			}

			otherType := AttributeNumeric
			if attributeType == otherType {
				otherType = AttributeString
			}
			if err := record.checkType(otherType); err == nil {
				t.Error("wrong record type accepted")
			}
			record.close()
			record.close()
			if _, err := record.size(); !errors.Is(err, ErrClosed) {
				t.Fatalf("closed record size error = %v", err)
			}
		})
	}
}

func TestAttributeRecordInitializationFailures(t *testing.T) {
	called := false
	fail := func(*attributeRecord, string, AttributeType) int {
		called = true
		return 1
	}
	if record, err := newAttributeRecordWithInitializer("name", AttributeString, fail); err == nil || record != nil {
		t.Fatalf("failed record = %#v, %v", record, err)
	}
	if !called {
		t.Fatal("failure initializer was not called")
	}

	for name, attributeType := range map[string]AttributeType{
		"":      AttributeString,
		"valid": 0,
	} {
		called = false
		if record, err := newAttributeRecordWithInitializer(name, attributeType, fail); err == nil || record != nil {
			t.Fatalf("invalid record = %#v, %v", record, err)
		}
		if called {
			t.Fatal("initializer called for invalid input")
		}
	}
	if record, err := newAttributeRecordWithInitializer("name", AttributeString, nil); err == nil || record != nil {
		t.Fatalf("nil initializer record = %#v, %v", record, err)
	}
}

func TestAttributeRecordListLifecycleAndFailures(t *testing.T) {
	list, err := newAttributeRecordList(2)
	if err != nil {
		t.Fatalf("newAttributeRecordList() error = %v", err)
	}
	if size, err := list.size(); err != nil || size != 2 {
		t.Fatalf("list size = %d, %v", size, err)
	}
	list.close()
	list.close()
	if _, err := list.size(); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed list size error = %v", err)
	}

	for _, size := range []int{-1} {
		if failed, err := newAttributeRecordList(size); err == nil || failed != nil {
			t.Fatalf("invalid list = %#v, %v", failed, err)
		}
	}
	if failed, err := newAttributeRecordListWithInitializer(0, nil); err == nil || failed != nil {
		t.Fatalf("nil initializer list = %#v, %v", failed, err)
	}
	if failed, err := newAttributeRecordListWithInitializer(
		0,
		func(*attributeRecordList, int) int { return 1 },
	); err == nil || failed != nil {
		t.Fatalf("failed list = %#v, %v", failed, err)
	}
}

func TestAttributeValidationErrorsHaveContext(t *testing.T) {
	if err := validateAttributeString("vertex label", "bad\x00value"); err == nil || !strings.Contains(err.Error(), "vertex label") {
		t.Fatalf("validation error = %v", err)
	}
}
