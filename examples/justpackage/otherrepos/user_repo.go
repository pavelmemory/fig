package otherrepos

import (
	"fmt"
	"strings"
)

type MemUserRepo struct {
}

func (mur *MemUserRepo) Find(name string) {
	fmt.Println("find user", strings.ToUpper(name), "in memory")
}

func (mur *MemUserRepo) Save(name string) {
	fmt.Println("save user", strings.ToUpper(name), "to memory")
}

type FileUserRepo struct {
}

func (fur FileUserRepo) Find(name string) {
	fmt.Println("find user", strings.ToUpper(name), "in file")
}

func (fur FileUserRepo) Save(name string) {
	fmt.Println("save user", strings.ToUpper(name), "to file")
}
