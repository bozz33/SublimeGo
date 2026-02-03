package table

// SelectFilter represente un filtre de type select
type SelectFilter struct {
	Key      string
	LabelStr string
	Options  []FilterOption
}

// Select cree un nouveau filtre select
func Select(key string) *SelectFilter {
	return &SelectFilter{
		Key:      key,
		LabelStr: key,
		Options:  make([]FilterOption, 0),
	}
}

// Label definit le label du filtre
func (f *SelectFilter) Label(label string) *SelectFilter {
	f.LabelStr = label
	return f
}

// WithOptions definit les options du filtre
func (f *SelectFilter) WithOptions(options []FilterOption) *SelectFilter {
	f.Options = options
	return f
}

// Implementation de l'interface Filter
func (f *SelectFilter) GetKey() string            { return f.Key }
func (f *SelectFilter) GetLabel() string          { return f.LabelStr }
func (f *SelectFilter) GetType() string           { return "select" }
func (f *SelectFilter) GetOptions() []FilterOption { return f.Options }

// BooleanFilter represente un filtre booleen
type BooleanFilter struct {
	Key      string
	LabelStr string
}

// Boolean cree un nouveau filtre booleen
func Boolean(key string) *BooleanFilter {
	return &BooleanFilter{
		Key:      key,
		LabelStr: key,
	}
}

// Label definit le label du filtre
func (f *BooleanFilter) Label(label string) *BooleanFilter {
	f.LabelStr = label
	return f
}

// Implementation de l'interface Filter
func (f *BooleanFilter) GetKey() string   { return f.Key }
func (f *BooleanFilter) GetLabel() string { return f.LabelStr }
func (f *BooleanFilter) GetType() string  { return "boolean" }
func (f *BooleanFilter) GetOptions() []FilterOption {
	return []FilterOption{
		{Value: "true", Label: "Oui"},
		{Value: "false", Label: "Non"},
	}
}
