package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"x-revalidate-bot/pkg/xserver"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var Verbose bool

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose logging")
}

func main() {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	if err := rootCmd.Execute(); err != nil {
		slog.Error("Error executing command", "error", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "updater",
	Short: "Updater for XServer Free VPS Expiration",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		level := slog.LevelInfo
		if Verbose {
			level = slog.LevelDebug
		}
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		})))
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := godotenv.Load(); err != nil {
			slog.Error("Error loading .env file", "error", err)
			os.Exit(1)
		}

		if err := runInternally(); err != nil {
			os.Exit(1)
		}
	},
}

type UserAgentHeaders struct {
	Chrome map[string]string `json:"chrome"`
}

func parseHeaderFile(reader io.Reader) (map[string]string, error) {
	var headers map[string]string
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&headers); err != nil {
		return nil, err
	}
	return headers, nil
}

func getHeaders() (map[string]string, error) {
	headerPath := filepath.Join("assets", "headers", "chrome-macos.json")

	if _, err := os.Stat(headerPath); os.IsNotExist(err) {
		headerPath = filepath.Join("..", "..", "assets", "headers", "chrome-macos.json")
	}

	file, err := os.Open(headerPath)
	if err != nil {
		return nil, fmt.Errorf("error opening header file: %w", err)
	}
	defer file.Close()

	headers, err := parseHeaderFile(file)
	if err != nil {
		return nil, err
	}

	headers["host"] = ""
	headers["connection"] = ""
	headers["accept-encoding"] = ""
	headers["accept-language"] = "ja"

	return headers, nil
}

func runInternally() error {
	vpsID := os.Getenv("VPS_ID")
	x2sessid := os.Getenv("X2SESSID")
	deviceKey := os.Getenv("XSERVER_DEVICEKEY")
	if vpsID == "" || x2sessid == "" || deviceKey == "" {
		slog.Error("VPS_ID, X2SESSID, and XSERVER_DEVICEKEY environment variables are required")
		return fmt.Errorf("missing required environment variables")
	}
	slog.Info("Starting VPS renewal process", "vps_id", vpsID)
	slog.Debug("Credentials loaded", "x2sessid", maskCredential(x2sessid), "device_key", maskCredential(deviceKey))

	headers, err := getHeaders()
	if err != nil {
		slog.Error("Error getting headers", "error", err)
		return err
	}

	xs, err := xserver.NewClient(xserver.ClientOptions{
		SessionID: x2sessid,
		DeviceKey: deviceKey,
		Headers:   headers,
		Logger:    slog.Default(),
	})
	if err != nil {
		slog.Error("Error creating XServer client", "error", err)
		return err
	}
	ctx := context.Background()

	uniqueID, err := xs.GetCSRFTokenAsUniqueID(ctx, xserver.VPSID(vpsID))
	if err != nil {
		slog.Error("Error getting unique ID", "error", err, "vps_id", vpsID)
		return err
	}
	slog.Info("Unique ID retrieved", "unique_id", uniqueID)

	if err := xs.ExtendFreeVPSExpiration(ctx, xserver.VPSID(vpsID), uniqueID); err != nil {
		slog.Error("Error extending free VPS", "error", err, "vps_id", vpsID, "unique_id", uniqueID)
		return err
	}

	return nil
}
