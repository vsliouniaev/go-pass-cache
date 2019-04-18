package main

import "crypto/rand"

func getEntropy() string {
	b := make([]byte, 256)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	return string(b)
}
