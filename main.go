package main

import (
	"fmt"
	"os"

	"zup/pkg/setup"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "zup",
		Short: "Automates local repo setup using AI",
	}

	rootCmd.AddCommand(setup.RunCmd)
	rootCmd.AddCommand(setup.SetOpenAIKeyCmd)
	rootCmd.AddCommand(setup.GetOpenAIKeyCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
