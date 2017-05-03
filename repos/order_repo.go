package repos

type OrderRepo interface {
	Make() bool
	Remove() int
}

type MemOrderRepo struct {
}

func (mor *MemOrderRepo) Make() bool {
	return true
}

func (mor *MemOrderRepo) Remove() int {
	return 3
}
