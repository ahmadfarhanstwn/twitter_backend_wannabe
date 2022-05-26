package util

import (
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GetRandomString(length int) string {
	alphabet := "abcdefghijklmnopqrstuvwxyz"

	var res strings.Builder

	for i := 0; i < length; i++ {
		n := alphabet[rand.Intn(len(alphabet))]
		res.WriteByte(n)
	}

	return res.String()
}

func GetRandomEmail() string {
	name := GetRandomString(7)
	domain := GetRandomString(5)
	return name + "@" + domain + ".com"
} 