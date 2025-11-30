package internal

// SymbolTable represents a mapping of names to addresses.
type SymbolTable interface {
	GetAddress(name string) (int, bool)
	SetAddress(name string, addr int)
}

// BasicSymbolTable is a simple implementation of SymbolTable.
type BasicSymbolTable struct {
	symbols map[string]int
}

// NewSymbolTable creates a new empty symbol table.
func NewSymbolTable() *BasicSymbolTable {
	return &BasicSymbolTable{
		symbols: make(map[string]int),
	}
}

// GetAddress returns the address of a symbol.
func (st *BasicSymbolTable) GetAddress(name string) (int, bool) {
	addr, ok := st.symbols[name]
	return addr, ok
}

// SetAddress sets the address of a symbol.
func (st *BasicSymbolTable) SetAddress(name string, addr int) {
	st.symbols[name] = addr
}
