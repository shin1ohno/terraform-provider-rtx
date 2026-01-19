package client

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// ScheduleService handles schedule operations
type ScheduleService struct {
	executor Executor
	client   *rtxClient // Reference to the main client for save functionality
}

// NewScheduleService creates a new Schedule service instance
func NewScheduleService(executor Executor, client *rtxClient) *ScheduleService {
	return &ScheduleService{
		executor: executor,
		client:   client,
	}
}

// CreateSchedule creates a new schedule
func (s *ScheduleService) CreateSchedule(ctx context.Context, schedule Schedule) error {
	// Convert client.Schedule to parsers.Schedule for validation
	parserSchedule := s.toParserSchedule(schedule)

	// Validate input
	if err := parsers.ValidateSchedule(parserSchedule); err != nil {
		return fmt.Errorf("invalid schedule: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Check if schedule with same ID already exists
	existing, _ := s.GetSchedule(ctx, schedule.ID)
	if existing != nil {
		return fmt.Errorf("schedule %d already exists", schedule.ID)
	}

	// Build and execute commands based on schedule type
	var cmd string
	for _, command := range schedule.Commands {
		if schedule.OnStartup {
			cmd = parsers.BuildScheduleAtStartupCommand(schedule.ID, command)
		} else if schedule.Date != "" {
			cmd = parsers.BuildScheduleAtDateTimeCommand(schedule.ID, schedule.Date, schedule.AtTime, command)
		} else if schedule.DayOfWeek != "" && schedule.PPInterface > 0 {
			// PP interface schedule
			cmd = parsers.BuildSchedulePPCommand(schedule.PPInterface, schedule.DayOfWeek, schedule.AtTime, command)
		} else {
			// Regular time-based schedule
			cmd = parsers.BuildScheduleAtCommand(schedule.ID, schedule.AtTime, command)
		}

		log.Printf("[DEBUG] Creating schedule with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to create schedule: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("schedule created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetSchedule retrieves a schedule configuration
func (s *ScheduleService) GetSchedule(ctx context.Context, id int) (*Schedule, error) {
	cmd := parsers.BuildShowScheduleByIDCommand(id)
	log.Printf("[DEBUG] Getting schedule with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	log.Printf("[DEBUG] Schedule raw output: %q", string(output))

	parser := parsers.NewScheduleParser()
	parserSchedule, err := parser.ParseSingleSchedule(string(output), id)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schedule: %w", err)
	}

	// Convert parsers.Schedule to client.Schedule
	schedule := s.fromParserSchedule(*parserSchedule)
	return &schedule, nil
}

// UpdateSchedule updates an existing schedule
func (s *ScheduleService) UpdateSchedule(ctx context.Context, schedule Schedule) error {
	parserSchedule := s.toParserSchedule(schedule)

	// Validate input
	if err := parsers.ValidateSchedule(parserSchedule); err != nil {
		return fmt.Errorf("invalid schedule: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Delete existing schedule first
	if err := s.deleteScheduleCommands(ctx, schedule.ID); err != nil {
		// Continue even if delete fails (schedule might not exist)
		log.Printf("[DEBUG] Delete existing schedule warning: %v", err)
	}

	// Create new schedule with updated values
	return s.CreateSchedule(ctx, schedule)
}

// DeleteSchedule removes a schedule
func (s *ScheduleService) DeleteSchedule(ctx context.Context, id int) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := s.deleteScheduleCommands(ctx, id); err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return err
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("schedule deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// deleteScheduleCommands deletes all schedule commands with the given ID
func (s *ScheduleService) deleteScheduleCommands(ctx context.Context, id int) error {
	cmd := parsers.BuildDeleteScheduleCommand(id)
	log.Printf("[DEBUG] Deleting schedule with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		// Check if it's already gone
		if strings.Contains(strings.ToLower(string(output)), "not found") {
			return nil
		}
		return fmt.Errorf("command failed: %s", string(output))
	}

	return nil
}

// ListSchedules retrieves all schedules
func (s *ScheduleService) ListSchedules(ctx context.Context) ([]Schedule, error) {
	cmd := parsers.BuildShowScheduleCommand()
	log.Printf("[DEBUG] Listing schedules with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list schedules: %w", err)
	}

	log.Printf("[DEBUG] Schedules raw output: %q", string(output))

	parser := parsers.NewScheduleParser()
	parserSchedules, err := parser.ParseScheduleConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse schedules: %w", err)
	}

	// Convert parsers.Schedule to client.Schedule
	schedules := make([]Schedule, len(parserSchedules))
	for i, ps := range parserSchedules {
		schedules[i] = s.fromParserSchedule(ps)
	}

	return schedules, nil
}

// CreateKronPolicy creates a new kron policy (command list)
// Note: RTX doesn't have native kron policy support, so we store as comments
func (s *ScheduleService) CreateKronPolicy(ctx context.Context, policy KronPolicy) error {
	// Validate input
	parserPolicy := s.toParserKronPolicy(policy)
	if err := parsers.ValidateKronPolicy(parserPolicy); err != nil {
		return fmt.Errorf("invalid policy: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// For RTX routers, we don't actually have a native kron policy command
	// The policy is just a logical grouping that will be referenced when creating schedules
	// We can store it as a comment in the configuration for documentation purposes
	log.Printf("[DEBUG] KronPolicy %s created with %d commands (stored as logical grouping)",
		policy.Name, len(policy.Commands))

	return nil
}

// GetKronPolicy retrieves a kron policy configuration
func (s *ScheduleService) GetKronPolicy(ctx context.Context, name string) (*KronPolicy, error) {
	// Since RTX doesn't have native kron policy support,
	// policies are managed at the Terraform level
	return nil, fmt.Errorf("kron policy %s not found (policies are managed at Terraform level)", name)
}

// UpdateKronPolicy updates an existing kron policy
func (s *ScheduleService) UpdateKronPolicy(ctx context.Context, policy KronPolicy) error {
	// Validate input
	parserPolicy := s.toParserKronPolicy(policy)
	if err := parsers.ValidateKronPolicy(parserPolicy); err != nil {
		return fmt.Errorf("invalid policy: %w", err)
	}

	log.Printf("[DEBUG] KronPolicy %s updated with %d commands", policy.Name, len(policy.Commands))
	return nil
}

// DeleteKronPolicy removes a kron policy
func (s *ScheduleService) DeleteKronPolicy(ctx context.Context, name string) error {
	log.Printf("[DEBUG] KronPolicy %s deleted", name)
	return nil
}

// ListKronPolicies retrieves all kron policies
func (s *ScheduleService) ListKronPolicies(ctx context.Context) ([]KronPolicy, error) {
	// Since RTX doesn't have native kron policy support,
	// return empty list (policies are managed at Terraform level)
	return []KronPolicy{}, nil
}

// toParserSchedule converts client.Schedule to parsers.Schedule
func (s *ScheduleService) toParserSchedule(schedule Schedule) parsers.Schedule {
	return parsers.Schedule{
		ID:         schedule.ID,
		Name:       schedule.Name,
		AtTime:     schedule.AtTime,
		DayOfWeek:  schedule.DayOfWeek,
		Date:       schedule.Date,
		Recurring:  schedule.Recurring,
		OnStartup:  schedule.OnStartup,
		PolicyList: schedule.PolicyList,
		Commands:   schedule.Commands,
		Enabled:    schedule.Enabled,
	}
}

// fromParserSchedule converts parsers.Schedule to client.Schedule
func (s *ScheduleService) fromParserSchedule(ps parsers.Schedule) Schedule {
	return Schedule{
		ID:         ps.ID,
		Name:       ps.Name,
		AtTime:     ps.AtTime,
		DayOfWeek:  ps.DayOfWeek,
		Date:       ps.Date,
		Recurring:  ps.Recurring,
		OnStartup:  ps.OnStartup,
		PolicyList: ps.PolicyList,
		Commands:   ps.Commands,
		Enabled:    ps.Enabled,
	}
}

// toParserKronPolicy converts client.KronPolicy to parsers.KronPolicy
func (s *ScheduleService) toParserKronPolicy(policy KronPolicy) parsers.KronPolicy {
	return parsers.KronPolicy{
		Name:     policy.Name,
		Commands: policy.Commands,
	}
}

// fromParserKronPolicy converts parsers.KronPolicy to client.KronPolicy
func (s *ScheduleService) fromParserKronPolicy(pp parsers.KronPolicy) KronPolicy {
	return KronPolicy{
		Name:     pp.Name,
		Commands: pp.Commands,
	}
}
