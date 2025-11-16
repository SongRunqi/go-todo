package app

// GetAllCommands returns a human-friendly summary of available CLI commands.
// NOTE: this is a simplified placeholder so the agent can respond without
// depending on the cmd package (which would create a cycle).
func GetAllCommands() string {
	return `Available commands:
- todo ask <text>          : use natural language to manage tasks
- todo list                : show active tasks
- todo get <id>            : show details of a task
- todo complete <id>       : mark a task completed
- todo delete <id>         : remove a task
- todo update <id>         : update task fields
- todo init                : initialize configuration
- todo back [get|restore]  : manage backup tasks
- todo lang [list|set|current] : manage language settings
- todo version             : show version info`
}
