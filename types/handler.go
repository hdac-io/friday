package types

// Handler defines the core of the state transition function of an application.
type Handler func(ctx Context, msg Msg, simulate bool, txIndex int, msgIndex int) Result

// AnteHandler authenticates transactions, before their internal messages are handled.
// If newCtx.IsZero(), ctx is used instead.
type AnteHandler func(ctx Context, tx Tx, simulate bool, txIndex int) (newCtx Context, result Result, abort bool)
