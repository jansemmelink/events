package email_test

import (
	"testing"

	"github.com/jansemmelink/events/email"
)

func Test1(t *testing.T) {
	msg := email.Message{
		From:        email.Email{"jan.semmelink@gmail.com", "Jan"},
		To:          []email.Email{{"jan.semmelink@gmail.com", "Jan"}},
		Cc:          []email.Email{{"jan.semmelink@gmail.com", "Jan"}},
		Subject:     "Test 1-2-3",
		ContentType: "text/html",
		Content:     "<H1>Testing Header</H1><P>P1...</P>",
	}
	if err := email.Send(msg); err != nil {
		t.Fatal(err)
	}
}
