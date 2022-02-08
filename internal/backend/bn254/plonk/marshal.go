// Copyright 2020 ConsenSys Software Inc.
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

package plonk

import (
	curve "github.com/consensys/gnark-crypto/ecc/bn254"

	"errors"
	"io"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// WriteTo writes binary encoding of Proof to w
func (proof *Proof) WriteTo(w io.Writer) (int64, error) {
	enc := curve.NewEncoder(w)

	toEncode := []interface{}{
		&proof.LRO[0],
		&proof.LRO[1],
		&proof.LRO[2],
		&proof.Z,
		&proof.H[0],
		&proof.H[1],
		&proof.H[2],
	}

	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			return enc.BytesWritten(), err
		}
	}

	n, err := proof.BatchedProof.WriteTo(w)
	if err != nil {
		return n + enc.BytesWritten(), err
	}
	n2, err := proof.ZShiftedOpening.WriteTo(w)

	return n + n2 + enc.BytesWritten(), err
}

// ReadFrom reads binary representation of Proof from r
func (proof *Proof) ReadFrom(r io.Reader) (int64, error) {
	dec := curve.NewDecoder(r)
	toDecode := []interface{}{
		&proof.LRO[0],
		&proof.LRO[1],
		&proof.LRO[2],
		&proof.Z,
		&proof.H[0],
		&proof.H[1],
		&proof.H[2],
	}

	for _, v := range toDecode {
		if err := dec.Decode(v); err != nil {
			return dec.BytesRead(), err
		}
	}

	n, err := proof.BatchedProof.ReadFrom(r)
	if err != nil {
		return n + dec.BytesRead(), err
	}
	n2, err := proof.ZShiftedOpening.ReadFrom(r)
	return n + n2 + dec.BytesRead(), err
}

// WriteTo writes binary encoding of ProvingKey to w
func (pk *ProvingKey) WriteTo(w io.Writer) (n int64, err error) {
	// encode the verifying key
	n, err = pk.Vk.WriteTo(w)
	if err != nil {
		return
	}

	// fft domains
	n2, err := pk.DomainSmall.WriteTo(w)
	if err != nil {
		return
	}
	n += n2

	n2, err = pk.DomainBig.WriteTo(w)
	if err != nil {
		return
	}
	n += n2

	// sanity check len(Permutation) == 3*int(pk.DomainSmall.Cardinality)
	if len(pk.Permutation) != (3 * int(pk.DomainSmall.Cardinality)) {
		return n, errors.New("invalid permutation size, expected 3*domain cardinality")
	}

	enc := curve.NewEncoder(w)
	// note: type Polynomial, which is handled by default binary.Write(...) op and doesn't
	// encode the size (nor does it convert from Montgomery to Regular form)
	// so we explicitly transmit []fr.Element
	toEncode := []interface{}{
		([]fr.Element)(pk.Ql),
		([]fr.Element)(pk.Qr),
		([]fr.Element)(pk.Qm),
		([]fr.Element)(pk.Qo),
		([]fr.Element)(pk.CQk),
		([]fr.Element)(pk.LQk),
		([]fr.Element)(pk.S1Canonical),
		([]fr.Element)(pk.S2Canonical),
		([]fr.Element)(pk.S3Canonical),
		pk.Permutation,
	}

	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			return n + enc.BytesWritten(), err
		}
	}

	return n + enc.BytesWritten(), nil
}

// ReadFrom reads from binary representation in r into ProvingKey
func (pk *ProvingKey) ReadFrom(r io.Reader) (int64, error) {
	pk.Vk = &VerifyingKey{}
	n, err := pk.Vk.ReadFrom(r)
	if err != nil {
		return n, err
	}

	n2, err := pk.DomainSmall.ReadFrom(r)
	n += n2
	if err != nil {
		return n, err
	}

	n2, err = pk.DomainBig.ReadFrom(r)
	n += n2
	if err != nil {
		return n, err
	}

	pk.Permutation = make([]int64, 3*pk.DomainSmall.Cardinality)

	dec := curve.NewDecoder(r)
	toDecode := []interface{}{
		(*[]fr.Element)(&pk.Ql),
		(*[]fr.Element)(&pk.Qr),
		(*[]fr.Element)(&pk.Qm),
		(*[]fr.Element)(&pk.Qo),
		(*[]fr.Element)(&pk.CQk),
		(*[]fr.Element)(&pk.LQk),
		(*[]fr.Element)(&pk.S1Canonical),
		(*[]fr.Element)(&pk.S2Canonical),
		(*[]fr.Element)(&pk.S3Canonical),
		&pk.Permutation,
	}

	for _, v := range toDecode {
		if err := dec.Decode(v); err != nil {
			return n + dec.BytesRead(), err
		}
	}

	return n + dec.BytesRead(), nil

}

// WriteTo writes binary encoding of VerifyingKey to w
func (vk *VerifyingKey) WriteTo(w io.Writer) (n int64, err error) {
	enc := curve.NewEncoder(w)

	toEncode := []interface{}{
		vk.Size,
		&vk.SizeInv,
		&vk.Generator,
		vk.NbPublicVariables,
		&vk.S[0],
		&vk.S[1],
		&vk.S[2],
		&vk.Ql,
		&vk.Qr,
		&vk.Qm,
		&vk.Qo,
		&vk.Qk,
	}

	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			return enc.BytesWritten(), err
		}
	}

	return enc.BytesWritten(), nil
}

// ReadFrom reads from binary representation in r into VerifyingKey
func (vk *VerifyingKey) ReadFrom(r io.Reader) (int64, error) {
	dec := curve.NewDecoder(r)
	toDecode := []interface{}{
		&vk.Size,
		&vk.SizeInv,
		&vk.Generator,
		&vk.NbPublicVariables,
		&vk.S[0],
		&vk.S[1],
		&vk.S[2],
		&vk.Ql,
		&vk.Qr,
		&vk.Qm,
		&vk.Qo,
		&vk.Qk,
	}

	for _, v := range toDecode {
		if err := dec.Decode(v); err != nil {
			return dec.BytesRead(), err
		}
	}

	return dec.BytesRead(), nil
}
