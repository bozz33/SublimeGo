package layouts

import "context"

type navGroupsKey struct{}

// WithNavGroups returns a context carrying the given nav groups.
// Use this in multi-panel setups to inject per-panel navigation.
func WithNavGroups(ctx context.Context, groups []NavGroup) context.Context {
	return context.WithValue(ctx, navGroupsKey{}, groups)
}

// GetNavGroups returns nav groups from context, falling back to the global slice.
func GetNavGroups(ctx context.Context) []NavGroup {
	if groups, ok := ctx.Value(navGroupsKey{}).([]NavGroup); ok {
		return groups
	}
	return navGroups
}
