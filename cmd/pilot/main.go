package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kawaburger/pilot/internal/api"
	"github.com/kawaburger/pilot/internal/auth"
	"github.com/kawaburger/pilot/internal/claude"
	"github.com/kawaburger/pilot/internal/config"
	"github.com/kawaburger/pilot/internal/tmux"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func main() {
	root := &cobra.Command{
		Use:   "pilot",
		Short: "CLI tool for remote-controlling Claude Code",
	}

	root.AddCommand(serveCmd())
	root.AddCommand(adminCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func serveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the Pilot server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			commands, err := config.LoadCommands()
			if err != nil {
				return fmt.Errorf("loading commands: %w", err)
			}

			pilotHome := config.PilotHome()

			dbPath := filepath.Join(pilotHome, "pilot.db")
			store, err := auth.NewStore(dbPath)
			if err != nil {
				return fmt.Errorf("opening auth store: %w", err)
			}
			defer store.Close()

			jwtMgr, err := auth.NewJWTManager(pilotHome)
			if err != nil {
				return fmt.Errorf("creating JWT manager: %w", err)
			}

			tmux.CleanupStaleSessions()

			s := &api.Server{
				Config:    cfg,
				Commands:  commands,
				Store:     store,
				JWT:       jwtMgr,
				Watchers:  claude.NewWatcherManager(),
				StartedAt: time.Now(),
			}

			r := api.NewRouter(s)

			addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
			fmt.Printf("Pilot server listening on %s\n", addr)
			return r.Run(addr)
		},
	}
}

func adminCmd() *cobra.Command {
	admin := &cobra.Command{
		Use:   "admin",
		Short: "Admin operations",
	}

	create := &cobra.Command{
		Use:   "create <username>",
		Short: "Create a new user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			username := args[0]

			password, _ := cmd.Flags().GetString("password")
			if password == "" {
				fmt.Print("Password: ")
				passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
				fmt.Println()
				if err != nil {
					return fmt.Errorf("reading password: %w", err)
				}
				password = string(passwordBytes)
			}

			if password == "" {
				return fmt.Errorf("password cannot be empty")
			}

			pilotHome := config.PilotHome()
			dbPath := filepath.Join(pilotHome, "pilot.db")
			store, err := auth.NewStore(dbPath)
			if err != nil {
				return fmt.Errorf("opening auth store: %w", err)
			}
			defer store.Close()

			user, err := store.CreateUser(username, password)
			if err != nil {
				return fmt.Errorf("creating user: %w", err)
			}

			fmt.Printf("User %q created (id=%d)\n", user.Username, user.ID)
			return nil
		},
	}

	create.Flags().StringP("password", "p", "", "Password (if not provided, will prompt)")
	admin.AddCommand(create)
	return admin
}
