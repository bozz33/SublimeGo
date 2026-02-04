package generics

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/bozz33/SublimeGo/pkg/form"
)

// RenderComponent is the smart switch that decides which template to call
func RenderComponent(c form.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		switch v := c.(type) {
		// Layouts
		case *form.Section:
			return Section(v).Render(ctx, w)
		case *form.Grid:
			return Grid(v).Render(ctx, w)
		case *form.Tabs:
			return Tabs(v).Render(ctx, w)

		// Fields
		case *form.TextInput:
			return TextInput(v).Render(ctx, w)
		case *form.Textarea:
			return Textarea(v).Render(ctx, w)
		case *form.Select:
			return SelectField(v).Render(ctx, w)
		case *form.Checkbox:
			return CheckboxField(v).Render(ctx, w)
		case *form.FileUpload:
			return FileUploadField(v).Render(ctx, w)

		default:
			return nil
		}
	})
}
