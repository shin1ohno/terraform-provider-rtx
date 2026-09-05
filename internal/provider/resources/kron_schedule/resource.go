package kron_schedule

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                   = &KronScheduleResource{}
	_ resource.ResourceWithImportState    = &KronScheduleResource{}
	_ resource.ResourceWithValidateConfig = &KronScheduleResource{}
	_ resource.ResourceWithModifyPlan     = &KronScheduleResource{}
)

// NewKronScheduleResource creates a new kron schedule resource.
func NewKronScheduleResource() resource.Resource {
	return &KronScheduleResource{}
}

// KronScheduleResource defines the resource implementation.
type KronScheduleResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *KronScheduleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kron_schedule"
}

// Schema defines the schema for the resource.
func (r *KronScheduleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a kron schedule (scheduled task) on RTX routers. Schedules can be time-based (daily), startup-based, or date-specific (one-time).",
		Attributes: map[string]schema.Attribute{
			"schedule_id": schema.Int64Attribute{
				Description: "The schedule ID (1-65535). Must be unique across all schedules on the router.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"name": schema.StringAttribute{
				Description: "Optional name/description for the schedule.",
				Optional:    true,
			},
			"at_time": schema.StringAttribute{
				// HH:MM:SS is accepted as well as HH:MM because that is how the
				// router renders a schedule back: `0:00` goes in, `show config`
				// prints `00:00:00`. Rejecting it would make a value read off
				// the device (or produced by `terraform import`) unwritable.
				Description: "Time to execute the schedule in HH:MM or HH:MM:SS format (24-hour). Required unless on_startup is true.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^(\d{1,2}):(\d{2})(:(\d{2}))?$`),
						"must be in HH:MM or HH:MM:SS format (e.g., '12:00', '6:30', '00:00:00')",
					),
					timeFormatValidator{},
					stringvalidator.ConflictsWith(path.MatchRoot("on_startup")),
				},
			},
			"day_of_week": schema.StringAttribute{
				Description: "Day(s) of week to execute the schedule. Examples: 'mon', 'mon-fri', 'sat,sun'. If not specified with at_time, schedule runs daily.",
				Optional:    true,
				Validators: []validator.String{
					dayOfWeekValidator{},
					stringvalidator.ConflictsWith(path.MatchRoot("on_startup"), path.MatchRoot("date")),
				},
			},
			"date": schema.StringAttribute{
				Description: "Specific date for one-time schedule execution in YYYY/MM/DD format. Cannot be combined with on_startup or day_of_week.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^(\d{4})/(\d{2})/(\d{2})$`),
						"must be in YYYY/MM/DD format (e.g., '2025/01/15')",
					),
					dateFormatValidator{},
					stringvalidator.ConflictsWith(path.MatchRoot("on_startup"), path.MatchRoot("day_of_week")),
				},
			},
			"recurring": schema.BoolAttribute{
				Description: "Whether the schedule repeats. Automatically set to false for date-specific schedules.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"on_startup": schema.BoolAttribute{
				Description: "Execute this schedule when the router starts up. Cannot be combined with at_time or date.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Validators: []validator.Bool{
					boolConflictsWithStrings(path.MatchRoot("at_time"), path.MatchRoot("date"), path.MatchRoot("day_of_week")),
				},
			},
			"policy_list": schema.StringAttribute{
				Description: "Name of a kron policy to execute. Use this OR command_lines, not both.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("command_lines")),
				},
			},
			"command_lines": schema.ListAttribute{
				// Not capped by a validator on purpose: a hard error would also
				// fire on `terraform destroy` (config validation runs for every
				// operation), stranding anyone who already has a longer list in
				// state — including the provider's own examples/schedule. The
				// warning in ValidateConfig carries the same message without
				// wedging them.
				Description: "List of commands to execute. Use this OR policy_list, not both. " +
					"An RTX schedule ID holds exactly ONE command: a second `schedule at <id>` line " +
					"replaces the first rather than adding to it, so only the last element survives " +
					"on the device.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.MatchRoot("policy_list")),
				},
			},
		},
	}
}

