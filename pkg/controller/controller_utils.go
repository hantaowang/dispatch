package controller

import "fmt"

func NameFunc(owner, namespace string) string {
	return fmt.Sprintf("%s-%s", owner, namespace)
}
