package cmd

import (
	"context"
	"flag"
	"testing"
)

func TestDevicesRegisterCommand_MissingName(t *testing.T) {
	cmd := DevicesRegisterCommand()

	if err := cmd.FlagSet.Parse([]string{"--udid", "UDID", "--platform", "IOS"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	if err := cmd.Exec(context.Background(), []string{}); err != flag.ErrHelp {
		t.Fatalf("expected flag.ErrHelp when --name is missing, got %v", err)
	}
}

func TestDevicesRegisterCommand_MissingUDID(t *testing.T) {
	cmd := DevicesRegisterCommand()

	if err := cmd.FlagSet.Parse([]string{"--name", "Device", "--platform", "IOS"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	if err := cmd.Exec(context.Background(), []string{}); err != flag.ErrHelp {
		t.Fatalf("expected flag.ErrHelp when --udid is missing, got %v", err)
	}
}

func TestDevicesRegisterCommand_MissingPlatform(t *testing.T) {
	cmd := DevicesRegisterCommand()

	if err := cmd.FlagSet.Parse([]string{"--name", "Device", "--udid", "UDID"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	if err := cmd.Exec(context.Background(), []string{}); err != flag.ErrHelp {
		t.Fatalf("expected flag.ErrHelp when --platform is missing, got %v", err)
	}
}

func TestDevicesRegisterCommand_UDIDConflict(t *testing.T) {
	cmd := DevicesRegisterCommand()

	if err := cmd.FlagSet.Parse([]string{"--name", "Device", "--udid", "UDID", "--udid-from-system", "--platform", "MAC_OS"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	if err := cmd.Exec(context.Background(), []string{}); err != flag.ErrHelp {
		t.Fatalf("expected flag.ErrHelp when udid flags conflict, got %v", err)
	}
}

func TestDevicesUpdateCommand_MissingID(t *testing.T) {
	cmd := DevicesUpdateCommand()

	if err := cmd.FlagSet.Parse([]string{"--status", "ENABLED"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	if err := cmd.Exec(context.Background(), []string{}); err != flag.ErrHelp {
		t.Fatalf("expected flag.ErrHelp when --id is missing, got %v", err)
	}
}

func TestDevicesUpdateCommand_MissingStatus(t *testing.T) {
	cmd := DevicesUpdateCommand()

	if err := cmd.FlagSet.Parse([]string{"--id", "DEVICE_ID"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	if err := cmd.Exec(context.Background(), []string{}); err != flag.ErrHelp {
		t.Fatalf("expected flag.ErrHelp when --status is missing, got %v", err)
	}
}
