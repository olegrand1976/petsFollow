package gemini

import (
	"context"
	"testing"
)

func TestEmbedTexts_NotConfigured(t *testing.T) {
	c := &Client{}
	_, err := c.EmbedTexts(context.Background(), []string{"hello"})
	if err == nil || err.Error() != "gemini_not_configured" {
		t.Fatalf("want gemini_not_configured got %v", err)
	}
}