// ValidateConfig performs custom configuration validation.
func (r *KronScheduleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data KronScheduleModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that either at_time, on_startup, or date is specified
	atTime := fwhelpers.GetStringValue(data.AtTime)
	onStartup := fwhelpers.GetBoolValue(data.OnStartup)
	date := fwhelpers.GetStringValue(data.Date)

	if atTime == "" && !onStartup && date == "" {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"One of 'at_time', 'on_startup', or 'date' must be specified.",
		)
	}

	// Validate that either policy_list or command_lines is specified
	policyList := fwhelpers.GetStringValue(data.PolicyList)
	hasCommands := !data.CommandLines.IsNull() && !data.CommandLines.IsUnknown() && len(data.CommandLines.Elements()) > 0

	if policyList == "" && !hasCommands {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Either 'policy_list' or 'command_lines' must be specified.",
		)
	}

	// `recurring` has a static default of true, so an omitted value plans as
	// true even for a schedule that fires once — and the router can only ever
	// report back what the line's shape says. A contradiction here is a
	// guaranteed post-apply consistency failure, so name it at plan time.
	// ModifyPlan derives the value when config leaves it unset; this only
	// catches an explicit one that disagrees.
	if !data.Recurring.IsNull() && !data.Recurring.IsUnknown() {
		oneShot := onStartup || parsers.IsScheduleOneTimeDate(date)
		switch {
		case data.Recurring.ValueBool() && oneShot:
			resp.Diagnostics.AddAttributeError(
				path.Root("recurring"),
				"Invalid Configuration",
				"recurring cannot be true for an on_startup schedule or a YYYY/MM/DD date — both fire once.",
			)
		case !data.Recurring.ValueBool() && !oneShot:
			resp.Diagnostics.AddAttributeError(
				path.Root("recurring"),
				"Invalid Configuration",
				"recurring cannot be false for a plain at_time schedule — the RTX repeats it every day. "+
					"Add a YYYY/MM/DD date, or on_startup, for a schedule that fires once.",
			)
		}
	}

	if len(data.CommandLines.Elements()) > 1 {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("command_lines"),
			"Only the last command will survive on the router",
			"An RTX schedule ID holds exactly one command: each `schedule at <id>` line replaces "+
				"the previous one rather than adding to it. This resource writes one line per "+
				"element, so the router ends up running only the last, and the read-back will not "+
				"match the configured list. Split the commands across separate schedule IDs.",
		)
	}
}

// ModifyPlan derives `recurring` from the schedule's shape when config leaves
// it unset.
//
// The attribute carries booldefault.StaticBool(true), which is right for the
// common daily schedule and wrong for the other two shapes this resource
// supports: an on_startup schedule and a YYYY/MM/DD date both fire once, and
// the read path correctly reports Recurring=false for them. A planned true that
// Read can never return is a post-apply consistency error on every apply, so
// both of those modes were unusable. Deriving it here rather than dropping the
// default keeps `recurring` a known value in every existing user's plan instead
// of turning it into "(known after apply)".
func (r *KronScheduleResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.Config.Raw.IsNull() {
		return // destroy plan
	}

	var config, plan KronScheduleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || !config.Recurring.IsNull() {
		return
	}

	if fwhelpers.GetBoolValue(plan.OnStartup) || parsers.IsScheduleOneTimeDate(fwhelpers.GetStringValue(plan.Date)) {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("recurring"), types.BoolValue(false))...)
	}
}

// Configure adds the provider configured client to the resource.
func (r *KronScheduleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*fwhelpers.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *fwhelpers.ProviderData, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = providerData.Client
}

