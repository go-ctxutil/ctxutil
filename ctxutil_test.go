package ctxutil_test

import (
	"time"

	"golang.org/x/net/context"
	gc "gopkg.in/check.v1"

	"github.com/sdboyer/constext"
)

type ctxutilSuite struct {
}

var join = constext.Cons

var _ = gc.Suite(&ctxutilSuite{})

func (s *ctxutilSuite) TestJoinCancel1(c *gc.C) {
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx, ctxCancel := join(ctx1, context.Background())
	defer ctxCancel()
	cancel1()
	waitFor(c, ctx.Done())
	c.Assert(ctx.Err(), gc.Equals, context.Canceled)
}

func (s *ctxutilSuite) TestJoinCancel2(c *gc.C) {
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx, ctxCancel := join(context.Background(), ctx1)
	defer ctxCancel()
	cancel1()
	waitFor(c, ctx.Done())
	c.Assert(ctx.Err(), gc.Equals, context.Canceled)
}

func (s *ctxutilSuite) TestJoinCancelBoth1(c *gc.C) {
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	ctx, ctxCancel := join(ctx1, ctx2)
	defer ctxCancel()
	cancel1()
	waitFor(c, ctx.Done())
	c.Assert(ctx.Err(), gc.Equals, context.Canceled)
}

func (s *ctxutilSuite) TestErrNoErr(c *gc.C) {
	ctx, ctxCancel := join(context.Background(), context.Background())
	defer ctxCancel()
	c.Assert(ctx.Err(), gc.Equals, nil)
}

func (s *ctxutilSuite) TestJoinCancelBoth2(c *gc.C) {
	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	ctx2, cancel2 := context.WithCancel(context.Background())
	ctx, ctxCancel := join(ctx1, ctx2)
	defer ctxCancel()
	cancel2()
	waitFor(c, ctx.Done())
	c.Assert(ctx.Err(), gc.Equals, context.Canceled)
}

func (s *ctxutilSuite) TestDeadline1(c *gc.C) {
	t := time.Now().Add(5 * time.Second).UTC()
	ctx1, cancel1 := context.WithDeadline(context.Background(), t)
	defer cancel1()
	ctx, ctxCancel := join(ctx1, context.Background())
	defer ctxCancel()
	deadline, ok := ctx.Deadline()
	c.Assert(ok, gc.Equals, true)
	c.Assert(deadline, gc.Equals, t)
}

func (s *ctxutilSuite) TestDeadline2(c *gc.C) {
	t := time.Now().Add(5 * time.Second).UTC()
	ctx1, cancel1 := context.WithDeadline(context.Background(), t)
	defer cancel1()
	ctx, ctxCancel := join(context.Background(), ctx1)
	defer ctxCancel()
	deadline, ok := ctx.Deadline()
	c.Assert(ok, gc.Equals, true)
	c.Assert(deadline, gc.Equals, t)
}

func (s *ctxutilSuite) TestDeadlineBoth1(c *gc.C) {
	t1 := time.Now().Add(5 * time.Second).UTC()
	ctx1, cancel1 := context.WithDeadline(context.Background(), t1)
	defer cancel1()

	t2 := time.Now().Add(10 * time.Second).UTC()
	ctx2, cancel2 := context.WithDeadline(context.Background(), t2)
	defer cancel2()

	ctx, ctxCancel := join(ctx1, ctx2)
	defer ctxCancel()

	deadline, ok := ctx.Deadline()
	c.Assert(ok, gc.Equals, true)
	c.Assert(deadline, gc.Equals, t1)
}

func (s *ctxutilSuite) TestDeadlineBoth2(c *gc.C) {
	t1 := time.Now().Add(5 * time.Second).UTC()
	ctx1, cancel1 := context.WithDeadline(context.Background(), t1)
	defer cancel1()

	t2 := time.Now().Add(10 * time.Second).UTC()
	ctx2, cancel2 := context.WithDeadline(context.Background(), t2)
	defer cancel2()

	ctx, ctxCancel := join(ctx2, ctx1)
	defer ctxCancel()

	deadline, ok := ctx.Deadline()
	c.Assert(ok, gc.Equals, true)
	c.Assert(deadline, gc.Equals, t1)
}

func (s *ctxutilSuite) TestValue1(c *gc.C) {
	ctx1 := context.WithValue(context.Background(), "foo", "bar")
	ctx, ctxCancel := join(ctx1, context.Background())
	defer ctxCancel()
	c.Assert(ctx.Value("foo"), gc.Equals, "bar")
}

func (s *ctxutilSuite) TestValue2(c *gc.C) {
	ctx1 := context.WithValue(context.Background(), "foo", "bar")
	ctx, ctxCancel := join(context.Background(), ctx1)
	defer ctxCancel()
	c.Assert(ctx.Value("foo"), gc.Equals, "bar")
}

func (s *ctxutilSuite) TestValueBoth(c *gc.C) {
	ctx1 := context.WithValue(context.Background(), "foo", "bar1")
	ctx2 := context.WithValue(context.Background(), "foo", "bar2")
	ctx, ctxCancel := join(ctx1, ctx2)
	defer ctxCancel()
	c.Assert(ctx.Value("foo"), gc.Equals, "bar1")
}

func (s *ctxutilSuite) TestDoneRace(c *gc.C) {
	// This test is designed to be run with the race detector enabled.
	ctx1, cancel1 := context.WithDeadline(context.Background(), time.Now())
	defer cancel1()
	ctx2, cancel2 := context.WithDeadline(context.Background(), time.Now())
	defer cancel2()
	ctx, ctxCancel := join(ctx1, ctx2)
	defer ctxCancel()
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

func (s *ctxutilSuite) TestErrRace(c *gc.C) {
	// This test is designed to be run with the race detector enabled.
	ctx1, cancel1 := context.WithDeadline(context.Background(), time.Now())
	defer cancel1()
	ctx2, cancel2 := context.WithDeadline(context.Background(), time.Now())
	defer cancel2()
	ctx, ctxCancel := join(ctx1, ctx2)
	defer ctxCancel()
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

func waitFor(c *gc.C, ch <-chan struct{}) {
	select {
	case <-ch:
		return
	case <-time.After(time.Second):
		c.Fatalf("timed out")
	}
}
