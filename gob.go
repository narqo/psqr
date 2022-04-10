package psqr

import (
	"bytes"
	"encoding/gob"
)

var (
	_ gob.GobEncoder = &Quantile{}
	_ gob.GobDecoder = &Quantile{}
)

func (q *Quantile) GobEncode() ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(q.data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (q *Quantile) GobDecode(in []byte) error {
	var interim data
	if err := gob.NewDecoder(bytes.NewReader(in)).Decode(&interim); err != nil {
		return err
	}

	q.data = &interim

	return nil
}
