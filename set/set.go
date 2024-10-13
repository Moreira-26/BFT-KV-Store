package set

import (
	"cmp"
	"slices"
	"sort"
)

type Set[T cmp.Ordered] []T

func getOrderedKeys[T cmp.Ordered](s map[T]bool) (keys []T) {
	keys = make([]T, 0)
	for k := range s {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return
}

func New[T cmp.Ordered]() Set[T] {
	return make(Set[T], 0)
}

func FromSlice[T cmp.Ordered](slice []T) Set[T] {
	values := make(map[T]bool)
	for _, elem := range slice {
		values[elem] = true
	}

	return getOrderedKeys(values)
}

func Add[T cmp.Ordered](s Set[T], val T) Set[T] {
	i, exists := slices.BinarySearch(s, val)
	if exists {
		return s
	} else {
		return slices.Insert(s, i, val)
	}
}

func Remove[T cmp.Ordered](s Set[T], val T) Set[T] {
	i, exists := slices.BinarySearch(s, val)
	if exists {
		return append(s[:i], s[i+1:]...)
	} else {
		return s
	}
}

func Has[T cmp.Ordered](s Set[T], val T) bool {
	_, exists := slices.BinarySearch(s, val)
	return exists
}

func Union[T cmp.Ordered](s Set[T], o Set[T]) Set[T] {
	for _, v := range o {
		s = Add(s, v)
	}
	return s
}

func Diff[T cmp.Ordered](s Set[T], o Set[T]) Set[T] {
	for _, v := range o {
		s = Remove(s, v)
	}
	return s
}

func Intersect[T cmp.Ordered](s Set[T], o Set[T]) Set[T] {
	values := make(map[T]bool)
	for _, k := range s {
		values[k] = false
	}
	for _, k := range o {
		if _, exists := values[k]; exists {
			values[k] = true
		}
	}

	for k, v := range values {
		if !v {
			delete(values, k)
		}
	}

	return getOrderedKeys(values)
}
