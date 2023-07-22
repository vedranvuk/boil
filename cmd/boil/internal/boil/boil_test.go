package boil

import (
	"fmt"
	"testing"
)

func TestMetamap(t *testing.T) {
	var m, e = LoadMetamap("/home/vedran/projects/boil")
	if e != nil {
		t.Fatal(e)
	}

	fmt.Printf("Dirs:\n\n")
	for k, v := range m {
		fmt.Printf("%s\t%v\n", k, v)
	}

	fmt.Println()

	fmt.Printf("With metadata:\n\n")
	for k, v := range m.WithMetadata() {
		fmt.Printf("%s\t%v\n", k, v)
	}

}