package main

import (
	"github.com/YnaSolyax/godrain/cmd"
)

func main() {
	cmd.Execute()
	//logparser.GetLogFields("[Label] [Timestamp] [Date] [Node] [Time] [NodeRepeat] [Type] [Component] [Level] [Content]")

	//parse := logparser.Log{}
	//fmt.Println(parse.ParseLog("BGL_2k.log", "[Label] [Timestamp] [Date] [Node] [Time] [NodeRepeat] [Type] [Component] [Level] [Content]"))
	//fmt.Println(parse.ParseLog("", "пися пися пися пися [Timestamp] [Content] kakakak"))
	//fmt.Println(parse.ParseLog("Hadoop_2k.log", "[Label] [Timestamp] [Date] [Admin] [Month] [Day] [Time] [AdminAddr] [Content]"))
}