// Create creates the resource and sets the initial Terraform state.
func (r *KronScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KronScheduleModel
	var planData KronScheduleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Keep a second, untouched copy of the plan: read() mutates data in place
	// with what the router rendered back, and the Optional non-Computed
	// attributes have to echo the plan exactly (see reconcileWithDesired).
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scheduleID := strconv.FormatInt(data.ScheduleID.ValueInt64(), 10)
	ctx = logging.WithResource(ctx, "rtx_kron_schedule", scheduleID)
	logger := logging.FromContext(ctx)

	schedule := data.ToClient()
	logger.Debug().Str("resource", "rtx_kron_schedule").Msgf("Creating kron schedule: %+v", schedule)

	if err := r.client.CreateSchedule(ctx, schedule); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create kron schedule",
			fmt.Sprintf("Could not create kron schedule: %v", err),
		)
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if missingAfterWrite(&data, &planData, "create", &resp.Diagnostics) {
		return
	}

	data.reconcileWithDesired(&planData)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *KronScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KronScheduleModel
	var priorData KronScheduleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prior state, not the plan, is what a refresh reconciles against —
	// otherwise every refresh rewrites at_time into the router's rendering and
	// the next plan shows a diff that is pure formatting.
	resp.Diagnostics.Append(req.State.Get(ctx, &priorData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if resource was removed
	if data.ScheduleID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	data.reconcileWithDesired(&priorData)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the schedule from the router.
func (r *KronScheduleResource) read(ctx context.Context, data *KronScheduleModel, diagnostics *diag.Diagnostics) {
	scheduleID := int(data.ScheduleID.ValueInt64())

	ctx = logging.WithResource(ctx, "rtx_kron_schedule", strconv.Itoa(scheduleID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_kron_schedule").Msgf("Reading kron schedule: %d", scheduleID)

	schedule, err := r.client.GetSchedule(ctx, scheduleID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Debug().Str("resource", "rtx_kron_schedule").Msgf("Kron schedule %d not found, removing from state", scheduleID)
			data.ScheduleID = types.Int64Null()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read kron schedule", fmt.Sprintf("Could not read kron schedule %d: %v", scheduleID, err))
		return
	}

	data.FromClient(schedule)

	if schedule.Context != "" && schedule.Context != "*" {
		diagnostics.AddWarning(
			"Schedule uses an execution context this resource does not model",
			fmt.Sprintf("Schedule %d runs in the %q context on the router. rtx_kron_schedule only manages "+
				"the global (\"*\") context, so the next apply will rewrite the line without it.",
				scheduleID, schedule.Context),
		)
	}
}

// missingAfterWrite reports the router having no matching line right after a
// write, and returns true when it did.
//
// read() nulls schedule_id when `show config` finds nothing. Straight after a
// create or update that means the write did not land — the RTX rejects a
// malformed command with a message none of the write path's error patterns
// match, so the write can return success having done nothing. Naming it here
// beats handing Terraform a null id and letting core report the opaque
// "Provider produced inconsistent result after apply".
func missingAfterWrite(data, planData *KronScheduleModel, op string, diagnostics *diag.Diagnostics) bool {
	if !data.ScheduleID.IsNull() {
		return false
	}

	id := int(planData.ScheduleID.ValueInt64())
	fwhelpers.AppendDiagError(diagnostics,
		fmt.Sprintf("Kron schedule not found after %s", op),
		fmt.Sprintf("Schedule %d was written to the router but %q returned no matching line. The router "+
			"most likely rejected the command; run it on the console to see the error.",
			id, parsers.BuildShowScheduleByIDCommand(id)),
	)
	return true
}

// missingAfterUpdate is missingAfterWrite for the update path, which deletes
// before it re-creates. When the re-create does not land the router is left
// with NO schedule at that ID while state still holds the old one, so the
// message has to say that rather than implying the previous schedule survived.
func missingAfterUpdate(data, planData *KronScheduleModel, diagnostics *diag.Diagnostics) bool {
	if !data.ScheduleID.IsNull() {
		return false
	}

	id := int(planData.ScheduleID.ValueInt64())
	fwhelpers.AppendDiagError(diagnostics,
		"Kron schedule not found after update",
		fmt.Sprintf("Schedule %d could not be re-created. An update removes the old schedule before "+
			"writing the new one, so the router currently has NO schedule at %d while Terraform state "+
			"still describes the old one. The router most likely rejected the new command; run %q on "+
			"the console to confirm, then re-apply.",
			id, id, parsers.BuildShowScheduleByIDCommand(id)),
	)
	return true
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *KronScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KronScheduleModel
	var planData KronScheduleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scheduleID := strconv.FormatInt(data.ScheduleID.ValueInt64(), 10)
	ctx = logging.WithResource(ctx, "rtx_kron_schedule", scheduleID)
	logger := logging.FromContext(ctx)

	schedule := data.ToClient()
	logger.Debug().Str("resource", "rtx_kron_schedule").Msgf("Updating kron schedule: %+v", schedule)

	if err := r.client.UpdateSchedule(ctx, schedule); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update kron schedule",
			fmt.Sprintf("Could not update kron schedule: %v", err),
		)
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if missingAfterUpdate(&data, &planData, &resp.Diagnostics) {
		return
	}

	data.reconcileWithDesired(&planData)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *KronScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KronScheduleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scheduleID := int(data.ScheduleID.ValueInt64())

	ctx = logging.WithResource(ctx, "rtx_kron_schedule", strconv.Itoa(scheduleID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_kron_schedule").Msgf("Deleting kron schedule: %d", scheduleID)

	if err := r.client.DeleteSchedule(ctx, scheduleID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete kron schedule",
			fmt.Sprintf("Could not delete kron schedule %d: %v", scheduleID, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *KronScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importID := req.ID

	// Parse import ID as schedule ID
	id, err := strconv.ParseInt(importID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Invalid import ID format, expected schedule ID (integer): %v", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("schedule_id"), id)...)
}

// Custom validators

// timeFormatValidator validates that time values are in valid ranges.
type timeFormatValidator struct{}

func (v timeFormatValidator) Description(ctx context.Context) string {
	return "validates time is in valid HH:MM or HH:MM:SS format with valid hour (0-23), minute (0-59) and second (0-59)"
}

func (v timeFormatValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v timeFormatValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		return
	}

	timePattern := regexp.MustCompile(`^(\d{1,2}):(\d{2})(?::(\d{2}))?$`)
	matches := timePattern.FindStringSubmatch(value)
	if len(matches) != 4 {
		return // Already validated by RegexMatches
	}

	hour, _ := strconv.Atoi(matches[1])
	minute, _ := strconv.Atoi(matches[2])

	if matches[3] != "" {
		if second, _ := strconv.Atoi(matches[3]); second < 0 || second > 59 {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Time",
				fmt.Sprintf("Second %d is invalid, must be 0-59", second),
			)
		}
	}

	if hour < 0 || hour > 23 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Time",
			fmt.Sprintf("Hour %d is invalid, must be 0-23", hour),
		)
	}
	if minute < 0 || minute > 59 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Time",
			fmt.Sprintf("Minute %d is invalid, must be 0-59", minute),
		)
	}
}

// dateFormatValidator validates that date values are in valid ranges.
type dateFormatValidator struct{}

func (v dateFormatValidator) Description(ctx context.Context) string {
	return "validates date is in valid YYYY/MM/DD format with valid ranges"
}

func (v dateFormatValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v dateFormatValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		return
	}

	datePattern := regexp.MustCompile(`^(\d{4})/(\d{2})/(\d{2})$`)
	matches := datePattern.FindStringSubmatch(value)
	if len(matches) != 4 {
		return // Already validated by RegexMatches
	}

	year, _ := strconv.Atoi(matches[1])
	month, _ := strconv.Atoi(matches[2])
	day, _ := strconv.Atoi(matches[3])

	if year < 2000 || year > 2099 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Date",
			fmt.Sprintf("Year %d is invalid, must be 2000-2099", year),
		)
	}
	if month < 1 || month > 12 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Date",
			fmt.Sprintf("Month %d is invalid, must be 1-12", month),
		)
	}
	if day < 1 || day > 31 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Date",
			fmt.Sprintf("Day %d is invalid, must be 1-31", day),
		)
	}
}

