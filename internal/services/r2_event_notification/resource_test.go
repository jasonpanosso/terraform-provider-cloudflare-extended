package r2_event_notification_test

import (
	// "context"
	// "fmt"
	"os"
	"testing"

	// "github.com/cloudflare/cloudflare-go/v3"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/acctest"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/utils"
)

func TestAccCloudflareR2EventNotificationInitial_Create(t *testing.T) {
	t.Parallel()

	rnd := utils.GenerateRandomResourceName()
	name := "cloudflare-extended_r2_event_notification." + rnd
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	bucketName := os.Getenv("R2_BUCKET_NAME")
	queueID := os.Getenv("CLOUDFLARE_QUEUE_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.TestAccPreCheck(t)
			acctest.TestAccPreCheck_AccountID(t)
		},
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCloudflareR2EventNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareR2EventNotificationInitial(rnd, accountID, bucketName, queueID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "id", rnd),
				),
			},
		},
	})
}

func testAccCheckCloudflareR2EventNotificationInitial(rnd, accountID, bucketName, queueID string) string {
	return acctest.LoadTestCase("r2eventnotificationinitial.tf", rnd, accountID, bucketName, queueID)
}

func testAccCheckCloudflareR2EventNotificationDestroy(s *terraform.State) error {
	// accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare-extended_r2_event_notification" {
			continue
		}

		// client := acctest.SharedClient()
		// index, _ := client.Vectorize.Indexes.Get(
		// 	context.Background(),
		// 	rs.Primary.ID,
		// 	vectorize.IndexGetParams{
		// 		AccountID: cloudflare.F(accountID),
		// 	},
		// )
		//
		// if index != nil {
		// 	return fmt.Errorf("r2 event notification with id %s still exists", rs.Primary.ID)
		// }
	}

	return nil
}
