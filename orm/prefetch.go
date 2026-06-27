package orm

// PrefetchSpec configures a batched prefetch query.
type PrefetchSpec struct {
	Path     string
	QuerySet QuerySet
	ToAttr   string
}

// Prefetch creates a prefetch specification.
func Prefetch(path string, queryset QuerySet, toAttr string) PrefetchSpec {
	return PrefetchSpec{Path: path, QuerySet: queryset, ToAttr: toAttr}
}

// PrefetchPlan stores a compiled batched prefetch query.
type PrefetchPlan struct {
	Path    string
	Query   CompiledSQL
	ToAttr  string
	Batched bool
}

// PlanPrefetches compiles separate batched prefetch queries.
func PlanPrefetches(parent QuerySet, specs ...PrefetchSpec) ([]PrefetchPlan, error) {
	_ = parent
	plans := make([]PrefetchPlan, len(specs))
	for i, spec := range specs {
		compiled, err := spec.QuerySet.Iterator()
		if err != nil {
			return nil, err
		}
		plans[i] = PrefetchPlan{
			Path:    spec.Path,
			Query:   compiled,
			ToAttr:  spec.ToAttr,
			Batched: true,
		}
	}
	return plans, nil
}
