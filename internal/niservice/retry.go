package niservice

import "time"

func retry(tries int, sleep time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	result, err := fn()
	if err != nil {
		if tries--; tries > 0 {
			time.Sleep(sleep)
			return retry(tries, sleep, fn)
		}
	}
	return result, err
}
