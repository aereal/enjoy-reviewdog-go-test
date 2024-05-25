package reviewdoggotest_test

import "testing"

func Test_ok(t *testing.T) {
	t.Log("ok")
}

func Test_ng(t *testing.T) {
	t.Error("failing")
}

func Test_skip(t *testing.T) {
	t.SkipNow()
	t.Error("maybe skipped")
}
