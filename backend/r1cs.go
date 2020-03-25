// Copyright 2020 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by gnark/internal/generators DO NOT EDIT

package backend

import (
	"fmt"

	"github.com/consensys/gnark/curve"
	"github.com/consensys/gnark/curve/fr"

	"github.com/consensys/gnark/internal/utils/debug"
	"github.com/consensys/gnark/internal/utils/encoding/gob"
)

// R1CS decsribes a set of R1CS constraint
type R1CS struct {
	// Wires
	NbWires        int
	NbPublicWires  int // includes ONE wire
	NbPrivateWires int
	PrivateWires   []string         // private wire names
	PublicWires    []string         // public wire names
	WireTags       map[int][]string // optional tags -- debug info

	// Constraints
	NbConstraints   int // total number of constraints
	NbCOConstraints int // number of constraints that need to be solved, the first of the Constraints slice
	Constraints     []R1C
}

// Solve sets all the wires and returns the a, b, c vectors.
// the r1cs system should have been compiled before. The entries in a, b, c are in Montgomery form.
// assignment: map[string]value: contains the input variables
// a, b, c vectors: ab-c = hz
// wireValues =  [intermediateVariables | privateInputs | publicInputs]
func (r1cs *R1CS) Solve(assignment Assignments, a, b, c, wireValues []fr.Element) error {

	// compute the wires and the a, b, c polynomials
	debug.Assert(len(a) == r1cs.NbConstraints)
	debug.Assert(len(b) == r1cs.NbConstraints)
	debug.Assert(len(c) == r1cs.NbConstraints)
	debug.Assert(len(wireValues) == r1cs.NbWires)

	// keep track of wire that have a value
	wireInstantiated := make([]bool, r1cs.NbWires)

	// instantiate the public/ private inputs
	instantiateInputs := func(offset int, visibility Visibility, inputNames []string) error {
		for i := 0; i < len(inputNames); i++ {
			name := inputNames[i]
			if name == OneWire {
				wireValues[i+offset].SetOne()
				wireInstantiated[i+offset] = true
			} else {
				if val, ok := assignment[name]; ok {
					if visibility == Secret && val.IsPublic || visibility == Public && !val.IsPublic {
						return fmt.Errorf("%q: %w", name, ErrInputVisiblity)
					}
					wireValues[i+offset].Set(&val.Value)
					wireInstantiated[i+offset] = true
				} else {
					return fmt.Errorf("%q: %w", name, ErrInputNotSet)
				}
			}
		}
		return nil
	}
	// instantiate private inputs
	debug.Assert(len(r1cs.PrivateWires) == r1cs.NbPrivateWires)
	debug.Assert(len(r1cs.PublicWires) == r1cs.NbPublicWires)
	if r1cs.NbPrivateWires != 0 {
		offset := r1cs.NbWires - r1cs.NbPublicWires - r1cs.NbPrivateWires // private input start index
		if err := instantiateInputs(offset, Secret, r1cs.PrivateWires); err != nil {
			return err
		}
	}
	// instantiate public inputs
	{
		offset := r1cs.NbWires - r1cs.NbPublicWires // public input start index
		if err := instantiateInputs(offset, Public, r1cs.PublicWires); err != nil {
			return err
		}
	}

	// check if there is an inconsistant constraint
	var check fr.Element

	// Loop through the other Constraints
	for i, r1c := range r1cs.Constraints {

		if i < r1cs.NbCOConstraints {
			// computationalGraph : we need to solve the constraint
			// computationalGraph[i] contains exactly one uncomputed wire (due
			// to the graph being correctly ordered), we solve it
			r1cs.Constraints[i].solveR1c(wireInstantiated, wireValues)
		}

		// A this stage we are not guaranteed that a[i+sizecg]*b[i+sizecg]=c[i+sizecg] because we only query the values (computed
		// at the previous step)
		a[i], b[i], c[i] = r1c.instantiate(r1cs, wireValues)

		// check that the constraint is satisfied
		check.Mul(&a[i], &b[i])
		if !check.Equal(&c[i]) {
			invalidA := a[i]
			invalidB := b[i]
			invalidC := c[i]

			return fmt.Errorf("%w: %q * %q != %q", ErrUnsatisfiedConstraint,
				invalidA.String(),
				invalidB.String(),
				invalidC.String())
		}
	}

	return nil
}

// Inspect returns the tagged variables with their corresponding value
func (r1cs *R1CS) Inspect(wireValues []fr.Element) (map[string]fr.Element, error) {
	res := make(map[string]fr.Element)

	for wireID, tags := range r1cs.WireTags {
		for _, tag := range tags {
			if _, ok := res[tag]; ok {
				// TODO checking duplicates should be done in the frontend, probably in cs.ToR1CS()
				return nil, ErrDuplicateTag
			}
			res[tag] = wireValues[wireID]
		}

	}
	return res, nil
}

// method to solve a r1cs
type solvingMethod int

const (
	SingleOutput solvingMethod = iota
	BinaryDec
)

