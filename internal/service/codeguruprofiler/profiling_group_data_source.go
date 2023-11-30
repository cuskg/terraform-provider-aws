// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeguruprofiler

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/codeguruprofiler/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Profiling Group")
func newDataSourceProfilingGroup(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceProfilingGroup{}, nil
}

const (
	DSNameProfilingGroup = "Profiling Group Data Source"
)

type dataSourceProfilingGroup struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceProfilingGroup) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_codeguruprofiler_profiling_group"
}

func (d *dataSourceProfilingGroup) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	computePlatform := fwtypes.StringEnumType[awstypes.ComputePlatform]()

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"agent_orchestration_config": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[dsAgentOrchestrationConfig](ctx),
				Computed:    true,
				ElementType: fwtypes.NewObjectTypeOf[dsAgentOrchestrationConfig](ctx),
			},
			"arn": framework.ARNAttributeComputedOnly(),
			"compute_platform": schema.StringAttribute{
				CustomType: computePlatform,
				Computed:   true,
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
			},
			"profiling_status": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[dsProfilingStatus](ctx),
				Computed:    true,
				ElementType: fwtypes.NewObjectTypeOf[dsProfilingStatus](ctx),
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}
func (d *dataSourceProfilingGroup) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().CodeGuruProfilerClient(ctx)

	var data dataSourceProfilingGroupData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findProfilingGroupByName(ctx, conn, data.Name.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CodeGuruProfiler, create.ErrActionSetting, DSNameProfilingGroup, data.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = flex.StringToFramework(ctx, out.Name)
	data.Tags = flex.FlattenFrameworkStringValueMap(ctx, out.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceProfilingGroupData struct {
	ARN                      types.String                                                `tfsdk:"arn"`
	AgentOrchestrationConfig fwtypes.ListNestedObjectValueOf[dsAgentOrchestrationConfig] `tfsdk:"agent_orchestration_config"`
	ComputePlatform          fwtypes.StringEnum[awstypes.ComputePlatform]                `tfsdk:"compute_platform"`
	ID                       types.String                                                `tfsdk:"id"`
	Name                     types.String                                                `tfsdk:"name"`
	ProfilingStatus          fwtypes.ListNestedObjectValueOf[dsProfilingStatus]          `tfsdk:"profiling_status"`
	Tags                     types.Map                                                   `tfsdk:"tags"`
}

type dsAgentOrchestrationConfig struct {
	ProfilingEnabled types.Bool `tfsdk:"profiling_enabled"`
}

type dsProfilingStatus struct {
	LatestAgentOrchestratedAt    types.String                                             `tfsdk:"latest_agent_orchestrated_at"`
	LatestAgentProfileReportedAt types.String                                             `tfsdk:"latest_agent_profile_reported_at"`
	LatestAggregatedProfile      fwtypes.ListNestedObjectValueOf[dsAggregatedProfileTime] `tfsdk:"latest_aggregated_profile"`
}

type dsAggregatedProfileTime struct {
	Period types.String `tfsdk:"period"`
	Start  types.String `tfsdk:"start"`
}
