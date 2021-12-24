package result

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestOptionUnwrap(t *testing.T) {
	type test struct {
		input         *Option[string]
		expectError   bool
		expectedValue string
	}

	notSetOption := NewOption[string]()
	someOption := NewOption[string]().SetSome("some")
	noneOption := NewOption[string]().SetNone()
	errOption := NewOption[string]().SetError(errors.New("error"))

	tests := []test{
		{input: notSetOption, expectError: true, expectedValue: ""},
		{input: someOption, expectError: false, expectedValue: "some"},
		{input: noneOption, expectError: true, expectedValue: ""},
		{input: errOption, expectError: true, expectedValue: ""},
	}

	for _, tc := range tests {
		gotValue, gotError := tc.input.Unwrap()
		if gotError != nil && !tc.expectError {
			t.Fatalf("expected no error, got error %v", gotError.Error())
		} else if gotError == nil && tc.expectError {
			t.Fatalf("expected error, got no error")
		}

		if gotValue != tc.expectedValue {
			t.Fatalf("expected %v, got %v", tc.expectedValue, gotValue)
		}
	}
}

func TestOptionWait(t *testing.T) {
	option := NewOption[string]()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var waitSuccess bool

	wg := sync.WaitGroup{}
	wg.Add(1)

	var value string
	var err error

	go func() {
		defer wg.Done()

		value, err = option.Wait(ctx)
		waitSuccess = true
	}()

	option.SetSome("foo")

	wg.Wait()

	if !waitSuccess {
		t.Fatal("wait failed")
	}

	if err != nil {
		t.Fatalf("wait gave unexpected error: %v", err.Error())
	}

	if value != "foo" {
		t.Fatalf("wait gave unexpected value: %v", value)
	}

	type test struct {
		input         *Option[string]
		kind          OptionKind
		expectedValue string
		expectError   bool
	}

	someOption := NewOption[string]()
	noneOption := NewOption[string]()
	errOption := NewOption[string]()
	neverOption := NewOption[string]()

	tests := []test{
		{input: someOption, kind: Some, expectError: false, expectedValue: "some"},
		{input: noneOption, kind: None, expectError: true, expectedValue: ""},
		{input: errOption, kind: Error, expectError: true, expectedValue: ""},
		{input: neverOption, kind: never, expectError: true, expectedValue: ""},
	}

	for _, tc := range tests {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		var waitSuccess bool

		wg := sync.WaitGroup{}
		wg.Add(1)

		var gotValue string
		var gotError error

		go func(opt *Option[string]) {
			defer wg.Done()

			gotValue, gotError = opt.Wait(ctx)
			waitSuccess = true
		}(tc.input)

		switch tc.kind {
		case Some:
			tc.input.SetSome(tc.expectedValue)
		case None:
			tc.input.SetNone()
		case Error:
			tc.input.SetError(fmt.Errorf("error"))
		case never:
			// nothing ever set, let wait context fail
		}

		wg.Wait()

		if !waitSuccess {
			t.Fatal("wait failed")
		}

		if gotError != nil && !tc.expectError {
			t.Fatalf("expected no error, got error %v", gotError.Error())
		} else if gotError == nil && tc.expectError {
			t.Fatalf("expected error, got no error")
		}

		if gotValue != tc.expectedValue {
			t.Fatalf("expected %v, got %v", tc.expectedValue, gotValue)
		}
	}
}

func TestOptionKind(t *testing.T) {
	type test struct {
		input *Option[string]
		want  OptionKind
	}

	notSetOption := NewOption[string]()
	someOption := NewOption[string]().SetSome("some")
	noneOption := NewOption[string]().SetNone()
	errOption := NewOption[string]().SetError(errors.New("error"))

	tests := []test{
		{input: notSetOption, want: NotSet},
		{input: someOption, want: Some},
		{input: noneOption, want: None},
		{input: errOption, want: Error},
	}

	for _, tc := range tests {
		got := tc.input.Kind()
		if got != tc.want {
			t.Fatalf("expected %v, got %v", tc.want, got)
		}
	}
}
