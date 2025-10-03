package instruction

import (
	"compiler/nir/value"
	"fmt"
)

// StoreInstruction stores a value into a memory location
// Example: Store(%x, %value)
type StoreInstruction struct {
	destination value.Value
	value       value.Value
}

func NewStoreInstruction(destination value.Value, value value.Value) *StoreInstruction {
	return &StoreInstruction{
		destination: destination,
		value:       value,
	}
}

func (store *StoreInstruction) GetDestination() value.Value {
	return store.destination
}

func (store *StoreInstruction) GetValue() value.Value {
	return store.value
}

func (store *StoreInstruction) String() string {
	return fmt.Sprintf("Store(%s, %s)", store.destination.String(), store.value.String())
}

func (store *StoreInstruction) GetResult() value.Value {
	return nil
}
