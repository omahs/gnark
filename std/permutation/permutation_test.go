package permutation

import (
	"math/big"
	"testing"
)

// type Circuit struct {
// 	In []frontend.Variable
// }

// func (c *Circuit) Define(api frontend.API) error {
// 	var inSum frontend.Variable = 0
// 	var permSum frontend.Variable = 0
// 	permuted := Permute(api, c.In)
// 	if len(permuted) != len(c.In) {
// 		return fmt.Errorf("permuted length mismatch")
// 	}
// 	for i := range permuted {
// 		inSum = api.Add(inSum, c.In[i])
// 		permSum = api.Add(permSum, permuted[i])
// 	}
// 	api.AssertIsEqual(inSum, permSum)
// 	return nil
// }

// func TestShuffle(t *testing.T) {
// 	assert := test.NewAssert(t)
// 	var err error
// 	mod := ecc.BN254.ScalarField()
// 	n := 50
// 	in := make([]frontend.Variable, n)
// 	for i := range in {
// 		in[i], err = rand.Int(rand.Reader, mod)
// 		assert.NoError(err)
// 	}
// 	circuit := &Circuit{
// 		In: make([]frontend.Variable, n),
// 	}
// 	assignment := &Circuit{
// 		In: in,
// 	}
// 	assert.ProverSucceeded(circuit, assignment, test.WithCurves(ecc.BN254), test.WithProverOpts(backend.WithHints(SwitchHint)))
// }

func TestSortingPermutation(t *testing.T) {
	input := []*big.Int{
		big.NewInt(3),
		big.NewInt(4),
		big.NewInt(1),
		big.NewInt(2),
		big.NewInt(0),
	}
	perm := sortedWithPermutation(input)
	output := make([]*big.Int, len(input))
	for i := range perm {
		output[perm[i][1]] = input[perm[i][0]]
	}
}

func TestNewBipartite(t *testing.T) {
	p := permutation{
		// {0, 7}, {1, 6}, {2, 5}, {3, 8}, {4, 0}, {5, 3}, {6, 2}, {7, 1}, {8, 4},
		// {0, 6}, {2, 4}, {4, 0}, {6, 2},
		// {2, 4}, {4, 2},
		// {0, 0}, {1, 2}, {2, 1},
		// {0,1}, {1,0},
		{0, 0}, {1, 1},
		// {0, 0},
	}
	bp := newBipartite(p)
	t.Log(bp)
	bp.color()
	t.Log(bp)
	pre, post := bp.switchStates()
	t.Log(pre)
	t.Log(post)
}

func TestRouting(t *testing.T) {
	p := permutation{
		{0, 7}, {1, 6}, {2, 5}, {3, 8}, {4, 0}, {5, 3}, {6, 2}, {7, 1}, {8, 4},
		// {0, 6}, {2, 4}, {4, 0}, {6, 2},
		// {0, 3}, {1, 2}, {2, 0}, {3, 1},
		// {2, 4}, {4, 2},
		// {0, 0}, {1, 2}, {2, 1},
		// {0,1}, {1,0},
		// {0, 0}, {1, 1},
		// {0, 0},
	}
	ss := routing(p)
	t.Log(ss)
}

func TestPerm2(t *testing.T) {
	a := []int{4, 7, 6, 5, 2, 8, 0, 1, 3}
	p := sortedWithPermutation2(a)
	t.Log(p)
}
