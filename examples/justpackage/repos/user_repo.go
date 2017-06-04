package repos

import "fmt"

type UserRepo interface {
	Find(name string)
	Save(name string)
}

type MemUserRepo struct {
	Prefix string
}

func (mur *MemUserRepo) Find(name string) {
	fmt.Println(mur.Prefix, "find user", name, "in memory")
}

func (mur *MemUserRepo) Save(name string) {
	fmt.Println(mur.Prefix, "save user", name, "to memory")
}

type FileUserRepo struct {
	Prefix string
}

func (fur *FileUserRepo) Find(name string) {
	fmt.Println(fur.Prefix, "find user", name, "in file")
}

func (fur *FileUserRepo) Save(name string) {
	fmt.Println(fur.Prefix, "save user", name, "to file")
}
