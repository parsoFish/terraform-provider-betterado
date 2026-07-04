# Look up an existing project to scope the subscription.
data "betterado_project" "example" {
  name = "my-project"
}

# Look up the subscriber identity.
data "betterado_identity_user" "subscriber" {
  mail = "team@example.com"
}

# Create an ADO notification subscription that sends an email
# when a work item is changed in the project.
resource "betterado_notification_subscription" "example" {
  project_id        = data.betterado_project.example.id
  subscription_type = "ms.vss-work.workitem-changed-event"
  subscriber_id     = data.betterado_identity_user.subscriber.id
  channel_type      = "EmailHtml"
  channel_address   = "team@example.com"

  # Optional: override filter type (defaults to "Expression")
  filter_type = "Expression"
}
