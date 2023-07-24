package boil

import (
	"fmt"
	"sort"
	"testing"
)

func TestMetamap(t *testing.T) {
	var dr = &DiskRepository{"/home/vedran/projects/boil/_testdata/templates"}
	var a []string
	var m, e = dr.LoadMetamap()
	if e != nil {
		t.Fatal(e)
	}
	for k := range m {
		a = append(a, k)
	}
	sort.Strings(a)
	for _, v := range a {
		s := "nil"
		if m[v] != nil {
			s = m[v].Name
		}
		fmt.Printf("%s\t%v\n", v, s)
	}
}
