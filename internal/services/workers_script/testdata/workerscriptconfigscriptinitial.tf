resource "cloudflare-extended_workers_script" "%[1]s" {
  account_id  = "%[2]s"
  script_name = "%[1]s"
  main_module = "%[1]s"

  parts = {
    %[1]s = {
      part   = "%[3]s"
      module = true
    }
  }
}
