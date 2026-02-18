package plugin

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// reset clears the global registry between tests.
func reset() {
	mu.Lock()
	plugins = nil
	mu.Unlock()
}

type mockPlugin struct {
	name    string
	bootErr error
	booted  bool
}

func (m *mockPlugin) Name() string { return m.name }
func (m *mockPlugin) Boot() error {
	m.booted = true
	return m.bootErr
}

func TestRegisterAndAll(t *testing.T) {
	reset()
	p := &mockPlugin{name: "test-plugin"}
	Register(p)

	all := All()
	assert.Len(t, all, 1)
	assert.Equal(t, "test-plugin", all[0].Name())
}

func TestRegisterMultiple(t *testing.T) {
	reset()
	Register(&mockPlugin{name: "a"})
	Register(&mockPlugin{name: "b"})
	Register(&mockPlugin{name: "c"})

	assert.Len(t, All(), 3)
}

func TestGet(t *testing.T) {
	reset()
	Register(&mockPlugin{name: "alpha"})
	Register(&mockPlugin{name: "beta"})

	p := Get("alpha")
	assert.NotNil(t, p)
	assert.Equal(t, "alpha", p.Name())

	missing := Get("gamma")
	assert.Nil(t, missing)
}

func TestBoot(t *testing.T) {
	reset()
	p1 := &mockPlugin{name: "p1"}
	p2 := &mockPlugin{name: "p2"}
	Register(p1)
	Register(p2)

	err := Boot()
	assert.NoError(t, err)
	assert.True(t, p1.booted)
	assert.True(t, p2.booted)
}

func TestBootError(t *testing.T) {
	reset()
	Register(&mockPlugin{name: "ok"})
	Register(&mockPlugin{name: "fail", bootErr: errors.New("boot failed")})
	Register(&mockPlugin{name: "skipped"})

	err := Boot()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fail")
}

func TestAllReturnsCopy(t *testing.T) {
	reset()
	Register(&mockPlugin{name: "x"})

	a := All()
	b := All()
	assert.Equal(t, a, b)

	// Mutating the returned slice should not affect the registry
	a[0] = &mockPlugin{name: "mutated"}
	assert.Equal(t, "x", All()[0].Name())
}
