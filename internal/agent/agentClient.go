package agent

import (
	_ "embed"

	"github.com/SongRunqi/go-todo/app"
)

//go:embed SYSTEM.md
var systemPrompt string

var agentContext app.AgentContext

func setSystemPrompts() {
	msg := app.Msg{Role: "System", Content: systemPrompt}
	agentContext.InteractionHistory = append(agentContext.InteractionHistory, msg)
}

func StartUp() {

}

func oneStep() {

}

func init() {
	setSystemPrompts()
}
