resource "cloudflare-extended_queue_consumer" "%[1]s" {
  account_id  = "%[2]s"
  queue_id    = "%[3]s"
  type        = "http_pull"

  settings = {
    batch_size = 100
  }
}
