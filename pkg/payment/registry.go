package payment

import (
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

func snapshot() map[string]Payment {
	regMu.RLock()
	defer regMu.RUnlock()
	out := make(map[string]Payment, len(impls))
	for k, v := range impls {
		out[k] = v
	}
	return out
}
