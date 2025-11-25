package app

// GetAllCommands returns a human-friendly summary of available CLI commands.
// NOTE: this is a simplified placeholder so the agent can respond without
// depending on the cmd package (which would create a cycle).

var c = AgentContext{}

func agentSetUp() {

}

func init() {
	agentSetUp()
}
