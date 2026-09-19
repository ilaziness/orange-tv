package payment

import (
	"fmt"
	"strings"
	"sync"
)

var (
	regMu sync.RWMutex
	impls = map[string]Payment{}
)

func normalizeProvider(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}

// Register binds a vendor implementation. Call from vendor init() only.
func Register(provider string, impl Payment) {
	if impl == nil {
		panic("payment: Register nil implementation")
	}
	p := normalizeProvider(provider)
	if p == "" {
		panic("payment: Register empty provider")
	}
	regMu.Lock()
	defer regMu.Unlock()
	impls[p] = impl
}

func lookup(provider string) (Payment, error) {
	p := normalizeProvider(provider)
	regMu.RLock()
	defer regMu.RUnlock()
	impl := impls[p]
	if impl == nil {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotRegistered, p)
	}
	return impl, nil
}
