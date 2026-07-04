# Read an existing notification subscription by its ID.
data "betterado_notification_subscription" "example" {
  id = "00000000-0000-0000-0000-000000000001"
}

# Use the subscription attributes in other resources or outputs.
output "subscription_type" {
  value = data.betterado_notification_subscription.example.subscription_type
}

output "channel_type" {
  value = data.betterado_notification_subscription.example.channel_type
}

output "status" {
  value = data.betterado_notification_subscription.example.status
}
