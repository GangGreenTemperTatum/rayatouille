package config

import "charm.land/bubbles/v2/key"

// OverrideBinding updates a key.Binding with custom keys if overrides are provided.
// The help key label is updated to show the first custom key.
func OverrideBinding(b *key.Binding, keys []string) {
	if len(keys) == 0 {
		return
	}
	h := b.Help()
	b.SetKeys(keys...)
	b.SetHelp(keys[0], h.Desc)
}
