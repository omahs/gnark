/*
Package permutation implements AS-Waksman routing network.

Arbitrary size (AS) Waksman routing network is a network of layered switches
between two wires which allows to reorder the inputs in any order by defining
the switch states.

See https://hal.inria.fr/inria-00072871/document.
*/
package permutation

import (
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
)

type permutation [][2]int

func (p permutation) isValid() bool {
	// all indices must exist
	a1 := make(map[int]struct{})
	a2 := make(map[int]struct{})
	for i := range p {
		if _, ok := a1[p[i][0]]; ok {
			return false
		}
		if _, ok := a2[p[i][1]]; ok {
			return false
		}
		a1[p[i][0]] = struct{}{}
		a2[p[i][1]] = struct{}{}
	}
	if len(a1) != len(p) || len(a2) != len(p) {
		return false
	}
	for i := 0; i < len(p); i++ {
		if _, ok := a1[i]; !ok {
			return false
		}
		if _, ok := a2[i]; !ok {
			return false
		}
	}
	return true
}

func sortedWithPermutation(in []*big.Int) permutation {
	p := make(permutation, len(in))
	for i := range p {
		p[i][0] = i
		p[i][1] = i
	}
	sort.Slice(p, func(i, j int) bool {
		return in[p[i][0]].Cmp(in[p[j][0]]) < 0
	})
	for i := range p {
		p[i][1] = i
	}
	return p
}

func sortedWithPermutation2(in []int) permutation {
	p := make(permutation, len(in))
	for i := range p {
		p[i][0] = i
		p[i][1] = i
	}
	sort.Slice(p, func(i, j int) bool {
		return in[p[i][1]] < in[p[j][1]]
	})
	for i := range p {
		p[i][1] = i
	}
	return p
}

func permutationFromMapping(before, after []int) permutation {
	if len(before) != len(after) {
		panic("diff lengths")
	}
	return nil
}

type vertex struct {
	vals  []int
	edges []*edge
	index int
}

func (v vertex) String() string {
	var es []string
	for _, e := range v.edges {
		es = append(es, e.String())
	}
	var vs []string
	for _, vv := range v.vals {
		vs = append(vs, strconv.Itoa(vv))
	}
	return fmt.Sprintf("V([%s], {%s})",
		strings.Join(vs, ","), strings.Join(es, ","))
}

func (v vertex) degreeUnknown() int {
	var d int
	for _, e := range v.edges {
		if e.direction == none {
			d++
		}
	}
	return d
}

type direction string

const (
	up   direction = "UP"
	down direction = "DOWN"
	none direction = "?"
)

func (d direction) other() direction {
	switch d {
	case up:
		return down
	case down:
		return up
	default:
		return none
	}
}

type edge struct {
	vertices   [2]*vertex
	permPoints [2]int
	direction
}

func (e edge) String() string {
	return fmt.Sprintf("E(%d <-> %d: direction: %s)",
		e.permPoints[0], e.permPoints[1], e.direction)
}

type bipartite struct {
	left  []*vertex
	right []*vertex
	edges []*edge
	permutation
	isColored bool
	isOdd     bool
}

func newBipartite(p permutation) *bipartite {
	if !p.isValid() {
		return nil
	}
	bp := bipartite{
		left:        make([]*vertex, (len(p)+1)/2),
		right:       make([]*vertex, (len(p)+1)/2),
		isOdd:       len(p)%2 == 1,
		permutation: p,
		isColored:   false,
	}
	for i := 0; i < len(p)/2; i++ {
		bp.left[i] = &vertex{
			vals:  make([]int, 2),
			index: i,
		}
		bp.right[i] = &vertex{
			vals:  make([]int, 2),
			index: i,
		}
	}
	if bp.isOdd {
		bp.left[len(p)/2] = &vertex{
			vals:  make([]int, 1),
			index: len(p) / 2,
		}
		bp.right[len(p)/2] = &vertex{
			vals:  make([]int, 1),
			index: len(p) / 2,
		}
	}
	// m := make(map[int]int)
	// for i, pp := range p {
	// 	m[pp[0]] = i
	// }
	for _, pp := range p {
		// bp.left[m[pp[0]]/2].vals[m[pp[0]]%2] = pp[0]
		// bp.right[m[pp[1]]/2].vals[m[pp[1]]%2] = pp[0]
		bp.left[pp[0]/2].vals[pp[0]%2] = pp[0]
		bp.right[pp[1]/2].vals[pp[1]%2] = pp[0]
		edge := &edge{
			vertices: [2]*vertex{
				// bp.left[m[pp[0]]/2],
				// bp.right[m[pp[1]]/2],
				bp.left[pp[0]/2],
				bp.right[pp[1]/2],
			},
			permPoints: [2]int{pp[0], pp[1]},
			direction:  none,
		}
		edge.vertices[0].edges = append(edge.vertices[0].edges, edge)
		edge.vertices[1].edges = append(edge.vertices[1].edges, edge)
		bp.edges = append(bp.edges, edge)
	}
	return &bp
}

func (bp bipartite) String() string {
	var ls, rs []string
	for _, l := range bp.left {
		ls = append(ls, l.String())
	}
	for _, r := range bp.right {
		rs = append(rs, r.String())
	}
	return fmt.Sprintf("left %s\nright %s",
		strings.Join(ls, "\n"), strings.Join(rs, "\n"))
}

