package pkg

import (
	"encoding/json"
	"fmt"
)

type LoggingManager interface {
	Log(key string, value interface{})
}

type Logging struct{}

func (l *Logging) Log(key string, value interface{}) {
	if j, err := json.Marshal(map[string]interface{}{"component": key, "content": value}); err == nil {
		fmt.Println(string(j))
	}
}

func NewLogging() LoggingManager {
	return &Logging{}
}
