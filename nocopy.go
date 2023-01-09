package httpexpect

// Every matcher struct, e.g. Value, Object, Array, etc. contains a chain instance.

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
