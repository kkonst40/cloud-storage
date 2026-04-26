package profile

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Username string `json:"username"`
}

type RegisterResponse struct {
	Username string `json:"username"`
}

type ProfileResponse struct {
	Username string `json:"username"`
}
