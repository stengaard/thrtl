package main

import (
	"testing"
	"time"
)

func TestByteTime(t *testing.T) {
	tests := []struct {
		kbps      int64
		bytesRead int
		expected  time.Duration
	}{
		{kbps: 3, bytesRead: 768, expected: 250 * time.Millisecond},
		{kbps: 1, bytesRead: 4096, expected: 4 * time.Second},
		{kbps: 1, bytesRead: 1024, expected: time.Second},
		{kbps: 1, bytesRead: 512, expected: 500 * time.Millisecond},
		{kbps: 2, bytesRead: 512, expected: 250 * time.Millisecond},
		{kbps: 200, bytesRead: 1024, expected: 5 * time.Millisecond},
		// less than 1 ms - worth sleeping?
		{kbps: 1024, bytesRead: 1024, expected: 976562 * time.Nanosecond},
		{kbps: 200 * 1024, bytesRead: 1024, expected: 4882 * time.Nanosecond},
	}

	for _, test := range tests {
		res := byteTime(test.kbps, test.bytesRead)
		if res != test.expected {
			t.Errorf("byteTime(%d, %d) == %v, expected %v", test.kbps, test.bytesRead, res, test.expected)
		}
	}

}
