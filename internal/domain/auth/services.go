package auth

type LoginRequest struct {
	Name     string
	Password string
}

type LoginService interface {
	Login(request LoginRequest) (string, error)
}
