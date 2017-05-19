package repos

import "fmt"

type OrderRepo interface {
	Create()
	Remove(id string)
}

type MemOrderRepo struct {
	Count int `fig:"skip[true]"`
}

func (mor *MemOrderRepo) Create() {
	fmt.Println("create order in memory")
}

func (mor *MemOrderRepo) Remove(id string) {
	fmt.Println("remove", id, "order from memory")
}
