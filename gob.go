package psqr

import (
	"bytes"
	"encoding/gob"
)

var (
	_ gob.GobEncoder = &Quantile{}
	_ gob.GobDecoder = &Quantile{}
)

// internalQuantile mimics Quantile while exporting its fields and being itself not exported.
// This enables easy marshalling without exposing internal structure to consumers.
type internalQuantile struct {
	P       float64
	Filled  bool
	Pos     [nMarkers]int
	NPos    [nMarkers]float64
	DN      [nMarkers]float64
	Heights []float64
}

func (q *Quantile) GobEncode() ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(internalQuantile{
		P:       q.p,
		Filled:  q.filled,
		Pos:     q.pos,
		NPos:    q.npos,
		DN:      q.dn,
		Heights: q.heights,
	}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (q *Quantile) GobDecode(data []byte) error {
	var interim internalQuantile
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&interim); err != nil {
		return err
	}

	q.p = interim.P
	q.filled = interim.Filled
	q.pos = interim.Pos
	q.npos = interim.NPos
	q.dn = interim.DN
	q.heights = interim.Heights

	return nil
}
