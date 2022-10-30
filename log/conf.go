package log

var logType string = "console"

func SetLogType(lt string) {
	if lt == "mongo" {
		logType = "mongo"
	} else {
		logType = "console"
	}
}
