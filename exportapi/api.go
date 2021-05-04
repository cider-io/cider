package exportapi

var Tasks map[string]Task

func GetCurrentLoad() int {
	unfinishedTasks := 0
	for _, task := range Tasks {
		if task.Status != Stopped {
			unfinishedTasks++
		}
	}
	return unfinishedTasks
}
