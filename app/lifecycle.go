package app

import "context"

// Ready resolves dependency order and runs app ready hooks exactly once.
func (r *Registry) Ready(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.ready {
		return nil
	}

	ordered, err := r.resolveDependencyOrder()
	if err != nil {
		return err
	}

	started := make([]Config, 0, len(ordered))
	for _, config := range ordered {
		if err := ctx.Err(); err != nil {
			r.ordered = started
			return err
		}
		if err := config.Ready(ctx, r); err != nil {
			r.ordered = started
			return err
		}
		started = append(started, config)
	}

	r.ordered = ordered
	r.ready = true
	return nil
}

// Shutdown runs app shutdown hooks in reverse dependency order.
func (r *Registry) Shutdown(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.ordered) == 0 {
		return nil
	}

	for i := len(r.ordered) - 1; i >= 0; i-- {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := r.ordered[i].Shutdown(ctx); err != nil {
			return err
		}
	}

	r.ready = false
	r.ordered = nil
	return nil
}
