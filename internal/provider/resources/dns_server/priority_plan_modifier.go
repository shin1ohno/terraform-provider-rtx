package dns_server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// autoPriorityPlanModifier computes priority automatically when priority_start is set.
type autoPriorityPlanModifier struct{}

func AutoPriorityModifier() planmodifier.Int64 {
	return autoPriorityPlanModifier{}
}

func (m autoPriorityPlanModifier) Description(ctx context.Context) string {
	return "Automatically computes priority based on priority_start and priority_step when in auto mode."
}

func (m autoPriorityPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m autoPriorityPlanModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// Get the full plan to access priority_start and priority_step
	var plan DNSServerModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	priorityStart := fwhelpers.GetInt64Value(plan.PriorityStart)
	if priorityStart == 0 {
		// Manual mode - don't modify
		return
	}

	priorityStep := fwhelpers.GetInt64Value(plan.PriorityStep)
	if priorityStep == 0 {
		priorityStep = DefaultPriorityStep
	}

	// Calculate the index of this server_select entry from the path
	// Path looks like: server_select[0].priority
	pathSteps := req.Path.Steps()
	if len(pathSteps) < 2 {
		return
	}

	// Find the index from the path
	// The path structure is: server_select -> [index] -> priority
	for i, step := range pathSteps {
		if stepStr := step.String(); stepStr == "server_select" && i+1 < len(pathSteps) {
			// The next step should be the index
			indexStep := pathSteps[i+1]
			// ElementKeyInt returns the index
			if keyInt, ok := indexStep.(interface{ ElementKeyInt() int }); ok {
				index := keyInt.ElementKeyInt()
				calculatedPriority := priorityStart + (index * priorityStep)
				resp.PlanValue = types.Int64Value(int64(calculatedPriority))
				return
			}
		}
	}
}
