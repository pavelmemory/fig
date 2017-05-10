package repos

type UserRepo interface {
	Find() []string
}

type MemUserRepo struct {
	Message string
}

func (mur *MemUserRepo) Find() []string {
	return []string{"mem", mur.Message}
}

type FileUserRepo struct {
	Message string
}

func (fur *FileUserRepo) Find() []string {
	return []string{"file", fur.Message}
}
