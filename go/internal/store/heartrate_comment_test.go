package store

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestNormalizeHeartRateComment(t *testing.T) {
	t.Parallel()

	ptr := func(s string) *string { return &s }

	t.Run("nil", func(t *testing.T) {
		if got := NormalizeHeartRateComment(nil); got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("blank", func(t *testing.T) {
		if got := NormalizeHeartRateComment(ptr("  \n\t ")); got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("trim", func(t *testing.T) {
		got := NormalizeHeartRateComment(ptr("  agité ce matin  "))
		if got == nil || *got != "agité ce matin" {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("truncate runes", func(t *testing.T) {
		long := strings.Repeat("é", MaxHeartRateCommentLen+10)
		got := NormalizeHeartRateComment(ptr(long))
		if got == nil {
			t.Fatal("expected truncated comment")
		}
		if n := utf8.RuneCountInString(*got); n != MaxHeartRateCommentLen {
			t.Fatalf("rune count = %d, want %d", n, MaxHeartRateCommentLen)
		}
	})
}
