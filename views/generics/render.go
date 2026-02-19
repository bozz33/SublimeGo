package generics

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/bozz33/sublimego/form"
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
		case *form.TextareaInput:
			return Textarea(v).Render(ctx, w)
		case *form.SelectInput:
			return SelectField(v).Render(ctx, w)
		case *form.CheckboxInput:
			return CheckboxField(v).Render(ctx, w)
		case *form.FileUploadInput:
			return FileUploadField(v).Render(ctx, w)
		case *form.DatePicker:
			return DatePickerField(v).Render(ctx, w)
		case *form.HiddenField:
			return HiddenInputField(v).Render(ctx, w)
		case *form.ToggleInput:
			return ToggleField(v).Render(ctx, w)
		case *form.RepeaterField:
			return RepeaterFieldView(v).Render(ctx, w)

		default:
			return nil
		}
	})
}
