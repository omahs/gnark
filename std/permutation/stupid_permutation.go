package permutation

import (
	"math/big"
	"sort"

	"github.com/consensys/gnark/frontend"
)

// TODO: remove the file. It is a simple substitute until fix the routing network.

func Sort(api frontend.API, in []frontend.Variable) []frontend.Variable {
	sorted, err := api.NewHint(stupidSortHint, len(in), in...)
	if err != nil {
		panic(err)
	}
	var leftProd frontend.Variable = 1
	var rightProd frontend.Variable = 1
	for i := range in {
		leftProd = api.Mul(leftProd, in[i])
		rightProd = api.Mul(rightProd, sorted[i])
	}
	api.AssertIsEqual(leftProd, rightProd)
	return sorted
}

// sortHint creates a Waksman routing which returns the inputs sorted.
func stupidSortHint(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	sorted := make([]*big.Int, len(inputs))
	for i := range sorted {
		sorted[i] = new(big.Int).Set(inputs[i])
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Cmp(sorted[j]) < 0
	})
	for i := range sorted {
		outputs[i].Set(sorted[i])
	}
	return nil
}
