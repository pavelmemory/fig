package repos

type Module struct {
	Name      string `fig:"env[ENV_NAME]"`
	UserRepo  `fig:"impl[github.com/pavelmemory/fig/repos/FileUserRepo]"`
	BookRepo  `fig:""`
	OrderRepo `fig:""`
}
