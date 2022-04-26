package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/change-engine/terraform-provider-slack-token/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ tfsdk.ResourceType = refreshResourceType{}
var _ tfsdk.Resource = refreshResource{}

type refreshResourceType struct{}

func (t refreshResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Slack Token Refresh resource",

		Attributes: map[string]tfsdk.Attribute{
			"expires": {
				MarkdownDescription: "Next refresh time.",
				Optional:            true,
				Computed:            true,
				Type:                types.Int64Type,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					timeoutUnkownModifier{},
				},
			},
			"token": {
				MarkdownDescription: "Current API token.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					timeoutUnkownModifier{},
				},
			},
			"refresh_token": {
				MarkdownDescription: "Current refresh token.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					timeoutUnkownModifier{},
				},
			},
		},
	}, nil
}

func (t refreshResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return refreshResource{
		provider: provider,
	}, diags
}

type refreshResourceData struct {
	Expires      types.Int64  `tfsdk:"expires"`
	Token        types.String `tfsdk:"token"`
	RefreshToken types.String `tfsdk:"refresh_token"`
}

type refreshResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	Expires      int64  `json:"exp"`
}

type refreshResource struct {
	provider provider
}

func (r refreshResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	resp.Diagnostics.AddError("Refresh Error", "Refresh Token cannot be created only imported.")
	return
}

func (r refreshResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Not posible
}

func (r refreshResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var state refreshResourceData
	var data refreshResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := client.New()
	var resultJson refreshResponse
	err := client.Request(ctx, "tooling.tokens.rotate?refresh_token="+state.RefreshToken.Value, resultJson)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rotate token, got error: %s", err))
		return
	}

	data.Token = types.String{Value: resultJson.Token}
	data.Expires = types.Int64{Value: resultJson.Expires - 60*60*3}
	data.RefreshToken = types.String{Value: resultJson.RefreshToken}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r refreshResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	resp.State.RemoveResource(ctx)
}

func (r refreshResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("refresh_token"), req, resp)
}

type timeoutUnkownModifier struct{}

func (m timeoutUnkownModifier) Description(ctx context.Context) string {
	return "Allow refresh before token expires."
}

func (m timeoutUnkownModifier) MarkdownDescription(ctx context.Context) string {
	return "Allow refresh before token expires."
}

func (m timeoutUnkownModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	var data refreshResourceData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if data.Expires.Value < time.Now().Unix() {
		if resp.AttributePlan.Type(ctx) == types.Int64Type {
			resp.AttributePlan = types.Int64{Unknown: true}
		}
		if resp.AttributePlan.Type(ctx) == types.StringType {
			resp.AttributePlan = types.String{Unknown: true}
		}
	}
}
