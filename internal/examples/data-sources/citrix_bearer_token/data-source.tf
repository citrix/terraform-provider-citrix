# Bearer token can be fetched with the citrix_bearer_token data source.
data "citrix_bearer_token" "example_bearer_token" {}

# The bearer token can then be output with terraform. `sensitive = true` is required for outputing sensitive data.
output "example_bearer_token_output" {
    value = data.citrix_bearer_token.example_bearer_token.bearer_token
    sensitive = true
}
