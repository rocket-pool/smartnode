package state

import "errors"

// ErrStaticMode is returned by static EC/BC backends when a method is invoked
// that cannot be answered from a pre-loaded NetworkState snapshot. Handlers
// that have not yet been ported to consume NetworkStateProvider directly will
// surface this error instead of attempting a live EC/CC call.
var ErrStaticMode = errors.New("operation not supported in static network state mode")
