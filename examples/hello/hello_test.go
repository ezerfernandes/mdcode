package main

import (
	"bytes"
	"testing"
)

func Test_main(t *testing.T) {
	var buf bytes.Buffer
	out = &buf

	main()

	const expected = "Hello, Testable World!\n"

	if buf.String() != expected {
		t.Errorf("\nexpected: %s\nactual:   %s\n", expected, buf.String())
	}
}
