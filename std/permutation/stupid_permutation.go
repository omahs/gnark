package permutation

import (
	"math/big"
	"sort"

	"github.com/consensys/gnark/frontend"
)

// TODO: remove the file. It is a simple substitute until fix the routing network.

func Sort(api frontend.API, in []frontend.Variable) []frontend.Variable {
	sorted, err := api.NewHint(StupidSortHint, len(in), in...)
	if err != nil {
		panic(err)
	}
	var inProd frontend.Variable = 1
	for i := range in {
		inProd = api.Mul(api.Add(in[i], 999), inProd)
	}
	var sortedProd frontend.Variable = 1
	for i := range sorted {
		sortedProd = api.Mul(api.Add(sorted[i], 999), sortedProd)
	}
	api.AssertIsEqual(inProd, sortedProd)
	return sorted
}

// sortHint creates a Waksman routing which returns the inputs sorted.
func StupidSortHint(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
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
