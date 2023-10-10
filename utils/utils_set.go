package utils

import (
	"fmt"
	"sort"
	"strings"
)

type SetKey interface {
	Key() string
}

type Set[T SetKey] interface {
	Add(item T)
	AddList(items ...T)
	AddSet(set Set[T])
	Contains(item T) bool
	ContainsKey(k string) bool
	Find(k SetKey) (T, bool)
	Remove(item T)
	ToList() []T
	ToKeyList() []string
	Size() int
	Clone() Set[T]
	String() string
}

type setImpl[T SetKey] struct {
	s map[string]T
}

func NewSet[T SetKey]() Set[T] {
	return new(setImpl[T])
}

func (s *setImpl[T]) Add(item T) {
	if s.s == nil {
		s.s = make(map[string]T)
	}
	s.s[item.Key()] = item
}

func (s *setImpl[T]) Contains(item T) bool {
	if s.s == nil {
		return false
	}
	return s.ContainsKey(item.Key())
}

func (s *setImpl[T]) ContainsKey(k string) bool {
	if s.s == nil {
		return false
	}

	_, ok := s.s[k]
	return ok
}

func (s *setImpl[T]) Find(k SetKey) (a T, ok bool) {
	if s.s == nil {
		return a, false
	}
	v, ok := s.s[k.Key()]
	return v, ok
}

func (s *setImpl[T]) ToList() []T {
	if s == nil {
		return nil
	}
	var list []T
	for _, v := range s.s {
		list = append(list, v)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Key() < list[j].Key()
	}) // to make the result stable
	return list
}

func (s *setImpl[T]) ToKeyList() []string {
	if s == nil {
		return nil
	}

	keys := make([]string, 0, s.Size())
	for _, v := range s.s {
		keys = append(keys, v.Key())
	}
	sort.Strings(keys)
	return keys
}

func (s *setImpl[T]) AddList(items ...T) {
	for _, item := range items {
		s.Add(item)
	}
}

func (s *setImpl[T]) AddSet(set Set[T]) {
	s.AddList(set.ToList()...)
}

func (s *setImpl[T]) Remove(item T) {
	delete(s.s, item.Key())
}

func (s *setImpl[T]) Size() int {
	if s == nil {
		return 0
	}
	return len(s.s)
}

func (s *setImpl[T]) Clone() Set[T] {
	clone := NewSet[T]()
	clone.AddSet(s)
	return clone
}

func (s *setImpl[T]) String() string {
	var items []string
	for _, item := range s.s {
		items = append(items, item.Key())
	}
	sort.Strings(items)
	return fmt.Sprintf("{%v}", strings.Join(items, ", "))
}

func ListToSet[T SetKey](items ...T) Set[T] {
	s := NewSet[T]()
	s.AddList(items...)
	return s
}

func UnionSet[T SetKey](ss ...Set[T]) Set[T] {
	if len(ss) == 0 {
		return NewSet[T]()
	}
	if len(ss) == 1 {
		return ss[0].Clone()
	}
	s := NewSet[T]()
	for _, set := range ss {
		s.AddSet(set)
	}
	return s
}

func AndSet[T SetKey](ss ...Set[T]) Set[T] {
	if len(ss) == 0 {
		return NewSet[T]()
	}
	if len(ss) == 1 {
		return ss[0].Clone()
	}
	s := NewSet[T]()
	for _, item := range ss[0].ToList() {
		contained := true
		for _, set := range ss[1:] {
			if !set.Contains(item) {
				contained = false
				break
			}
		}
		if contained {
			s.Add(item)
		}
	}
	return s
}

// DiffSet returns a set of items that are in s1 but not in s2.
// DiffSet({1, 2, 3, 4}, {2, 3}) = {1, 4}
func DiffSet[T SetKey](s1, s2 Set[T]) Set[T] {
	s := NewSet[T]()
	for _, item := range s1.ToList() {
		if !s2.Contains(item) {
			s.Add(item)
		}
	}
	return s
}

func CombSet[T SetKey](s Set[T], numberOfItems int) []Set[T] {
	return combSetIterate(s.ToList(), NewSet[T](), 0, numberOfItems)
}

func combSetIterate[T SetKey](itemList []T, currSet Set[T], depth, numberOfItems int) []Set[T] {
	if currSet.Size() == numberOfItems {
		return []Set[T]{currSet.Clone()}
	}
	if depth == len(itemList) || currSet.Size() > numberOfItems {
		return nil
	}
	var res []Set[T]
	currSet.Add(itemList[depth])
	res = append(res, combSetIterate(itemList, currSet, depth+1, numberOfItems)...)
	currSet.Remove(itemList[depth])
	res = append(res, combSetIterate(itemList, currSet, depth+1, numberOfItems)...)
	return res
}
