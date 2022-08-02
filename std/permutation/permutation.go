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

type vertex struct {
	vals  []int
	edges []*edge
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
	left      []*vertex
	right     []*vertex
	isColored bool
	isOdd     bool
}

func newBipartite(p permutation) *bipartite {
	bp := bipartite{
		left:  make([]*vertex, (len(p)+1)/2),
		right: make([]*vertex, (len(p)+1)/2),
		isOdd: len(p)%2 == 1,
	}
	for i := 0; i < len(p)/2; i++ {
		bp.left[i] = &vertex{
			vals: make([]int, 2),
		}
		bp.right[i] = &vertex{
			vals: make([]int, 2),
		}
	}
	if bp.isOdd {
		bp.left[len(p)/2] = &vertex{
			vals: make([]int, 1),
		}
		bp.right[len(p)/2] = &vertex{
			vals: make([]int, 1),
		}
	}
	m := make(map[int]int)
	for i, pp := range p {
		m[pp[0]] = i
	}
	for _, pp := range p {
		bp.left[m[pp[0]]/2].vals[m[pp[0]]%2] = pp[0]
		bp.right[m[pp[1]]/2].vals[m[pp[1]]%2] = pp[0]
		edge := &edge{
			vertices: [2]*vertex{
				bp.left[m[pp[0]]/2],
				bp.right[m[pp[1]]/2],
			},
			permPoints: [2]int{pp[0], pp[1]},
			direction:  none,
		}
		edge.vertices[0].edges = append(edge.vertices[0].edges, edge)
		edge.vertices[1].edges = append(edge.vertices[1].edges, edge)
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

func routing(p permutation) [][]switchState {
	return nil
}

// sortHint creates a Waksman routing which returns the inputs sorted. The hint
// returns the (upper) values of all switches in order.
func sortHint(_ *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	return nil
}
