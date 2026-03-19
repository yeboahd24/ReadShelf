package claude

import (
	"context"
	"fmt"
	"strings"

	"github.com/dominic/readshelf/internal/core/domain"
	"github.com/dominic/readshelf/internal/core/port/outbound"
)

// aiClient is a placeholder that returns a stub response.
// Replace with actual Claude API calls when ready.
type aiClient struct{}

func NewAIClient(_ string) outbound.AIClient {
	return &aiClient{}
}

func (c *aiClient) Recall(_ context.Context, query string, passages []domain.Annotation) (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Based on your annotations, here is what I found about: %q\n\n", query))
	for i, p := range passages {
		sb.WriteString(fmt.Sprintf("%d. [%s, p.%d] %s\n", i+1, p.BookTitle, p.Page, p.Content))
	}
	return sb.String(), nil
}
