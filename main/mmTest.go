package main

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var (
	PAGE = 16
	WIPE = make([]byte, PAGE)
)

type Mmap []byte

func (mm Mmap) Add(d []byte) int {
	if len(d) > PAGE {
		return -1
	}
	for i := 0; i < len(mm); i += PAGE {
		if mm[i] == 0x00 {
			copy(mm[i:i+PAGE], WIPE)
			copy(mm[i:i+PAGE], d)
			return i / PAGE
		}
	}
	return -1
}

func (mm Mmap) Set(d []byte, n int) int {
	of := n * PAGE
	if len(d) > PAGE {
		return -1
	}
	copy(mm[of:of+PAGE], WIPE)
	copy(mm[of:of+PAGE], d)
	return n
}

func (mm Mmap) Get(n int) []byte {
	of := n * PAGE
	if of > len(mm)-1 || mm[of] == 0x00 {
		return nil
	}
	return mm[of : of+PAGE]
}

func (mm Mmap) Del(n int) {
	of := n * PAGE
	if of < len(mm)-1 {
		copy(mm[of:of+PAGE], WIPE)
	}
}

func (mm Mmap) String() string {
	var pages []string
	for i := 0; i < len(mm); i += PAGE {
		pages = append(pages, strconv.Itoa(i/PAGE)+`=[`+string(mm[i:i+PAGE])+`]`)
	}
	return strings.Join(pages, ",")
}

func (mm Mmap) Compact() {
	var eb, fb, fe int = -1, -1, -1
	for i := 0; i < len(mm); i += PAGE {
		if mm[i] == 0x00 {
			if eb == -1 {
				eb = i
			} else if fe == -1 && fb != -1 {
				fe = i
			}
		} else if fb == -1 && eb != -1 {
			fb = i
		}
		if eb != -1 && fb != -1 && (fe != -1 || i == len(mm)-PAGE) {
			if i == len(mm)-PAGE {
				fe = len(mm)
			}
			dif := fb - eb
			ln := fe - fb
			copy(mm[eb:(eb+ln)], mm[fb:fe])
			copy(mm[(fe-dif):fe], make([]byte, dif))
			i = fe - dif - PAGE
			eb, fb, fe = -1, -1, -1
		}
	}
}

func (mm Mmap) Len() int {
	return len(mm) / PAGE
}

func (mm Mmap) Less(i, j int) bool {
	pi, pj := i*PAGE, j*PAGE

	if mm[pi] == 0x00 {
		if mm[pi] == mm[pj] {
			return true
		}
		return false
	}
	if mm[pj] == 0x00 {
		return true
	}

	return bytes.Compare(mm[pi:pi+PAGE], mm[pj:pj+PAGE]) == -1

}

func (mm Mmap) Swap(i, j int) {
	pi, pj := i*PAGE, j*PAGE
	tmp := make([]byte, PAGE)
	copy(tmp, mm[pi:pi+PAGE])
	copy(mm[pi:pi+PAGE], mm[pj:pj+PAGE])
	copy(mm[pj:pj+PAGE], tmp)
}

func main() {
	m := make(Mmap, 256) // 16 pages
	fillAndPrint(m)
	holeAndPrint(m)
	//m.Compact()
	//fmt.Printf("%s\n\n", m)
	sort.Stable(m)
	fmt.Printf("%s\n\n", m)
}

func holeAndPrint(m Mmap) {
	m.Del(2)
	m.Del(2)
	m.Del(5)
	m.Del(6)
	m.Del(9)
	m.Del(12)
	m.Del(13)
	m.Del(14)
	fmt.Printf("%s\n\n", m)
}

func fillAndPrint(m Mmap) {
	m.Add([]byte(`dog`))
	m.Add([]byte(`cat`))
	m.Add([]byte(`bird`))
	m.Add([]byte(`worm`))
	m.Add([]byte(`monkey`))
	m.Add([]byte(`duck`))
	m.Add([]byte(`dragon`))
	m.Add([]byte(`pirate`))
	m.Add([]byte(`sheep`))
	m.Add([]byte(`fish`))
	m.Add([]byte(`donkey`))
	m.Add([]byte(`dogg`))
	m.Add([]byte(`horse`))
	m.Add([]byte(`cow`))
	m.Add([]byte(`chicken`))
	m.Add([]byte(`dogggy`))
	fmt.Printf("%s\n\n", m)
}
