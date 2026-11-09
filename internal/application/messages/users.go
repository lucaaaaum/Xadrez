package messages

type ListOnlineUsersRequest struct {
}

type ListOnlineUsersResponse struct {
	Users []string `json:"users"`
}
