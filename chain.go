package httpexpect

type chain struct {
	ctx     *Context
	failbit bool
}

func makeChain(reporterOrCtx Reporter) chain {
	switch v := reporterOrCtx.(type) {
	case *Context:
		return chain{v, false}
	default:
		return chain{&Context{
			AssertionHandler: DefaultAssertionHandler{
				Reporter:  v,
				Formatter: DefaultFormatter{},
			},
		}, false}
	}
}

func (c *chain) failed() bool {
	return c.failbit
}

func (c *chain) fail(failure Failure) {
	if c.failbit {
		return
	}
	c.failbit = true
	c.ctx.AssertionHandler.Failure(c.ctx, failure)
}

func (c *chain) reset() {
	c.failbit = false
}

func (c *chain) assertFailed(r Reporter) {
	if !c.failbit {
		r.Errorf("expected chain is failed, but it's ok")
	}
}

func (c *chain) assertOK(r Reporter) {
	if c.failbit {
		r.Errorf("expected chain is ok, but it's failed")
	}
}
