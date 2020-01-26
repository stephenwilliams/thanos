// Copyright (c) The Thanos Authors.
// Licensed under the Apache License 2.0.

package testutil

import (
	"testing"
)

// TB represents union of test and benchmark.
// This allows the same test suite to be run by both benchmark and test, helping to reuse more code.
// The reason is that usually benchmarks are not being run on CI, especially for short tests, so you need to recreate
// usually similar tests for `Test<Name>(t *testing.T)` methods. Example of usage is presented here:
//
//	 func TestTestOrBench(t *testing.T) {
//		tb := NewTB(t)
//		tb.Run("1", func(tb TB) { testorbenchComplexTest(tb) })
//		tb.Run("2", func(tb TB) { testorbenchComplexTest(tb) })
//	}
//
//	func BenchmarkTestOrBench(b *testing.B) {
//		tb := NewTB(t)
//		tb.Run("1", func(tb TB) { testorbenchComplexTest(tb) })
//		tb.Run("2", func(tb TB) { testorbenchComplexTest(tb) })
//	}
type TB interface {
	testing.TB
	IsBenchmark() bool
	Run(name string, f func(t TB)) bool
	N() int
	ResetTimer()
}

// UnionTB implements TB as well as testing.TB interfaces.
type UnionTB struct {
	testing.TB
}

// NewTB creates UnionTB from testing.TB.
func NewTB(tb testing.TB) TB { return &UnionTB{TB: tb} }

func (tb *UnionTB) Run(name string, f func(t TB)) bool {
	if b, ok := tb.TB.(*testing.B); ok {
		return b.Run(name, func(nested *testing.B) { f(&UnionTB{TB: nested}) })
	}
	if t, ok := tb.TB.(*testing.T); ok {
		return t.Run(name, func(nested *testing.T) { f(&UnionTB{TB: nested}) })
	}
	panic("not a benchmark and not a test")
}

func (tb *UnionTB) N() int {
	if b, ok := tb.TB.(*testing.B); ok {
		return b.N
	}
	return 1
}

func (tb *UnionTB) ResetTimer() {
	if b, ok := tb.TB.(*testing.B); ok {
		b.ResetTimer()
		return
	}
}

func (tb *UnionTB) IsBenchmark() bool {
	if _, ok := tb.TB.(*testing.B); ok {
		return true
	}
	return false
}
