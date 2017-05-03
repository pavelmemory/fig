package repos

type BookRepo interface {
	Get() string
	Add(string)
}

type FileBookRepo struct {
}

func (fbr FileBookRepo) Get() string {
	return "qwqe"
}

func (fbr FileBookRepo) Add(string) {

}
