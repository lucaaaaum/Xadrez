package auth

type LoginRequest struct {
	name     string
	password string
}

type LoginService interface {
	Login(request LoginRequest) (string, error)
}
