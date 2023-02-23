package log

import (
	"context"
	"testing"
)

func TestXxx(t *testing.T) {
	// SetLevel(ErrorLevel)
	Error(context.TODO(), "我问2问")
}

func TestSlog(t *testing.T) {
	Trace(context.TODO(), "33")
}
