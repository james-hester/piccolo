package internal

import (
	"fmt"
)

// ConfigOp represents a configuration word setting.
type ConfigOp struct {
	Address int
	Value   int
}

func (op ConfigOp) Assembly() string {
	return fmt.Sprintf(" __CONFIG 0x%X, 0x%X", op.Address, op.Value)
}

func (op ConfigOp) Encode(ctx *AssemblerContext) error {
	ctx.Config[op.Address] = uint16(op.Value)
	return nil
}
