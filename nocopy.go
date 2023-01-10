package httpexpect

// noCopy struct is a special type that is used to prevent a value from being copied.
// so `go vet` gives a warning if this struct is copied.
// It has two methods, `Lock` and `Unlock`, that do not do anything.
// The purpose of the `noCopy` struct is to prevent a value that has it as a field from being copied.
// Every matcher struct, e.g. Value, Object, Array, etc. contains a chain instance.

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
