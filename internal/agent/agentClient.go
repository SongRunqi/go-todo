package agent

import (
	"fmt"
	"os"

	"github.com/SongRunqi/go-todo/app"
)

var agentContext app.AgentContext

func setSystemPrompts() {
	file, err := os.ReadFile("../../agentcmd/SYSTEM.md")
	if err != nil {
		panic(err)
		fmt.Printf("%v\n", err)
	}
	s := string(file)
	msg := app.Msg{Role: "System", Content: s}
	agentContext.InteractionHistory = append(agentContext.InteractionHistory, msg)

}

func StartUp() {

}

func oneStep() {

}

func init() {
	setSystemPrompts()
}
