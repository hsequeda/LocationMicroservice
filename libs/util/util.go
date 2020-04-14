package util

import "golang.org/x/crypto/bcrypt"

// VerifyPassword
// Example
/*
	if err:=VerifyPassword("test", passwordHash);err!=nil{
		log.Print("Valid Password")
	}else{
		log.Print("Invalid Password")
	}
*/
func VerifyPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

// GeneratePassword
// Example
/*
	passwordHash,err:= GeneratePasswordHash("test")
	if err!=nil{
		log.Fatal(err)
	}
	log.Print(passwordHash)

*/
func GeneratePasswordHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
