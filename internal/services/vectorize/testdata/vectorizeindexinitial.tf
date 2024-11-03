resource "cloudflare-extended_vectorize_index" "%[1]s" {
  name       = "%[1]s"
  account_id = "%[2]s"
  dimensions = "%[3]d"
  metric     = "%[4]s"

  metadata_indexes = {
    test = "string",
    test2 = "boolean",
    test3 = "number",
  }
}
