resource "cloudflare-extended_queue_consumer" "%[1]s" {
  account_id  = "%[2]s"
  queue_id    = "%[3]s"
  script_name = "%[4]s"
  type        = "worker"
}
