package repos2

type MemUserRepo struct {
	Message string
}

func (mur *MemUserRepo) Find() []string {
	return []string{"repos2", mur.Message}
}

type FileUserRepo struct {
	Message string
}

func (fur *FileUserRepo) Find() []string {
	return []string{"repos2", fur.Message}
}
