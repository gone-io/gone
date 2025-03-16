package gone

import (
	"fmt"
	"reflect"
	"strings"
)

// circularDepsError creates an error for circular dependencies
// Parameters:
// - circularDeps: list of coffins involved in circular dependencies
// Returns: error object containing information about the circular dependency components
func circularDepsError(circularDeps []dependency) Error {
	var names []string
	prefix := "\t"

	for _, dep := range circularDeps {
		prefix = strings.Join([]string{prefix, "\t"}, "")
		names = append(names, strings.Join([]string{prefix, dep.String()}, ""))
	}
	return NewInnerErrorWithParams(CircularDependency, "circular dependency:\n%s", strings.Join(names, " depend on\n"))
}

// checkCircularDepsAndGetBestInitOrder checks for circular dependencies and determines the optimal initialization order
// Parameters:
// - initiatorDepsMap: dependency map for initiators
// Returns:
// - circularDeps: list of T involved in circular dependencies (forms a loop with first and last elements the same)
// - initOrder: optimal initialization order
func checkCircularDepsAndGetBestInitOrder[T comparable](initiatorDepsMap map[T][]T) (circularDeps []T, initOrder []T) {
	// Record in-degree (number of dependencies) for each node
	inDegree := make(map[T]int)
	// Reverse map for tracing back circular dependencies
	reverseDeps := make(map[T][]T)

	// Initialize in-degree and reverse dependency map
	for node, deps := range initiatorDepsMap {
		if _, exists := inDegree[node]; !exists {
			inDegree[node] = 0
		}
		for _, dep := range deps {
			if _, exists := inDegree[dep]; !exists {
				inDegree[dep] = 0
			}
			inDegree[node]++
			reverseDeps[dep] = append(reverseDeps[dep], node)
		}
	}

	// Find nodes with in-degree of 0 as starting points
	var queue []T
	for node := range inDegree {
		if inDegree[node] == 0 {
			queue = append(queue, node)
		}
	}

	// Perform topological sort
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		initOrder = append(initOrder, node)

		for _, dependent := range reverseDeps[node] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// If there are remaining nodes with non-zero in-degree, detect circular dependencies
	visited := make(map[T]bool)
	var path []T
	var dfs func(T) bool

	dfs = func(node T) bool {
		if visited[node] {
			for i := len(path) - 1; i >= 0; i-- {
				if path[i] == node {
					circularDeps = append(circularDeps, path[i:]...)
					circularDeps = append(circularDeps, node) // Close the loop
					return true
				}
			}
			return false
		}

		visited[node] = true
		path = append(path, node)
		for _, dep := range initiatorDepsMap[node] {
			if dfs(dep) {
				return true
			}
		}
		path = path[:len(path)-1]
		return false
	}

	for node, degree := range inDegree {
		if degree > 0 && !visited[node] {
			if dfs(node) {
				break
			}
		}
	}

	return
}

func (s *Core) getDepByName(name string) (*coffin, error) {
	co := s.nameMap[name]
	if co != nil {
		return co, nil
	}
	return nil, NewInnerErrorWithParams(GonerNameNotFound, "Goner(name=%s) not found", name)
}

func (s *Core) getDepByType(t reflect.Type) (*coffin, error) {
	co := s.getDefaultCoffinByType(t)
	if co != nil {
		return co, nil
	}

	co = s.typeProviderDepMap[t]
	if co != nil {
		return co, nil
	}

	extend := ""
	if t.Kind() == reflect.Struct {
		extend = "; Maybe, you should use A Pointer to this type?"
	}

	return nil, NewInnerErrorWithParams(GonerTypeNotFound, "Type(type=%s) not found%s", GetTypeName(t), extend)
}

func (s *Core) getSliceDepsByType(t reflect.Type) (deps []*coffin) {
	if t.Kind() != reflect.Slice {
		return nil
	}

	co := s.getDefaultCoffinByType(t)
	if co != nil {
		return []*coffin{co}
	}

	co = s.typeProviderDepMap[t]
	if co != nil {
		return []*coffin{co}
	}

	coffins := s.getCoffinsByType(t.Elem())
	deps = append(deps, coffins...)
	co = s.typeProviderDepMap[t.Elem()]
	if co != nil {
		deps = append(deps, co)
	}
	return deps
}

func (s *Core) collectDeps() (map[dependency][]dependency, error) {
	depsMap := make(map[dependency][]dependency)
	for _, co := range s.coffins {
		fillDependency, initDependency, err := s.getGonerDeps(co)
		if err != nil {
			return nil, ToError(err)
		}
		if len(fillDependency) > 0 {
			depsMap[dependency{co, fillAction}] = fillDependency
		}
		if len(initDependency) > 0 {
			depsMap[dependency{co, initAction}] = initDependency
		}
	}
	if s.log.GetLevel() <= DebugLevel {
		for d, deps := range depsMap {
			s.log.Debugf("Found %d dependencies for %s:\n%s\n\n", len(deps), d, deps)
		}
	}
	return depsMap, nil
}

func (s *Core) getGonerDeps(co *coffin) (fillDependencies, initDependencies []dependency, err error) {
	fillDependencies, err = s.getGonerFillDeps(co)
	if !co.lazyFill {
		initDependencies = append(initDependencies, dependency{
			coffin: co,
			action: fillAction,
		})
	}
	return
}

func (s *Core) getGonerFillDeps(co *coffin) (fillDependencies []dependency, err error) {
	of := reflect.TypeOf(co.goner)
	if of.Kind() != reflect.Ptr {
		return nil, NewInnerError("goner must be a pointer", GonerTypeNotMatch)
	}

	elem := of.Elem()
	switch elem.Kind() {
	case reflect.Struct:
		for i := 0; i < elem.NumField(); i++ {
			field := elem.Field(i)
			if tag, ok := field.Tag.Lookup(goneTag); ok {
				gonerName, _ := ParseGoneTag(tag)
				if gonerName == "" || gonerName == "*" {
					gonerName = DefaultProviderName
				}
				if gonerName != DefaultProviderName {
					depCo, err := s.getDepByName(gonerName)
					if err != nil {
						return nil, ToErrorWithMsg(err, fmt.Sprintf("Cannot find matched value for field %q of %q", field.Name, GetTypeName(elem)))
					}

					if depCo.needInitBeforeUse {
						fillDependencies = append(fillDependencies, dependency{
							coffin: depCo,
							action: initAction,
						})
					}
				} else {
					if field.Type.Kind() == reflect.Slice {
						sliceDeps := s.getSliceDepsByType(field.Type)
						for _, depCo := range sliceDeps {
							if depCo.needInitBeforeUse {
								fillDependencies = append(fillDependencies, dependency{
									coffin: depCo,
									action: initAction,
								})
							}
						}
					} else {
						depCo, err := s.getDepByType(field.Type)
						if err != nil {
							if t, ok := field.Tag.Lookup(optionTag); ok && t == allowNil {
								continue
							}
							return nil, ToErrorWithMsg(err, fmt.Sprintf("Cannot find matched value for field %q of %q", field.Name, GetTypeName(elem)))
						}

						if depCo != nil {
							if depCo.needInitBeforeUse {
								fillDependencies = append(fillDependencies, dependency{
									coffin: depCo,
									action: initAction,
								})
							}
						}
					}
				}
			}
		}
		return RemoveRepeat(fillDependencies), nil
	default:
		return nil, nil
	}
}
