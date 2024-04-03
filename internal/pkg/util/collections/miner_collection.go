// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Modified copy of k8s.io/apimachinery/pkg/util/sets/int64.go
// Modifications
//   - int64 became *v1beta1.Miner
//   - Empty type is removed
//   - Sortable data type is removed in favor of util.MinersByCreationTimestamp
//   - nil checks added to account for the pointer
//   - Added Filter, AnyFilter, and Oldest methods
//   - Added FromMinerList initializer
//   - Updated Has to also check for equality of Miners
//   - Removed unused methods

package collections

import (
	"sort"

	"github.com/superproj/onex/internal/pkg/util/conditions"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

// Miners is a set of Miners.
type Miners map[string]*v1beta1.Miner

// MinersByVersion sorts the list of Miner by spec.version, using their names as tie breaker.
// miners with no version are placed lower in the order.
type minersByVersion []*v1beta1.Miner

func (v minersByVersion) Len() int      { return len(v) }
func (v minersByVersion) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v minersByVersion) Less(i, j int) bool {
	/* TODO
	vi, _ := semver.ParseTolerant(*v[i].Spec.Version)
	vj, _ := semver.ParseTolerant(*v[j].Spec.Version)
	comp := version.Compare(vi, vj, version.WithBuildTags())
	if comp == 0 {
		return v[i].Name < v[j].Name
	}
	return comp == -1
	*/
	return true
}

// minersByCreationTimestamp sorts a list of Miner by creation timestamp, using their names as a tie breaker.
type minersByCreationTimestamp []*v1beta1.Miner

func (o minersByCreationTimestamp) Len() int      { return len(o) }
func (o minersByCreationTimestamp) Swap(i, j int) { o[i], o[j] = o[j], o[i] }
func (o minersByCreationTimestamp) Less(i, j int) bool {
	if o[i].CreationTimestamp.Equal(&o[j].CreationTimestamp) {
		return o[i].Name < o[j].Name
	}
	return o[i].CreationTimestamp.Before(&o[j].CreationTimestamp)
}

// New creates an empty Miners.
func New() Miners {
	return make(Miners)
}

// FromMiners creates a Miners from a list of values.
func FromMiners(miners ...*v1beta1.Miner) Miners {
	ss := make(Miners, len(miners))
	ss.Insert(miners...)
	return ss
}

// FromMinerList creates a Miners from the given MinerList.
func FromMinerList(minerList *v1beta1.MinerList) Miners {
	ss := make(Miners, len(minerList.Items))
	for i := range minerList.Items {
		ss.Insert(&minerList.Items[i])
	}
	return ss
}

// ToMinerList creates a MinerList from the given Miners.
func ToMinerList(miners Miners) v1beta1.MinerList {
	ml := v1beta1.MinerList{}
	for _, m := range miners {
		ml.Items = append(ml.Items, *m)
	}
	return ml
}

// Insert adds items to the set.
func (s Miners) Insert(miners ...*v1beta1.Miner) {
	for i := range miners {
		if miners[i] != nil {
			m := miners[i]
			s[m.Name] = m
		}
	}
}

// Difference returns a copy without miners that are in the given collection.
func (s Miners) Difference(miners Miners) Miners {
	return s.Filter(func(m *v1beta1.Miner) bool {
		_, found := miners[m.Name]
		return !found
	})
}

// SortedByCreationTimestamp returns the miners sorted by creation timestamp.
func (s Miners) SortedByCreationTimestamp() []*v1beta1.Miner {
	res := make(minersByCreationTimestamp, 0, len(s))
	for _, value := range s {
		res = append(res, value)
	}
	sort.Sort(res)
	return res
}

// UnsortedList returns the slice with contents in random order.
func (s Miners) UnsortedList() []*v1beta1.Miner {
	res := make([]*v1beta1.Miner, 0, len(s))
	for _, value := range s {
		res = append(res, value)
	}
	return res
}

// Len returns the size of the set.
func (s Miners) Len() int {
	return len(s)
}

// newFilteredMinerCollection creates a Miners from a filtered list of values.
func newFilteredMinerCollection(filter Func, miners ...*v1beta1.Miner) Miners {
	ss := make(Miners, len(miners))
	for i := range miners {
		m := miners[i]
		if filter(m) {
			ss.Insert(m)
		}
	}
	return ss
}

// Filter returns a Miners containing only the Miners that match all of the given MinerFilters.
func (s Miners) Filter(filters ...Func) Miners {
	return newFilteredMinerCollection(And(filters...), s.UnsortedList()...)
}

// AnyFilter returns a Miners containing only the Miners that match any of the given MinerFilters.
func (s Miners) AnyFilter(filters ...Func) Miners {
	return newFilteredMinerCollection(Or(filters...), s.UnsortedList()...)
}

// Oldest returns the Miner with the oldest CreationTimestamp.
func (s Miners) Oldest() *v1beta1.Miner {
	if len(s) == 0 {
		return nil
	}
	return s.SortedByCreationTimestamp()[0]
}

// Newest returns the Miner with the most recent CreationTimestamp.
func (s Miners) Newest() *v1beta1.Miner {
	if len(s) == 0 {
		return nil
	}
	return s.SortedByCreationTimestamp()[len(s)-1]
}

// DeepCopy returns a deep copy.
func (s Miners) DeepCopy() Miners {
	result := make(Miners, len(s))
	for _, m := range s {
		result.Insert(m.DeepCopy())
	}
	return result
}

// ConditionGetters returns the slice with miners converted into conditions.Getter.
func (s Miners) ConditionGetters() []conditions.Getter {
	res := make([]conditions.Getter, 0, len(s))
	for _, v := range s {
		value := *v
		res = append(res, &value)
	}
	return res
}

// Names returns a slice of the names of each miner in the collection.
// Useful for logging and test assertions.
func (s Miners) Names() []string {
	names := make([]string, 0, s.Len())
	for _, m := range s {
		names = append(names, m.Name)
	}
	return names
}

// SortedByVersion returns the miners sorted by version.
func (s Miners) sortedByVersion() []*v1beta1.Miner {
	res := make(minersByVersion, 0, len(s))
	for _, value := range s {
		res = append(res, value)
	}
	sort.Sort(res)
	return res
}

// LowestVersion returns the lowest version among all the miner with
// defined versions. If no miner has a defined version it returns an
// empty string.
func (s Miners) LowestVersion() *string {
	miners := s.Filter(WithVersion())
	if len(miners) == 0 {
		return nil
	}
	m := miners.sortedByVersion()[0]
	return &m.Spec.MinerType
}
