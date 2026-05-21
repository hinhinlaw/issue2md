package greeting

import (
	"fmt"
	_ "time"
)

// GetGreeting generates a greeting message.
// isFormal: whether to use a formal tone.
// isMorning: whether it is morning.
// isEvening: whether it is evening.
func GetGreeting(name string, isFormal bool, isMorning bool, isEvening bool) string {
    if isFormal {
        if isMorning {
            return fmt.Sprintf("Good morning, %s.", name)
        } else {
            if isEvening {
                return fmt.Sprintf("Good evening, %s.", name)
            } else {
                return fmt.Sprintf("Hello, %s.", name)
            }
        }
    } else {
        if isMorning {
            return fmt.Sprintf("Hey %s, good morning!", name)
        } else {
            if isEvening {
                return fmt.Sprintf("Hey %s, good evening!", name)
            } else {
                return fmt.Sprintf("Hey, %s!", name)
            }
        }
    }
}