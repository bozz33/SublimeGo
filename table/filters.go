package table

// SelectFilter represents a select type filter
type SelectFilter struct {
	filterKey string
	LabelStr  string
	Options   []FilterOption
}

// Select creates a new select filter
func Select(key string) *SelectFilter {
	return &SelectFilter{
		filterKey: key,
		LabelStr:  key,
		Options:   make([]FilterOption, 0),
	}
}

// WithLabel sets the filter label
func (f *SelectFilter) WithLabel(label string) *SelectFilter {
	f.LabelStr = label
	return f
}

// WithOptions sets the filter options
func (f *SelectFilter) WithOptions(options []FilterOption) *SelectFilter {
	f.Options = options
	return f
}

// Implementation of the Filter interface
func (f *SelectFilter) Key() string                   { return f.filterKey }
func (f *SelectFilter) Label() string                 { return f.LabelStr }
func (f *SelectFilter) Type() string                  { return "select" }
func (f *SelectFilter) FilterOptions() []FilterOption { return f.Options }

// BooleanFilter represents a boolean filter
type BooleanFilter struct {
	filterKey string
	LabelStr  string
}

// Boolean creates a new boolean filter
func Boolean(key string) *BooleanFilter {
	return &BooleanFilter{
		filterKey: key,
		LabelStr:  key,
	}
}

// WithLabel sets the filter label
func (f *BooleanFilter) WithLabel(label string) *BooleanFilter {
	f.LabelStr = label
	return f
}

// Implementation of the Filter interface
func (f *BooleanFilter) Key() string   { return f.filterKey }
func (f *BooleanFilter) Label() string { return f.LabelStr }
func (f *BooleanFilter) Type() string  { return "boolean" }
func (f *BooleanFilter) FilterOptions() []FilterOption {
	return []FilterOption{
		{Value: "true", Label: "Yes"},
		{Value: "false", Label: "No"},
	}
}