// dayOfWeekValidator validates day of week specifications.
type dayOfWeekValidator struct{}

func (v dayOfWeekValidator) Description(ctx context.Context) string {
	return "validates day of week is valid (e.g., 'mon', 'mon-fri', 'sat,sun')"
}

func (v dayOfWeekValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v dayOfWeekValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		return
	}

	validDays := map[string]bool{
		"sun": true, "mon": true, "tue": true, "wed": true,
		"thu": true, "fri": true, "sat": true,
	}

	// Handle range format (e.g., "mon-fri")
	if strings.Contains(value, "-") {
		parts := strings.Split(value, "-")
		if len(parts) != 2 {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Day Range",
				fmt.Sprintf("Invalid day range format %q", value),
			)
			return
		}
		if !validDays[strings.ToLower(parts[0])] {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Day",
				fmt.Sprintf("Invalid day %q in range", parts[0]),
			)
		}
		if !validDays[strings.ToLower(parts[1])] {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Day",
				fmt.Sprintf("Invalid day %q in range", parts[1]),
			)
		}
		return
	}

	// Handle comma-separated format (e.g., "mon,wed,fri")
	parts := strings.Split(value, ",")
	for _, part := range parts {
		day := strings.ToLower(strings.TrimSpace(part))
		if !validDays[day] {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Day",
				fmt.Sprintf("Invalid day %q, must be one of: sun, mon, tue, wed, thu, fri, sat", day),
			)
		}
	}
}

// boolConflictsWithStrings creates a validator that conflicts when bool is true and strings are set.
func boolConflictsWithStrings(paths ...path.Expression) boolConflictValidator {
	return boolConflictValidator{conflictPaths: paths}
}

type boolConflictValidator struct {
	conflictPaths []path.Expression
}

func (v boolConflictValidator) Description(ctx context.Context) string {
	return "validates that when true, conflicting string attributes are not set"
}

func (v boolConflictValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v boolConflictValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// Only check conflicts when the value is true
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// ConflictsWith is handled by the string validators on the conflicting attributes
	// This validator is kept for completeness but the actual conflict checking
	// is done by stringvalidator.ConflictsWith on at_time, date, and day_of_week
	_ = req.ConfigValue.ValueBool() // Check value but conflicts handled elsewhere
}
