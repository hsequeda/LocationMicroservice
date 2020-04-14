package actors

import "locationMicroService/libs/util"

type Admin struct {
	Id       int
	UserName string
	PassHash string
}

// NewAdmin returns a new instance of Admin
func NewAdmin(userName string, password string) (*Admin, error) {
	passwordHash, err := util.GeneratePasswordHash(password)
	if err != nil {
		return nil, err
	}

	return &Admin{UserName: userName, PassHash: passwordHash}, nil
}
