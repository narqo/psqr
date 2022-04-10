// Package psqr implements P-Square algorithm for estimating quantiles without storing observations.
package psqr

import (
	"sort"
)

// P-square maitains five markers that store points.
const nMarkers = 5

// Quantile represents an estimated p-quantile of a stream of observations.
type Quantile struct {
	// data is contains the actual information used for quantile calculations. It is unexported to avoid accidental
	// modification, while itself containing exported fields, allowing (un)marshalling.
	data *data
}

type data struct {
	P      float64
	Filled bool

	// marker positions, 1..nMarkers
	Pos [nMarkers]int
	// desired marker positions
	NPos [nMarkers]float64
	// increament in desired marker positions
	DN [nMarkers]float64
	// marker heights that store observations
	Heights []float64
}

// NewQuantile returns new p-quantile.
func NewQuantile(p float64) *Quantile {
	if p < 0 || p > 1 {
		panic("p-quantile is out of range")
	}
	q := &Quantile{
		data: &data{
			P:       p,
			Heights: make([]float64, 0, nMarkers),
		},
	}
	q.Reset()
	return q
}

// Reset resets the quantile.
func (q *Quantile) Reset() {
	p := q.data.P
	q.data.Filled = false
	q.data.Heights = q.data.Heights[:0]
	for i := 0; i < len(q.data.Pos); i++ {
		q.data.Pos[i] = i
	}
	q.data.NPos = [...]float64{
		0,
		2 * p,
		4 * p,
		2 + 2*p,
		4,
	}
	q.data.DN = [...]float64{
		0,
		p / 2,
		p,
		(1 + p) / 2,
		1,
	}
}

// Append appends v to the stream of observations.
func (q *Quantile) Append(v float64) {
	if len(q.data.Heights) != nMarkers {
		// no required number of observations has been appended yet
		q.data.Heights = append(q.data.Heights, v)
		return
	}
	if !q.data.Filled {
		q.data.Filled = true
		sort.Float64s(q.data.Heights)
	}
	q.append(v)
}

func (q *Quantile) append(v float64) {
	l := len(q.data.Heights) - 1

	k := -1
	if v < q.data.Heights[0] {
		k = 0
		q.data.Heights[0] = v
	} else if q.data.Heights[l] <= v {
		k = l - 1
		q.data.Heights[l] = v
	} else {
		for i := 1; i <= l; i++ {
			if q.data.Heights[i-1] <= v && v < q.data.Heights[i] {
				k = i - 1
				break
			}
		}
	}

	for i := 0; i < len(q.data.Pos); i++ {
		// increment positions greater than k
		if i > k {
			q.data.Pos[i]++
		}
		// update desired positions for all markers
		q.data.NPos[i] += q.data.DN[i]
	}

	q.adjustHeights()
}

func (q *Quantile) adjustHeights() {
	for i := 1; i < len(q.data.Heights)-1; i++ {
		n := q.data.Pos[i]
		np1 := q.data.Pos[i+1]
		nm1 := q.data.Pos[i-1]

		d := q.data.NPos[i] - float64(n)

		if (d >= 1 && np1-n > 1) || (d <= -1 && nm1-n < -1) {
			if d >= 0 {
				d = 1
			} else {
				d = -1
			}

			h := q.data.Heights[i]
			hp1 := q.data.Heights[i+1]
			hm1 := q.data.Heights[i-1]

			// try adjusting height using P-square formula
			hi := parabolic(d, hp1, h, hm1, float64(np1), float64(n), float64(nm1))

			if hm1 < hi && hi < hp1 {
				q.data.Heights[i] = hi
			} else {
				// use linear formula
				hd := q.data.Heights[i+int(d)]
				nd := q.data.Pos[i+int(d)]
				q.data.Heights[i] = h + d*(hd-h)/float64(nd-n)
			}

			q.data.Pos[i] += int(d)
		}
	}
}

// Value returns the current estimate of p-quantile.
func (q *Quantile) Value() float64 {
	if !q.data.Filled {
		// a fast path when not enought observations has been stored yet
		l := len(q.data.Heights)
		switch l {
		case 0:
			return 0
		case 1:
			return q.data.Heights[0]
		}
		sort.Float64s(q.data.Heights)
		rank := int(q.data.P * float64(l))
		return q.data.Heights[rank]
	}
	// if initialised with nMarkers observations third height stores current
	// estimate of p-quantile
	return q.data.Heights[2]
}

// calculates the adjustment of height using  piecewise parabolic (PP) prediction formula.
func parabolic(d, qp1, q, qm1, np1, n, nm1 float64) float64 {
	a := d / (np1 - nm1)
	b1 := (n - nm1 + d) * (qp1 - q) / (np1 - n)
	b2 := (np1 - n - d) * (q - qm1) / (n - nm1)
	return q + a*(b1+b2)
}
