package rangecheck

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/permutation"
	"github.com/consensys/gnark/test"
)

type RangeCheckCircuit struct {
	A      []frontend.Variable
	nbBits int
}

func (c *RangeCheckCircuit) Define(api frontend.API) error {
	checker, err := New(c.nbBits)
	if err != nil {
		return fmt.Errorf("new checker: %w", err)
	}
	for i := range c.A {
		checker.Check(c.A[i])
	}
	err = checker.Commit(api)
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func TestRangeCheck(t *testing.T) {
	assert := test.NewAssert(t)
	for _, nbBits := range []int{8, 12, 16} {
		for _, count := range []int{13, 110, 1000, 16384, 65536, 131072} {
			assert.Run(func(assert *test.Assert) {
				rangeCheckSubTest(assert, nbBits, count)
			}, fmt.Sprintf("nbBits=%d/count=%d", nbBits, count))
		}
	}
}

func TestRangeCheck1(t *testing.T) {
	rangeCheckSubTest(test.NewAssert(t), 3, 8)
}

func rangeCheckSubTest(assert *test.Assert, nbBits, count int) {
	bound := big.NewInt(1 << nbBits)
	vars := make([]frontend.Variable, count)
	for i := range vars {
		tmp, err := rand.Int(rand.Reader, bound)
		assert.NoError(err)
		vars[i] = tmp
	}
	circuit := RangeCheckCircuit{A: make([]frontend.Variable, count), nbBits: nbBits}
	witness := RangeCheckCircuit{A: vars, nbBits: nbBits}
	assert.ProverSucceeded(&circuit, &witness,
		test.WithCurves(ecc.BN254),
		test.WithProverOpts(backend.WithHints(permutation.StupidSortHint)))
}
