package store

import "encoding/json"

func encodeSessionQueue(queue []int) (string, error) {
	data, err := json.Marshal(queue)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func decodeSessionQueue(raw string) ([]int, error) {
	var queue []int
	if err := json.Unmarshal([]byte(raw), &queue); err != nil {
		return nil, err
	}
	return queue, nil
}
