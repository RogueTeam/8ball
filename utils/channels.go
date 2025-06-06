package utils

import "log"

func ConsumeChannel[T any](c chan T) {
	defer func() {
		err := recover()
		if err == nil {
			return
		}
		log.Println("Failed to consume channel:", err)
	}()
	for range c {
	}
}
