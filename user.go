package main

import "github.com/uber/h3-go"

const (
	Generic         = "GENERIC"
	ServiceProvider = "SERVICE_PROVIDER"
	Client          = "CLIENT"
)

type Category string

type User struct {
	Id          uint16      `json:"id"`
	Name        string      `json:"name"`
	Category    Category    `json:"category"`
	GeoCord     h3.GeoCoord `json:"geo_cord"`
	H3Positions map[int]h3.H3Index
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
		H3Positions: map[int]h3.H3Index{
			0:  h3.FromGeo(geoCord, 0),
			1:  h3.FromGeo(geoCord, 1),
			2:  h3.FromGeo(geoCord, 2),
			3:  h3.FromGeo(geoCord, 3),
			4:  h3.FromGeo(geoCord, 4),
			5:  h3.FromGeo(geoCord, 5),
			6:  h3.FromGeo(geoCord, 6),
			7:  h3.FromGeo(geoCord, 7),
			8:  h3.FromGeo(geoCord, 8),
			9:  h3.FromGeo(geoCord, 9),
			10: h3.FromGeo(geoCord, 10),
			11: h3.FromGeo(geoCord, 11),
			12: h3.FromGeo(geoCord, 12),
			13: h3.FromGeo(geoCord, 13),
			14: h3.FromGeo(geoCord, 14),
			15: h3.FromGeo(geoCord, 15),
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
