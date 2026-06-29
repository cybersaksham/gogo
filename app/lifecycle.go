package app

import "context"

// Ready resolves dependency order and runs app ready hooks exactly once.
func (r *Registry) Ready(ctx context.Context) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()

	r.mu.Lock()
	if r.ready {
		r.mu.Unlock()
		return nil
	}

	ordered, err := r.resolveDependencyOrder()
	if err != nil {
		r.mu.Unlock()
		return err
	}
	r.preparing = true
	r.mu.Unlock()

	started := make([]Config, 0, len(ordered))
	for _, config := range ordered {
		if err := ctx.Err(); err != nil {
			r.mu.Lock()
			r.ordered = started
			r.preparing = false
			r.mu.Unlock()
			return err
		}
		if err := config.Ready(ctx, r); err != nil {
			r.mu.Lock()
			r.ordered = started
			r.preparing = false
			r.mu.Unlock()
			return err
		}
		started = append(started, config)
	}

	r.mu.Lock()
	r.ordered = ordered
	r.preparing = false
	r.ready = true
	r.mu.Unlock()
	return nil
}

// Shutdown runs app shutdown hooks in reverse dependency order.
func (r *Registry) Shutdown(ctx context.Context) error {
	r.lifecycleMu.Lock()
	defer r.lifecycleMu.Unlock()

	r.mu.Lock()
	if len(r.ordered) == 0 {
		r.mu.Unlock()
		return nil
	}

	ordered := make([]Config, len(r.ordered))
	copy(ordered, r.ordered)
	r.mu.Unlock()

	for i := len(ordered) - 1; i >= 0; i-- {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := ordered[i].Shutdown(ctx); err != nil {
			return err
		}
	}

	r.mu.Lock()
	r.ready = false
	r.ordered = nil
	r.mu.Unlock()
	return nil
}
