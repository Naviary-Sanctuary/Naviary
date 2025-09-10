package types

// Symbol represents a variable or function in the symbol table
type Symbol struct {
	Name    string
	Type    Type
	Mutable bool
}

type SymbolTable struct {
	symbols map[string]*Symbol
	parent  *SymbolTable // parent scope. nil for global scope
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		symbols: make(map[string]*Symbol),
		parent:  nil,
	}
}

// add a new symbol to the symbol table
// Return false if symbol already defined
func (table *SymbolTable) Define(name string, symbolType Type, mutable bool) bool {
	if table.LookupLocal(name) != nil {
		return false
	}

	table.symbols[name] = &Symbol{
		Name:    name,
		Type:    symbolType,
		Mutable: mutable,
	}

	return true
}

func (table *SymbolTable) Lookup(name string) *Symbol {
	if symbol := table.LookupLocal(name); symbol != nil {
		return symbol
	}

	if table.parent != nil {
		return table.parent.Lookup(name)
	}

	return nil
}

func (table *SymbolTable) LookupLocal(name string) *Symbol {
	if symbol, exist := table.symbols[name]; exist {
		return symbol
	}

	return nil
}

func (table *SymbolTable) Parent() *SymbolTable {
	return table.parent
}

// NewChildScope creates a new symbol table with current table as parent
func (table *SymbolTable) NewChildScope() *SymbolTable {
	return &SymbolTable{
		symbols: make(map[string]*Symbol),
		parent:  table,
	}
}
