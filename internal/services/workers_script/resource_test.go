package workers_script_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	cfv1 "github.com/cloudflare/cloudflare-go"
	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/workers"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/acctest"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/utils"
)

const (
	moduleContent1 = `export default { fetch() { return new Response('Hello world 1'); }, };`
	moduleContent2 = `export default { fetch() { return new Response('Hello world 2'); }, };`
)

func TestAccCloudflareWorkerScript_ScriptEnt(t *testing.T) {
	t.Parallel()

	rnd := utils.GenerateRandomResourceName()
	name := "cloudflare-extended_workers_script." + rnd
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	bucketName := os.Getenv("R2_BUCKET_NAME")

	if bucketName == "" {
		t.Fatal("R2_BUCKET_NAME must be set for this acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.TestAccPreCheck(t)
			acctest.TestAccPreCheck_AccountID(t)
		},
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCloudflareWorkerScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareWorkerScriptConfigScriptInitial(rnd, accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWorkerScriptExists(name, nil),
					resource.TestCheckResourceAttr(name, "script_name", rnd),
				),
			},
			{
				Config: testAccCheckCloudflareWorkerScriptConfigScriptUpdate(rnd, accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWorkerScriptExists(name, nil),
					resource.TestCheckResourceAttr(name, "script_name", rnd),
				),
			},
			{
				Config: testAccCheckCloudflareWorkerScriptConfigScriptUpdateBinding(rnd, accountID, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareWorkerScriptExists(name, []string{"ai", "bucket"}),
					resource.TestCheckResourceAttr(name, "script_name", rnd),
				),
			},
		},
	})
}

func testAccCheckCloudflareWorkerScriptConfigScriptInitial(rnd, accountID string) string {
	return acctest.LoadTestCase("workerscriptconfigscriptinitial.tf", rnd, accountID, moduleContent1)
}

func testAccCheckCloudflareWorkerScriptConfigScriptUpdate(rnd, accountID string) string {
	return acctest.LoadTestCase("workerscriptconfigscriptupdate.tf", rnd, accountID, moduleContent2)
}

func testAccCheckCloudflareWorkerScriptConfigScriptUpdateBinding(rnd, accountID, bucketName string) string {
	return acctest.LoadTestCase("workerscriptconfigscriptupdatebinding.tf", rnd, accountID, moduleContent2, bucketName)
}

func testAccCheckCloudflareWorkerScriptExists(n string, bindings []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Worker Script ID is set")
		}
		client := cloudflare.NewClient()

		r, err := client.Workers.Scripts.Settings.Get(context.Background(), rs.Primary.ID, workers.ScriptSettingGetParams{AccountID: cloudflare.F(accountID)})
		if err != nil {
			return err
		}

		if r == nil {
			return fmt.Errorf("Worker Script not found")
		}

		foundBindings, err := getWorkerScriptBindings(context.Background(), accountID, rs.Primary.ID, nil)
		if err != nil {
			return fmt.Errorf("cannot list script bindings: %w", err)
		}

		for _, binding := range bindings {
			if _, ok := foundBindings[binding]; !ok {
				return fmt.Errorf("cannot find binding with name %s", binding)
			}
		}

		return nil
	}
}

func testAccCheckCloudflareWorkerScriptDestroy(s *terraform.State) error {
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare-extended_workers_script" {
			continue
		}

		client, err := cfv1.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN"))
		if err != nil {
			tflog.Error(context.TODO(), fmt.Sprintf("failed to create Cloudflare client: %s", err))
		}

		r, _ := client.GetWorker(context.Background(), cfv1.AccountIdentifier(accountID), rs.Primary.ID)

		if r.Script != "" {
			return fmt.Errorf("worker script with id %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

type ScriptBindings map[string]cfv1.WorkerBinding

func getWorkerScriptBindings(ctx context.Context, accountId, scriptName string, dispatchNamespace *string) (ScriptBindings, error) {
	client, err := cfv1.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN"))
	if err != nil {
		return nil, err
	}

	resp, err := client.ListWorkerBindings(
		ctx,
		cfv1.AccountIdentifier(accountId),
		cfv1.ListWorkerBindingsParams{ScriptName: scriptName, DispatchNamespace: dispatchNamespace},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list script bindings: %w", err)
	}

	bindings := make(ScriptBindings, len(resp.BindingList))

	for _, b := range resp.BindingList {
		bindings[b.Name] = b.Binding
	}

	return bindings, nil
}
