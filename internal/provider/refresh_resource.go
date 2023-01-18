package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/change-engine/terraform-provider-slack-token/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &refreshResource{}
	_ resource.ResourceWithImportState = &refreshResource{}
)

func NewRefreshResource() resource.Resource {
	return &refreshResource{}
}

type refreshResource struct{}

func (r *refreshResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_refresh"
}

func (r *refreshResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Slack Token Refresh resource",

		Attributes: map[string]schema.Attribute{
			"expires": schema.Int64Attribute{
				MarkdownDescription: "Next refresh time.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					timeoutUnkownModifier{},
				},
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Current API token.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					timeoutUnkownModifier{},
				},
			},
			"refresh_token": schema.StringAttribute{
				MarkdownDescription: "Current refresh token.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					timeoutUnkownModifier{},
				},
			},
		},
	}
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

func (r refreshResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Refresh Error", "Refresh Token cannot be created only imported.")
	return
}

func (r refreshResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Not posible
}

func (r refreshResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	err := client.Request(ctx, "tooling.tokens.rotate?refresh_token="+state.RefreshToken.ValueString(), resultJson)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rotate token, got error: %s", err))
		return
	}

	data.Token = types.StringValue(resultJson.Token)
	data.Expires = types.Int64Value(resultJson.Expires - 60*60*3)
	data.RefreshToken = types.StringValue(resultJson.RefreshToken)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r refreshResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *refreshResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("refresh_token"), req, resp)
}

type timeoutUnkownModifier struct{}

func (m timeoutUnkownModifier) Description(ctx context.Context) string {
	return "Allow refresh before token expires."
}

func (m timeoutUnkownModifier) MarkdownDescription(ctx context.Context) string {
	return "Allow refresh before token expires."
}

func (m timeoutUnkownModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	var data refreshResourceData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if data.Expires.ValueInt64() < time.Now().Unix() {
		resp.PlanValue = types.Int64Unknown()
	}
}

func (m timeoutUnkownModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	var data refreshResourceData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if data.Expires.ValueInt64() < time.Now().Unix() {
		resp.PlanValue = types.StringUnknown()

	}
}
