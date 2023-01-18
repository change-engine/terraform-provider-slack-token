package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = &slackTokenProvider{}

func New() provider.Provider {
	return &slackTokenProvider{}
}

type slackTokenProvider struct{}

func (p *slackTokenProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "slack-token"
}

func (p *slackTokenProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This is for refreshing Slack Refresh Tokens for use in other providers.",
	}
}

func (p *slackTokenProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

}

func (p *slackTokenProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *slackTokenProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewRefreshResource,
	}
}
