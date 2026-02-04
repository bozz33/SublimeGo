package layouts

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

// Page creates a complete page with the Base layout and content
func Page(title string, content templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return Base(title).Render(templ.WithChildren(ctx, content), w)
	})
}
