package typechecker

import "fmt"

type Registry struct {
	entries map[string]*Entry
	parent  *Registry
}

func New() *Registry {
	return &Registry{
		entries: make(map[string]*Entry),
		parent:  nil,
	}
}

func NewEnclosedRegistry(parent *Registry) *Registry {
	registry := New()
	registry.parent = parent

	return registry
}

func (registry *Registry) Lookup(name string) *Entry {
	entry, ok := registry.entries[name]
	if ok {
		return entry
	}

	if registry.parent != nil {
		return registry.parent.Lookup(name)
	}

	return nil
}

func (registry *Registry) LookupLocal(name string) *Entry {
	if entry, exist := registry.entries[name]; exist {
		return entry
	}

	return nil
}

func (registry *Registry) Register(name string, entry *Entry) error {
	if registry.LookupLocal(name) != nil {
		return fmt.Errorf("identifier %s already defined", name)
	}

	registry.entries[name] = entry
	return nil
}
