package form

import "context"

// Form is the main container.
type Form struct {
	Schema []Component
	Model  any
	State  map[string]any
}

// New creates a new form.
func New() *Form {
	return &Form{
		Schema: make([]Component, 0),
		State:  make(map[string]any),
	}
}

// SetSchema sets the form structure.
func (f *Form) SetSchema(components ...Component) *Form {
	f.Schema = components
	return f
}

// Bind binds a model to the form for pre-filling.
func (f *Form) Bind(model any) *Form {
	f.Model = model
	return f
}

// SaveProcessing handles logic before saving.
func (f *Form) SaveProcessing(ctx context.Context) error {
	return nil
}
