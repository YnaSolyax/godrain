package main

import (
	"github.com/YnaSolyax/godrain/cmd"
)

func main() {
	//ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	//defer stop()
	cmd.Execute()
}