func (bp bipartite) degreeUnknown() int {
	var d int
	for _, l := range bp.left {
		d += l.degreeUnknown()
	}
	for _, r := range bp.right {
		d += r.degreeUnknown()
	}
	return d / 2
}

func (bp *bipartite) color() {
	if bp.isColored {
		return
	}
	if bp.isOdd {
		// the lower subnetwork is always larger if the subnetwork are uneven.
		bp.left[len(bp.left)-1].edges[0].direction = down
		bp.right[len(bp.right)-1].edges[0].direction = down
	} else {
		// must ensure that the lower right does not swap. set the edge
		// direction which enforces that.
		if bp.right[len(bp.right)-1].vals[0] == bp.right[len(bp.right)-1].edges[0].permPoints[0] {
			bp.right[len(bp.right)-1].edges[0].direction = up
			bp.right[len(bp.right)-1].edges[1].direction = down
		} else {
			bp.right[len(bp.right)-1].edges[0].direction = down
			bp.right[len(bp.right)-1].edges[1].direction = up
		}
	}
	allOtherColor := func(vs []*vertex) bool {
		var colored bool
		for _, v := range vs {
			if v.degreeUnknown() == 1 {
				if v.edges[0].direction != none {
					v.edges[1].direction = v.edges[0].other()
				} else {
					v.edges[0].direction = v.edges[1].other()
				}
				colored = true
			}
		}
		return colored
	}
	for bp.degreeUnknown() > 0 {
		allOtherColor(bp.left)
		allOtherColor(bp.right)
		for _, v := range bp.left {
			if v.degreeUnknown() == 2 {
				v.edges[0].direction = up
				v.edges[1].direction = down
				break
			}
		}
	}
	bp.isColored = true
}

type switchState bool

func (ss switchState) String() string {
	switch ss {
	case straight:
		return "straight"
	case swap:
		return "swap"
	}
	panic("invalid")
}

const (
	straight switchState = false
	swap     switchState = true
)

func (bp *bipartite) switchStates() (pre, post []switchState) {
	if !bp.isColored {
		bp.color()
	}
	l := len(bp.left)
	if bp.isOdd {
		l--
	}
	pre = make([]switchState, l)
	post = make([]switchState, l)
	for i := 0; i < l; i++ {
		pre[i] = (bp.left[i].edges[0].direction == up) != (bp.left[i].vals[0] == bp.left[i].edges[0].permPoints[0])
		post[i] = (bp.right[i].edges[0].direction == up) != (bp.right[i].vals[0] == bp.right[i].edges[0].permPoints[0])
	}
	if bp.isOdd {
		pre = append(pre, straight)
		post = append(post, straight)
	}
	return
}

// func (bp *bipartite) graphs() (bipartite, bipartite) {
// 	pre, post := bp.switchStates()
// 	var ue, le []*edge
// 	for _, e := range bp.edges {
// 		e := &edge{
// 			permPoints: e.permPoints,
// 			direction:  none,
// 		}
// 		if e.direction == up {
// 			ue = append(ue, e)
// 		} else {
// 			le = append(le, e)
// 		}
// 	}
// }

func (bp *bipartite) innerPermutation() permutation {
	// pre, post := bp.switchStates()
	// p := make(permutation, len(bp.permutation))
	// for i := range pre {
	// 	var t1, t2 int
	// 	if pre[i] == swap {
	// 		t1 = 1
	// 	}
	// 	if post[i] == swap {
	// 		t2 = 1
	// 	}
	// 	p1 := bp.permutation[2*i+t1][0]
	// 	p2 := bp.permutation[2*i+(1-t2)]
	// }
	// if bp.isOdd {
	// 	p[len(p)-1] = bp.permutation[len(p)-1]
	// }
	// for i := 0; i < len(bp.left); i++ {
	// 	p1 := bp.left[i].vals[t1]
	// 	var p2 int

	// 	p2 = bp.left[i].edges[0].permPoints[t2]
	// 	// p2 := bp.right[i].vals[t2]
	// 	p = append(p, [2]int{p1, p2})
	// 	if len(bp.left[i].edges) == 2 {
	// 		pp1 := bp.left[i].vals[1-t1]
	// 		pp2 := bp.left[i].edges[1].permPoints[1-t2]
	// 		// pp2 := bp.right[i].vals[1-t2]
	// 		p = append(p, [2]int{pp1, pp2})
	// 	}
	// }
	return nil
}

func routing(p permutation) [][]switchState {
	bp := newBipartite(p)
	pre, post := bp.switchStates()
	innerPerm := bp.innerPermutation()
	fmt.Println(p, pre, post, innerPerm)
	fmt.Println(bp)
	return nil
	if len(p) == 1 {
		return [][]switchState{}
	}
	if len(p) == 2 {
		return [][]switchState{{pre[0] != post[0]}}
	}
	var perms [2]permutation
	for i, pp := range innerPerm {
		if i == len(innerPerm)-1 {
			perms[1] = append(perms[1], pp)
		} else {
			perms[i%2] = append(perms[i%2], pp)
		}
	}
	states := [][][]switchState{routing(perms[0]), routing(perms[1])}
	res := [][]switchState{pre}
	for i := 0; i < len(states[0]); i++ {
		var layer []switchState
		layer = append(layer, states[0][i]...)
		layer = append(layer, states[1][i]...)
		res = append(res, layer)
	}
	res = append(res, post)
	return res
}
