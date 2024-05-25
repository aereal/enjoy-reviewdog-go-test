package reviewdoggotest_test

import "testing"

func Test_ok(t *testing.T) {
	t.Log("ok")
}

func Test_ng(t *testing.T) {
	t.Error("failing")
}

func Test_ng_multiline(t *testing.T) {
	t.Error("failing line1\nline2")
}

func Test_skipnow(t *testing.T) {
	t.SkipNow()
}

func Test_skip(t *testing.T) {
	t.Skip("skipped")
}
