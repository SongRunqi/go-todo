package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/SongRunqi/go-todo/app"
)

// reminderCmd represents the reminder command
var reminderCmd = &cobra.Command{
	Use:   "reminder",
	Short: "管理任务提醒",
	Long:  `管理任务的系统通知提醒，包括设置提醒时间、启用和禁用提醒。`,
}

// reminderSetCmd sets reminder times for a task
var reminderSetCmd = &cobra.Command{
	Use:   "set <id> <duration1> [duration2 ...]",
	Short: "为任务设置提醒时间",
	Long: `为指定任务设置一个或多个提醒时间。

Duration 格式:
  - 分钟: 30m, 15m
  - 小时: 1h, 2h
  - 天: 1d, 2d
  - 组合: 1d12h, 2h30m

示例:
  todo reminder set 1 1h          # 提前1小时提醒
  todo reminder set 1 1h 30m      # 提前1小时和30分钟提醒
  todo reminder set 1 1d 1h 15m   # 提前1天、1小时和15分钟提醒`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := &app.Context{
			Store:       store,
			Todos:       todos,
			Args:        append([]string{"reminder"}, "set "+joinArgs(args)),
			CurrentTime: currentTime,
			Config:      &config,
		}

		reminderCmd := &app.ReminderSetCommand{}
		if err := reminderCmd.Execute(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}
	},
}

// reminderEnableCmd enables reminders for a task
var reminderEnableCmd = &cobra.Command{
	Use:   "enable <id>",
	Short: "启用任务的提醒",
	Long:  `启用指定任务的提醒。任务必须已设置提醒时间。`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := &app.Context{
			Store:       store,
			Todos:       todos,
			Args:        append([]string{"reminder"}, "enable "+args[0]),
			CurrentTime: currentTime,
			Config:      &config,
		}

		reminderCmd := &app.ReminderEnableCommand{}
		if err := reminderCmd.Execute(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}
	},
}

// reminderDisableCmd disables reminders for a task
var reminderDisableCmd = &cobra.Command{
	Use:   "disable <id>",
	Short: "禁用任务的提醒",
	Long:  `禁用指定任务的提醒。提醒配置会保留，可以之后重新启用。`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := &app.Context{
			Store:       store,
			Todos:       todos,
			Args:        append([]string{"reminder"}, "disable "+args[0]),
			CurrentTime: currentTime,
			Config:      &config,
		}

		reminderCmd := &app.ReminderDisableCommand{}
		if err := reminderCmd.Execute(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(reminderCmd)
	reminderCmd.AddCommand(reminderSetCmd)
	reminderCmd.AddCommand(reminderEnableCmd)
	reminderCmd.AddCommand(reminderDisableCmd)
}

// joinArgs joins arguments with spaces
func joinArgs(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		result += arg
	}
	return result
}
