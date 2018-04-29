// Package psqr implements P-Square algorithm for estimating quantiles without storing observations.
package psqr

import (
	"sort"
)

// P-square maitains five markers that store points.
const nMarkers = 5

// Quantile represents an estimated p-quantile of a stream of observations.
type Quantile struct {
	p      float64
	filled bool

	// marker positions, 1..nMarkers
	pos [nMarkers]int
	// desired marker positions
	npos [nMarkers]float64
	// increament in desired marker positions
	dn [nMarkers]float64
	// marker heights that store observations
	heights []float64
}

// NewQuantile returns new p-quantile.
func NewQuantile(p float64) *Quantile {
	if p < 0 || p > 1 {
		panic("p-quantile is out of range")
	}
	q := &Quantile{
		p:       p,
		heights: make([]float64, 0, nMarkers),
	}
	q.Reset()
	return q
}

// Reset resets the quantile.
func (q *Quantile) Reset() {
	p := q.p
	q.filled = false
	q.heights = q.heights[:0]
	for i := 0; i < len(q.pos); i++ {
		q.pos[i] = i
	}
	q.npos = [...]float64{
		0,
		2 * p,
		4 * p,
		2 + 2*p,
		4,
	}
	q.dn = [...]float64{
		0,
		p / 2,
		p,
		(1 + p) / 2,
		1,
	}
}

// Append appends v to the stream of observations.
func (q *Quantile) Append(v float64) {
	if len(q.heights) != nMarkers {
		// no required number of observations has been appended yet
		q.heights = append(q.heights, v)
		return
	}
	if !q.filled {
		q.filled = true
		sort.Float64s(q.heights)
	}
	q.append(v)

}

func (q *Quantile) append(v float64) {
	l := len(q.heights) - 1

	k := -1
	if v < q.heights[0] {
		k = 0
		q.heights[0] = v
	} else if q.heights[l] <= v {
		k = l - 1
		q.heights[l] = v
	} else {
		for i := 1; i <= l; i++ {
			if q.heights[i-1] <= v && v < q.heights[i] {
				k = i - 1
				break
			}
		}
	}

	for i := 0; i < len(q.pos); i++ {
		// increment positions greater than k
		if i > k {
			q.pos[i]++
		}
		// update desired positions for all markers
		q.npos[i] += q.dn[i]
	}

	q.adjustHeights()
}

func (q *Quantile) adjustHeights() {
	for i := 1; i < len(q.heights)-1; i++ {
		n := q.pos[i]
		np1 := q.pos[i+1]
		nm1 := q.pos[i-1]

		d := q.npos[i] - float64(n)

		if (d >= 1 && np1-n > 1) || (d <= -1 && nm1-n < -1) {
			if d >= 0 {
				d = 1
			} else {
				d = -1
			}

			h := q.heights[i]
			hp1 := q.heights[i+1]
			hm1 := q.heights[i-1]

			// try adjusting height using P-square formula
			hi := parabolic(d, hp1, h, hm1, float64(np1), float64(n), float64(nm1))

			if hm1 < hi && hi < hp1 {
				q.heights[i] = hi
			} else {
				// use linear formula
				hd := q.heights[i+int(d)]
				nd := q.pos[i+int(d)]
				q.heights[i] = h + d*(hd-h)/float64(nd-n)
			}

			q.pos[i] += int(d)
		}
	}
}

// Value returns the current estimate of p-quantile.
func (q *Quantile) Value() float64 {
	if !q.filled {
		// a fast path when not enought observations has been stored yet
		l := len(q.heights)
		switch l {
		case 0:
			return 0
		case 1:
			return q.heights[0]
		}
		sort.Float64s(q.heights)
		rank := int(q.p * float64(l))
		return q.heights[rank]
	}
	// if initialised with nMarkers observations third height stores current
	// estimate of p-quantile
	return q.heights[2]
}

// calculates the adjustment of height using  piecewise parabolic (PP) prediction formula.
func parabolic(d, qp1, q, qm1, np1, n, nm1 float64) float64 {
	a := d / (np1 - nm1)
	b1 := (n - nm1 + d) * (qp1 - q) / (np1 - n)
	b2 := (np1 - n - d) * (q - qm1) / (n - nm1)
	return q + a*(b1+b2)
}
