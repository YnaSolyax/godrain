package main

import (
	"github.com/YnaSolyax/godrain/logparser"
)

func main() {
	//cmd.Execute()
	logparser.GetLogFields("BGL.log", "[Label] [Timestamp] [Date] [Node] [Time] [NodeRepeat] [Type] [Component] [Level] [Content]")
	logparser.GetLogFields("Thunderbird.log", "[Label] [Timestamp] [Date] [Admin] [Month] [Day] [Time] [AdminAddr] [Content]")
}
