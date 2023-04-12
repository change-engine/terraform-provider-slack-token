resource "slack-token_refresh" "example" {
  # This resource _must_ be imported.
  # Generate a new "App Configuration Token" via https://api.slack.com/authentication/config-tokens
  # Copy the "Refresh Token", it should start `xoxe-...`
  # Then run `terraform import slack-token_refresh.example xoxe-...`
}
