package store

import "encoding/json"

type sessionQueueData struct {
	Queue   []int `json:"queue"`
	Current int   `json:"current"`
}

func encodeSessionQueue(queue []int, current int) (string, error) {
	data, err := json.Marshal(sessionQueueData{Queue: queue, Current: current})
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func decodeSessionQueue(raw string) ([]int, int, error) {
	var data sessionQueueData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		var legacy []int
		if err2 := json.Unmarshal([]byte(raw), &legacy); err2 != nil {
			return nil, -1, err
		}
		return legacy, -1, nil
	}
	return data.Queue, data.Current, nil
}
