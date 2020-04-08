package main

import "github.com/uber/h3-go"

const (
	Generic         = "GENERIC"
	ServiceProvider = "SERVICE_PROVIDER"
	Client          = "CLIENT"
)

type Category string

type User struct {
	Id          int         `json:"id"`
	Name        string      `json:"name"`
	Category    Category    `json:"category"`
	GeoCord     h3.GeoCoord `json:"geo_cord"`
	H3Positions []int64
}

var cacheUsers []*User

func NewUser(name string, lat, long float64, category Category) *User {
	geoCord := h3.GeoCoord{
		Latitude:  lat,
		Longitude: long,
	}

	return &User{
		Name:     name,
		Category: category,
		GeoCord:  geoCord,
		H3Positions: []int64{
			0:  int64(h3.FromGeo(geoCord, 0)),
			1:  int64(h3.FromGeo(geoCord, 1)),
			2:  int64(h3.FromGeo(geoCord, 2)),
			3:  int64(h3.FromGeo(geoCord, 3)),
			4:  int64(h3.FromGeo(geoCord, 4)),
			5:  int64(h3.FromGeo(geoCord, 5)),
			6:  int64(h3.FromGeo(geoCord, 6)),
			7:  int64(h3.FromGeo(geoCord, 7)),
			8:  int64(h3.FromGeo(geoCord, 8)),
			9:  int64(h3.FromGeo(geoCord, 9)),
			10: int64(h3.FromGeo(geoCord, 10)),
			11: int64(h3.FromGeo(geoCord, 11)),
			12: int64(h3.FromGeo(geoCord, 12)),
			13: int64(h3.FromGeo(geoCord, 13)),
			14: int64(h3.FromGeo(geoCord, 14)),
			15: int64(h3.FromGeo(geoCord, 15)),
		},
	}
}

func DeleteUser(id int) (*User, error) {
	panic("Implement me!")
}

func UpdateUser(id int, lat float64, long float64) (*User, error) {
	panic("Implement me!")
}

func AddUser(user *User) (*User, error) {
	return user, nil
}

func GetCloseUsers(lat float64, long float64, resolution int, category Category) ([]*User, error) {
	panic("Implement me!")
}

func GetAllUsers(category Category) ([]*User, error) {
	return []*User{NewUser("", 1, 2, Client)}, nil
}

func GetUser(id int) (*User, error) {
	panic("Implement me!")
}