// Term lightweight version of a term, no pointers
type Term struct {
	ID    int64      // index of the constraint used to compute this wire
	Coeff fr.Element // coefficient by which the wire is multiplied
}

// LinearExpression lightweight version of linear expression
type LinearExpression []Term

// R1C used to compute the wires (wo pointers)
type R1C struct {
	L      LinearExpression
	R      LinearExpression
	O      LinearExpression
	Solver solvingMethod
}

// compute left, right, o part of a r1cs constraint
// this function is called when all the wires have been computed
// it instantiates the l, r o part of a R1C
func (r1c *R1C) instantiate(r1cs *R1CS, wireValues []fr.Element) (a, b, c fr.Element) {

	var tmp fr.Element

	for _, t := range r1c.L {
		debug.Assert(len(wireValues) > int(t.ID), "trying to access out of bound wire in wiretracker")
		tmp.Mul(&t.Coeff, &wireValues[t.ID])
		a.Add(&a, &tmp)
	}

	for _, t := range r1c.R {
		debug.Assert(len(wireValues) > int(t.ID), "trying to access out of bound wire in wiretracker")
		tmp.Mul(&t.Coeff, &wireValues[t.ID])
		b.Add(&b, &tmp)
	}

	for _, t := range r1c.O {
		debug.Assert(len(wireValues) > int(t.ID), "trying to access out of bound wire in wiretracker")
		tmp.Mul(&t.Coeff, &wireValues[t.ID])
		c.Add(&c, &tmp)
	}

	return
}

// solveR1c computes a wire by solving a r1cs
// the function searches for the unset wire (either the unset wire is
// alone, or it can be computed without ambiguity using the other computed wires
// , eg when doing a binary decomposition: either way the missing wire can
// be computed without ambiguity because the r1cs is correctly ordered)
func (r1c *R1C) solveR1c(wireInstantiated []bool, wireValues []fr.Element) {

	switch r1c.Solver {

	// in this case we solve a R1C by isolating the uncomputed wire
	case SingleOutput:

		// the index of the non zero entry shows if L, R or O has an uninstantiated wire
		// the content is the ID of the wire non instantiated
		location := [3]int64{-1, -1, -1}

		var tmp, a, b, c, backupCoeff fr.Element

		for _, t := range r1c.L {
			if wireInstantiated[t.ID] {
				tmp.Mul(&t.Coeff, &wireValues[t.ID])
				a.Add(&a, &tmp)
			} else {
				backupCoeff.Set(&t.Coeff)
				location[0] = t.ID
			}
		}

		for _, t := range r1c.R {
			if wireInstantiated[t.ID] {
				tmp.Mul(&t.Coeff, &wireValues[t.ID])
				b.Add(&b, &tmp)
			} else {
				backupCoeff.Set(&t.Coeff)
				location[1] = t.ID
			}
		}

		for _, t := range r1c.O {
			if wireInstantiated[t.ID] {
				tmp.Mul(&t.Coeff, &wireValues[t.ID])
				c.Add(&c, &tmp)
			} else {
				backupCoeff.Set(&t.Coeff)
				location[2] = t.ID
			}
		}

		var zero fr.Element

		if location[0] != -1 {
			id := location[0]
			if b.Equal(&zero) {
				wireValues[id].SetZero()
			} else {
				wireValues[id].Div(&c, &b).
					Sub(&wireValues[id], &a).
					Mul(&wireValues[id], &backupCoeff)
			}
			wireInstantiated[id] = true
		} else if location[1] != -1 {
			id := location[1]
			if a.Equal(&zero) {
				wireValues[id].SetZero()
			} else {
				wireValues[id].Div(&c, &a).
					Sub(&wireValues[id], &b).
					Mul(&wireValues[id], &backupCoeff)
			}
			wireInstantiated[id] = true
		} else if location[2] != -1 {
			id := location[2]
			wireValues[id].Mul(&a, &b).
				Sub(&wireValues[id], &c).
				Mul(&wireValues[id], &backupCoeff)
			wireInstantiated[id] = true
		}

	// in the case the R1C is solved by directly computing the binary decomposition
	// of the variable
	case BinaryDec:

		// the binary decomposition must be called on the non Mont form of the number
		n := wireValues[r1c.O[0].ID].ToRegular()
		nbBits := len(r1c.L)

		// binary decomposition of n
		var i, j int
		for i*64 < nbBits {
			j = 0
			for j < 64 && i*64+j < len(r1c.L) {
				ithbit := (n[i] >> uint(j)) & 1
				if !wireInstantiated[r1c.L[i*64+j].ID] {
					wireValues[r1c.L[i*64+j].ID].SetUint64(ithbit)
					wireInstantiated[r1c.L[i*64+j].ID] = true
				}
				j++
			}
			i++
		}
	default:
		panic("unimplemented solving method")
	}
}

func (r1cs *R1CS) Write(path string) error {
	return gob.Write(path, r1cs, curve.ID)
}
