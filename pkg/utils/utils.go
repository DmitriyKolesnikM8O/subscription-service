package utils

import "time"

func ConnectWithTries(fn func() error, attempts int, delay time.Duration) error {
	var lastErr error
	for attempts > 0 {
		if err := fn(); err != nil {
			lastErr = err
			time.Sleep(delay)
			attempts--
			continue
		}
		return nil
	}
	return lastErr
}
