package email_test

import (
	"testing"

	"github.com/jansemmelink/events/email"
)

func TestValidate(t *testing.T) {
	validList := []string{
		"a@b.c",
		"a.b@b.c",
		"a-b@b.c",
		"a-b.c@b.c",
		"a-b.c-d@b.c",
		"a@b-c.d",
		"a0-1s.1a-b4@123.345",
	}
	invalidList := []string{
		"a@b",
	}
	errorCount := 0
	for i, e := range validList {
		if s, err := email.Valid(e); err != nil {
			errorCount++
			t.Errorf("failed to validate valid[%d]: %s: %+v", i, e, err)
		} else {
			t.Logf("valid: %s", s)
		}
	}
	if errorCount != 0 {
		t.Fatalf("%d valid emails could not validate", errorCount)
	}

	errorCount = 0
	for i, e := range invalidList {
		if _, err := email.Valid(e); err == nil {
			errorCount++
			t.Errorf("validated invalid[%d]: %s: %+v", i, e, err)
		} else {
			t.Logf("invalid: %s", e)
		}
	}
	if errorCount != 0 {
		t.Fatalf("%d invalid emails were validated", errorCount)
	}
}
