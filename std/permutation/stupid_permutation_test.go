package permutation

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sort"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/test"
)

type StupidPermutationCircuit struct {
	A []frontend.Variable
	B []*big.Int
}

func (c *StupidPermutationCircuit) Define(api frontend.API) error {
	sorted := Sort(api, c.A)
	if len(sorted) != len(c.A) {
		return fmt.Errorf("lengths mismatch")
	}
	for i := range c.B {
		api.AssertIsEqual(sorted[i], c.B[i])
	}
	return nil
}

func TestStupidSort(t *testing.T) {
	n := 1000
	circuit := StupidPermutationCircuit{
		A: make([]frontend.Variable, n),
	}
	vars := make([]frontend.Variable, n)
	sorted := make([]*big.Int, n)
	mod := fr.Modulus()
	var err error
	for i := range vars {
		sorted[i], err = rand.Int(rand.Reader, mod)
		if err != nil {
			t.Fatal(err)
		}
		vars[i] = new(big.Int).Set(sorted[i])
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Cmp(sorted[j]) < 0
	})
	circuit.B = sorted
	witness := StupidPermutationCircuit{
		A: vars,
		B: sorted,
	}
	assert := test.NewAssert(t)
	assert.ProverSucceeded(&circuit, &witness,
		test.WithProverOpts(backend.WithHints(stupidSortHint)),
		test.WithCurves(ecc.BN254),
	)
}
