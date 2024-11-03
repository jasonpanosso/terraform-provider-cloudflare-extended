resource "cloudflare-extended_r2_event_notification" "%[1]s" {
  account_id = "%[2]s"
  bucket_name = "%[3]s"
  queue_id     = "%[4]s"

  rules = [
    {
      actions = ["PutObject"],
      prefix = ".jpeg",
      suffix = ".png",
    },
  ]
}
