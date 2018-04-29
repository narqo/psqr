package psqr

import (
	"math/rand"
	"sort"
	"testing"
)

func TestQuantile(t *testing.T) {
	rand.Seed(100)

	pp := []float64{0.01, 0.1, 0.5, 0.9, 0.99, 0.999}

	eps := 0.01 // permissible error in quantile estimation comparing to sample based calculations
	dataSize := int(1e5 + 100)

	cases := map[string]func() float64{
		"rand": func() float64 { return rand.Float64() * 100 },
		"norm": rand.NormFloat64,
		"exp":  rand.ExpFloat64,
	}

	for name, nextv := range cases {
		t.Run(name, func(t *testing.T) {
			testdata := generateTestData(dataSize, nextv)
			for _, p := range pp {
				testPQuantile(t, p, eps, testdata)
			}
		})
	}
}

func testPQuantile(t *testing.T, p, eps float64, testdata []float64) {
	q := NewQuantile(p)
	for _, v := range testdata {
		q.Append(v)
	}

	// samples keeps a copy of test data to calculate sample-quantile
	samples := make([]float64, len(testdata))
	copy(samples, testdata)
	sort.Float64s(samples)

	var rank, lower, upper int
	n := float64(len(samples))
	if rank = int(p * n); rank < 1 {
		rank = 1
	}
	if lower = int((p - eps) * n); lower < 1 {
		lower = 1
	}
	if upper = int((p + eps) * n); upper > len(samples) {
		upper = len(samples)
	}

	want, min, max := samples[rank-1], samples[lower-1], samples[upper-1]
	if got := q.Value(); got < min || got > max {
		t.Errorf("for p=%v, got %v, want %v [%f, %f]", p, got, want, min, max)
	}
}

func TestQuantileSmallDataset(t *testing.T) {
	for p, want := range map[float64]float64{
		0.01: 1,
		0.1:  1,
		0.5:  3,
		0.9:  5,
		0.99: 5,
	} {
		q := NewQuantile(p)
		for _, v := range []float64{1, 2, 3, 4, 5} {
			q.Append(v)
		}
		if got := q.Value(); got != want {
			t.Errorf("for p=%v: got %v, want %v", p, got, want)
		}
	}
}

func TestQuantile_Reset(t *testing.T) {
	q := NewQuantile(0.5)

	if got, want := q.Value(), float64(0); got != want {
		t.Errorf("for empty stream, got %v, want %v", got, want)
	}

	for i := float64(0); i < 1e6; i++ {
		q.Append(i)
	}
	if got := q.Value(); got == float64(0) {
		t.Errorf("for filled stream, got %v, want non zere", got)
	}

	q.Reset()

	if got, want := q.Value(), float64(0); got != want {
		t.Errorf("after reset, got %v, want %v", got, want)
	}
}

func generateTestData(size int, nextv func() float64) []float64 {
	data := make([]float64, 0, size)
	for i := 0; i < cap(data); i++ {
		v := nextv()
		// add additional variance to dataset
		if i%20 == 0 {
			v = v*v + 1
		}
		data = append(data, v)
	}
	return data
}

func BenchmarkQuantile_Append(b *testing.B) {
	q := NewQuantile(0.5)
	for i := float64(0); i < float64(b.N); i++ {
		q.Append(i)
	}
}

func BenchmarkQuantile_Value(b *testing.B) {
	q := NewQuantile(0.5)
	for i := float64(0); i < 1e6; i++ {
		q.Append(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Value()
	}
}
