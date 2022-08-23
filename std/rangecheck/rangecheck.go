package rangecheck

import (
	"fmt"
	"sync"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/bits"
	"github.com/consensys/gnark/std/permutation"
)

type Checker struct {
	m         *sync.Mutex
	collected []frontend.Variable
	commited  bool
	nbBits    int
}

func New(nbBits int) (*Checker, error) {
	if nbBits > 32 {
		return nil, fmt.Errorf("given range %d larger than 32", nbBits)
	}
	return &Checker{
		m:      new(sync.Mutex),
		nbBits: nbBits,
	}, nil
}

func (c *Checker) Check(in ...frontend.Variable) {
	c.m.Lock()
	defer c.m.Unlock()
	c.collected = append(c.collected, in...)
}

func (c *Checker) Commit(api frontend.API) error {
	c.m.Lock()
	defer c.m.Unlock()
	if c.commited {
		return fmt.Errorf("range checker already commited")
	}
	// we need to create dummy variables to ensure that there are no large gaps
	// nbBits is constrained to be no more than 32
	bound := 1 << c.nbBits
	dummy := make([]frontend.Variable, bound)
	for i := range dummy {
		dummy[i] = i
	}
	toSort := append(dummy, c.collected...)
	sorted := permutation.Sort(api, toSort)
	// first is 0
	api.AssertIsEqual(sorted[0], 0)
	// last contains of all 1 bits
	// dropping bits as we only want to assert
	bits.ToBinary(api, sorted[len(sorted)-1], bits.WithNbDigits(c.nbBits))
	// the difference between every value is 0 or 1
	for i := 1; i < len(sorted); i++ {
		tmp := api.Sub(sorted[i], sorted[i-1])
		api.AssertIsBoolean(tmp)
	}
	return nil
}
