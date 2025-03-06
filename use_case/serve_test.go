package use_case

import (
	"github.com/gone-io/gone"
	"testing"
	"time"
)

type testDaemon struct {
	gone.Flag
	isStart bool
	isStop  bool
}

func (t *testDaemon) Start() error {
	println("testDaemon Start")
	t.isStart = true
	return nil
}

func (t *testDaemon) Stop() error {
	println("testDaemon Stop")
	t.isStop = true
	return nil
}

func TestServe(t *testing.T) {
	daemon := &testDaemon{}
	var t1, t2 time.Time
	ins := gone.
		Prepare()

	ins.
		Load(daemon).
		BeforeStart(func() {
			go func() {
				time.Sleep(5 * time.Millisecond)
				t1 = time.Now()
				ins.End()
			}()
		}).
		AfterStop(func() {
			t2 = time.Now()
		}).
		Serve()

	if !daemon.isStart {
		t.Fatal("daemon start failed")
	}
	if !daemon.isStop {
		t.Fatal("daemon stop failed")
	}
	if !t2.After(t1) {
		t.Fatal("daemon stop after serve failed")
	}
}
