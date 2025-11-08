package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/SongRunqi/go-todo/internal/logger"
	"github.com/SongRunqi/go-todo/internal/notifier"
)

var (
	checkInterval time.Duration
)

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "启动提醒守护进程",
	Long: `启动后台提醒守护进程，持续检查并发送任务提醒。

守护进程会定期检查所有任务，在设定的提醒时间发送系统通知。

示例:
  todo daemon                    # 使用默认检查间隔（1分钟）
  todo daemon --interval 30s     # 每30秒检查一次`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 启动提醒守护进程...")
		fmt.Printf("📊 检查间隔: %s\n", checkInterval)
		fmt.Println("按 Ctrl+C 停止")
		fmt.Println()

		// Create notifier
		systemNotifier := notifier.NewSystemNotifier()

		// Create reminder service
		reminderService := notifier.NewReminderService(store, systemNotifier)

		// Start the service
		err := reminderService.Start(checkInterval)
		if err != nil {
			logger.ErrorWithErr(err, "Failed to start reminder service")
			fmt.Fprintf(os.Stderr, "错误: 无法启动提醒服务: %v\n", err)
			os.Exit(1)
		}

		logger.Info("Reminder service started")
		fmt.Println("✅ 提醒服务已启动")

		// Wait for interrupt signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Block until signal received
		sig := <-sigChan
		fmt.Printf("\n\n收到信号 %v，正在停止...\n", sig)

		// Stop the service
		reminderService.Stop()
		logger.Info("Reminder service stopped")
		fmt.Println("✅ 提醒服务已停止")
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.Flags().DurationVarP(&checkInterval, "interval", "i", 1*time.Minute, "检查提醒的时间间隔")
}
