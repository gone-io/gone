package gone

import (
	"reflect"
	"strings"
)

// circularDepsError creates an error for circular dependencies
// Parameters:
// - circularDeps: list of coffins involved in circular dependencies
// Returns: error object containing information about the circular dependency Goners
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
		} else {
			visited[node] = true
			path = append(path, node)
			for _, dep := range initiatorDepsMap[node] {
				if dfs(dep) {
					return true
				}
			}
			path = path[:len(path)-1]
		}
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

func newDependenceAnalyzer(iKeeper iKeeper, logger Logger) *dependenceAnalyzer {
	return &dependenceAnalyzer{
		iKeeper: iKeeper,
		logger:  logger,
	}
}

type dependenceAnalyzer struct {
	Flag
	iKeeper
	logger Logger `gone:"*"`
}

func (s *dependenceAnalyzer) collectDeps() (map[dependency][]dependency, error) {
	depsMap := make(map[dependency][]dependency)
	for _, co := range s.iKeeper.getAllCoffins() {
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
	if s.logger.GetLevel() <= DebugLevel {
		for d, deps := range depsMap {
			s.logger.Debugf("Found %d dependencies for %s:\n%s\n\n", len(deps), d, deps)
		}
	}
	return depsMap, nil
}

func (s *dependenceAnalyzer) getGonerDeps(co *coffin) (fillDependencies, initDependencies []dependency, err error) {
	fillDependencies, err = s.getGonerFillDeps(co)
	if !co.lazyFill {
		initDependencies = append(initDependencies, dependency{
			coffin: co,
			action: fillAction,
		})
	}
	return
}

func (s *dependenceAnalyzer) getGonerFillDeps(co *coffin) (fillDependencies []dependency, err error) {
	of := reflect.TypeOf(co.goner)
	if of.Kind() != reflect.Ptr {
		return nil, NewInnerError("goner must be a pointer", GonerTypeNotMatch)
	}

	elem := of.Elem()
	switch elem.Kind() {
	case reflect.Struct:
		for i := 0; i < elem.NumField(); i++ {
			field := elem.Field(i)

			if isLazyField(&field) {
				continue
			}

			if err = s.analyzerFieldDependencies(
				field,
				co.Name(),
				func(asSlice, byName bool, extend string, depCoffins ...*coffin) error {
					for _, depCo := range depCoffins {
						if depCo.needInitBeforeUse {
							fillDependencies = append(fillDependencies, dependency{
								coffin: depCo,
								action: initAction,
							})
						}
					}
					return nil
				},
			); err != nil {
				return nil, err
			}
		}
		return RemoveRepeat(fillDependencies), nil
	default:
		return nil, nil
	}
}

func (s *dependenceAnalyzer) analyzerFieldDependencies(
	field reflect.StructField,
	coName string,
	process func(asSlice, byName bool, extend string, coffins ...*coffin) error,
) error {
	var tag string
	var suc bool
	if tag, suc = field.Tag.Lookup(goneTag); !suc {
		return nil
	}
	gonerName, extend := ParseGoneTag(tag)
	if gonerName == "" {
		gonerName = "*"
	}

	isAllowNil := isAllowNilField(&field)

	var depCo *coffin
	var byName bool
	if strings.Contains(gonerName, "*") || strings.Contains(gonerName, "?") {
		depCo = s.selectOneCoffin(field.Type, gonerName, func() {
			s.logger.Warnf("found multiple value without a default when filling filed %q of %q - using first one.", field.Name, coName)
		})
	} else {
		depCo = s.iKeeper.getByName(gonerName)
		byName = depCo != nil
	}

	if depCo != nil {
		return process(false, byName, extend, depCo)
	} else if field.Type.Kind() == reflect.Slice {
		isAllowNil = true
		elType := field.Type.Elem()
		depCos := s.iKeeper.getByTypeAndPattern(elType, gonerName)
		if len(depCos) > 0 {
			return process(true, byName, extend, depCos...)
		}
	}

	if !isAllowNil {
		return NewInnerErrorWithParams(GonerTypeNotMatch,
			"no compatible value found for field %q of %q",
			field.Name, coName,
		)
	}
	return nil
}

func (s *dependenceAnalyzer) checkCircularDepsAndGetBestInitOrder() (circularDeps []dependency, initOrder []dependency, err error) {
	var deps map[dependency][]dependency
	if deps, err = s.collectDeps(); err != nil {
		return
	}
	circularDeps, initOrder = checkCircularDepsAndGetBestInitOrder(deps)
	return
}
