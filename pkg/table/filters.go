package table

// SelectFilter represents a select type filter
type SelectFilter struct {
	Key      string
	LabelStr string
	Options  []FilterOption
}

// Select creates a new select filter
func Select(key string) *SelectFilter {
	return &SelectFilter{
		Key:      key,
		LabelStr: key,
		Options:  make([]FilterOption, 0),
	}
}

// Label sets the filter label
func (f *SelectFilter) Label(label string) *SelectFilter {
	f.LabelStr = label
	return f
}

// WithOptions sets the filter options
func (f *SelectFilter) WithOptions(options []FilterOption) *SelectFilter {
	f.Options = options
	return f
}

// Implementation of the Filter interface
func (f *SelectFilter) GetKey() string             { return f.Key }
func (f *SelectFilter) GetLabel() string           { return f.LabelStr }
func (f *SelectFilter) GetType() string            { return "select" }
func (f *SelectFilter) GetOptions() []FilterOption { return f.Options }

// BooleanFilter represents a boolean filter
type BooleanFilter struct {
	Key      string
	LabelStr string
}

// Boolean creates a new boolean filter
func Boolean(key string) *BooleanFilter {
	return &BooleanFilter{
		Key:      key,
		LabelStr: key,
	}
}

// Label sets the filter label
func (f *BooleanFilter) Label(label string) *BooleanFilter {
	f.LabelStr = label
	return f
}

// Implementation of the Filter interface
func (f *BooleanFilter) GetKey() string   { return f.Key }
func (f *BooleanFilter) GetLabel() string { return f.LabelStr }
func (f *BooleanFilter) GetType() string  { return "boolean" }
func (f *BooleanFilter) GetOptions() []FilterOption {
	return []FilterOption{
		{Value: "true", Label: "Yes"},
		{Value: "false", Label: "No"},
	}
}
