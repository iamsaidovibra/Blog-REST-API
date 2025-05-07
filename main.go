package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hashed := "$2a$10$l6LbP4lIldmcbFxjfCfvJeAZzS87LmUXpeIjmWnVG4M2Ss6we095i"
	password := "tomboyxDxDxD"

	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	if err != nil {
		fmt.Println("❌ Passwords DO NOT match:", err)
	} else {
		fmt.Println("✅ Passwords match!")
	}
}
