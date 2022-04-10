package channels_test

import (
	"testing"

	"github.com/joshchoo/go-sandbox/channels"
)

func TestMergeChannels(t *testing.T) {
	ch1 := make(chan string)
	ch2 := make(chan string)

	mergedCh := channels.MergeChannelsString(ch1, ch2)

	msg1 := "hello"
	msg2 := "world"

	go func() {
		ch1 <- msg1
		close(ch1)
	}()
	go func() {
		ch2 <- msg2
		close(ch2)
	}()

	var messages []string
	for msg := range mergedCh {
		messages = append(messages, msg)
	}

	if len(messages) != 2 {
		t.Errorf("Expected to receive %q messages, got %q.", 2, len(messages))
	}
	for _, msg := range messages {
		if msg != msg1 && msg != msg2 {
			t.Errorf("Expected to receive message %q or %q, got %q.", msg1, msg2, msg)
		}
	}
}
