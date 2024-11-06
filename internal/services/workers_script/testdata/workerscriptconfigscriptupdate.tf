resource "cloudflare-extended_workers_script" "%[1]s" {
  account_id          = "%[2]s"
  script_name         = "%[1]s"
  main_module         = "%[1]s"
  compatibility_flags = ["nodejs_compat"]
  compatibility_date  = "2024-10-22"

  parts = {
    %[1]s = {
      part   = "%[3]s"
      module = true
    }
  }
}
