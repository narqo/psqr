package psqr_test

import (
	"testing"

	"github.com/exaring/psqr"
)

func TestQuantileMarshalling(t *testing.T) {
	quant := psqr.NewQuantile(0.5)

	for i := 0; i < 1000; i++ {
		quant.Append(float64(i))
	}

	before := quant.Value()

	marshalled, err := quant.GobEncode()
	if err != nil {
		t.Errorf("unexpected encoding error: %s", err)
	}

	if err := quant.GobDecode(marshalled); err != nil {
		t.Errorf("unexpected decoding error: %s", err)
	}

	after := quant.Value()

	if before != after {
		t.Errorf("encoded/decoded values differ: expected %f != got %f", before, after)
	}
}

func TestQuantile_EncodingDecodingAndChanging(t *testing.T) {
	quant := psqr.NewQuantile(0.5)

	for i := 0; i < 1000; i++ {
		quant.Append(float64(i))
	}

	encoded, err := quant.GobEncode()
	if err != nil {
		t.Errorf("unexpected encoding error: %s", err)
	}

	decoded := &psqr.Quantile{}

	if err := decoded.GobDecode(encoded); err != nil {
		t.Errorf("unexpected decoding error: %s", err)
	}

	for i := 0; i < 1000; i++ {
		quant.Append(float64(i))
		decoded.Append(float64(i))
	}

	origAfter := quant.Value()
	decodedAfter := decoded.Value()

	if origAfter != decodedAfter {
		t.Errorf("encoded/decoded value did not keep internal representation: expected %f != got %f", origAfter, decodedAfter)
	}
}
