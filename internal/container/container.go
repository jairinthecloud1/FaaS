package container

import "fmt"

func Auth(username, password string) error {
	if username != "admin" || password != "admin" {
		return fmt.Errorf("invalid credentials")
	}
	return nil
}
