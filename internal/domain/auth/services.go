package auth

type LoginService interface {
	Login(name, password string) (string, error)
}
