package ctxutil_test

import (
	"context"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"

	"gopkg.in/ctxutil.v1"
)

func TestJoinCancel1(t *testing.T) {
	c := qt.New(t)
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx := ctxutil.Join(ctx1, context.Background())
	cancel1()
	waitFor(c, ctx.Done())
	c.Assert(ctx.Err(), qt.Equals, context.Canceled)
}

func TestJoinCancel2(t *testing.T) {
	c := qt.New(t)
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx := ctxutil.Join(context.Background(), ctx1)
	cancel1()
	waitFor(c, ctx.Done())
	c.Assert(ctx.Err(), qt.Equals, context.Canceled)
}

func TestJoinCancelBoth1(t *testing.T) {
	c := qt.New(t)
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	ctx := ctxutil.Join(ctx1, ctx2)
	cancel1()
	waitFor(c, ctx.Done())
	c.Assert(ctx.Err(), qt.Equals, context.Canceled)
}

func TestErrNoErr(t *testing.T) {
	c := qt.New(t)
	ctx := ctxutil.Join(context.Background(), context.Background())
	c.Assert(ctx.Err(), qt.Equals, nil)
}

func TestJoinCancelBoth2(t *testing.T) {
	c := qt.New(t)
	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	ctx2, cancel2 := context.WithCancel(context.Background())
	ctx := ctxutil.Join(ctx1, ctx2)
	cancel2()
	waitFor(c, ctx.Done())
	c.Assert(ctx.Err(), qt.Equals, context.Canceled)
}

func TestDeadline1(t *testing.T) {
	c := qt.New(t)
	tt := time.Now().Add(5 * time.Second).UTC()
	ctx1, cancel1 := context.WithDeadline(context.Background(), tt)
	defer cancel1()
	ctx := ctxutil.Join(ctx1, context.Background())
	deadline, ok := ctx.Deadline()
	c.Assert(ok, qt.Equals, true)
	c.Assert(deadline, qt.Equals, tt)
}

func TestDeadline2(t *testing.T) {
	c := qt.New(t)
	tt := time.Now().Add(5 * time.Second).UTC()
	ctx1, cancel1 := context.WithDeadline(context.Background(), tt)
	defer cancel1()
	ctx := ctxutil.Join(context.Background(), ctx1)
	deadline, ok := ctx.Deadline()
	c.Assert(ok, qt.Equals, true)
	c.Assert(deadline, qt.Equals, tt)
}

func TestDeadlineBoth1(t *testing.T) {
	c := qt.New(t)
	t1 := time.Now().Add(5 * time.Second).UTC()
	ctx1, cancel1 := context.WithDeadline(context.Background(), t1)
	defer cancel1()

	t2 := time.Now().Add(10 * time.Second).UTC()
	ctx2, cancel2 := context.WithDeadline(context.Background(), t2)
	defer cancel2()

	ctx := ctxutil.Join(ctx1, ctx2)

	deadline, ok := ctx.Deadline()
	c.Assert(ok, qt.Equals, true)
	c.Assert(deadline, qt.Equals, t1)
}

func TestDeadlineBoth2(t *testing.T) {
	c := qt.New(t)
	t1 := time.Now().Add(5 * time.Second).UTC()
	ctx1, cancel1 := context.WithDeadline(context.Background(), t1)
	defer cancel1()

	t2 := time.Now().Add(10 * time.Second).UTC()
	ctx2, cancel2 := context.WithDeadline(context.Background(), t2)
	defer cancel2()

	ctx := ctxutil.Join(ctx2, ctx1)

	deadline, ok := ctx.Deadline()
	c.Assert(ok, qt.Equals, true)
	c.Assert(deadline, qt.Equals, t1)
}

func TestValue1(t *testing.T) {
	c := qt.New(t)
	ctx1 := context.WithValue(context.Background(), "foo", "bar")
	ctx := ctxutil.Join(ctx1, context.Background())
	c.Assert(ctx.Value("foo"), qt.Equals, "bar")
}

func TestValue2(t *testing.T) {
	c := qt.New(t)
	ctx1 := context.WithValue(context.Background(), "foo", "bar")
	ctx := ctxutil.Join(context.Background(), ctx1)
	c.Assert(ctx.Value("foo"), qt.Equals, "bar")
}

func TestValueBoth(t *testing.T) {
	c := qt.New(t)
	ctx1 := context.WithValue(context.Background(), "foo", "bar1")
	ctx2 := context.WithValue(context.Background(), "foo", "bar2")
	ctx := ctxutil.Join(ctx1, ctx2)
	c.Assert(ctx.Value("foo"), qt.Equals, "bar1")
}

func TestDoneRace(t *testing.T) {
	c := qt.New(t)
	// This test is designed to be run with the race detector enabled.
	ctx1, cancel1 := context.WithDeadline(context.Background(), time.Now())
	defer cancel1()
	ctx2, cancel2 := context.WithDeadline(context.Background(), time.Now())
	defer cancel2()
	ctx := ctxutil.Join(ctx1, ctx2)
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		done <- struct{}{}
	}()
	go func() {
		<-ctx.Done()
		done <- struct{}{}
	}()
	waitFor(c, done)
	waitFor(c, done)
}

func TestErrRace(t *testing.T) {
	c := qt.New(t)
	// This test is designed to be run with the race detector enabled.
	ctx1, cancel1 := context.WithDeadline(context.Background(), time.Now())
	defer cancel1()
	ctx2, cancel2 := context.WithDeadline(context.Background(), time.Now())
	defer cancel2()
	ctx := ctxutil.Join(ctx1, ctx2)
	done := make(chan struct{})
	go func() {
		ctx.Err()
		done <- struct{}{}
	}()
	go func() {
		ctx.Err()
		done <- struct{}{}
	}()
	waitFor(c, done)
	waitFor(c, done)
}

func waitFor(c *qt.C, ch <-chan struct{}) {
	select {
	case <-ch:
		return
	case <-time.After(time.Second):
		c.Fatalf("timed out")
	}
}
