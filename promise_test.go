package result

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestPromiseWait(t *testing.T) {
	type test struct {
		promiseFunc   func(ctx context.Context) (string, error)
		ctx           context.Context
		waitCtx       context.Context
		expectedValue string
		expectError   bool
	}

	normalFunc := func(ctx context.Context) (string, error) {
		return "normal", nil
	}

	errorFunc := func(ctx context.Context) (string, error) {
		return "", fmt.Errorf("error")
	}

	ctxFunc := func(ctx context.Context) (string, error) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	canceledContext, cancel := context.WithCancel(context.Background())
	cancel()

	neverFunc := func(ctx context.Context) (string, error) {
		select {
		case <-ctx.Done():

		}

		return "never", nil
	}

	timeoutContext, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	tests := []test{
		{promiseFunc: normalFunc, ctx: context.Background(), waitCtx: context.Background(), expectedValue: "normal", expectError: false},
		{promiseFunc: errorFunc, ctx: context.Background(), waitCtx: context.Background(), expectedValue: "", expectError: true},
		{promiseFunc: ctxFunc, ctx: canceledContext, waitCtx: context.Background(), expectedValue: "", expectError: true},
		{promiseFunc: neverFunc, ctx: context.Background(), waitCtx: timeoutContext, expectedValue: "", expectError: true},
	}

	for _, tc := range tests {
		p := NewPromise(tc.ctx, tc.promiseFunc)

		wg := sync.WaitGroup{}
		wg.Add(1)

		var gotValue string
		var gotError error

		go func(p *Promise[string]) {
			defer wg.Done()
			gotValue, gotError = p.Wait(tc.waitCtx)
		}(p)

		wg.Wait()

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
