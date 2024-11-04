package r2_event_notification_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/event_notifications"
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
					resource.TestCheckResourceAttr(name, "bucket_name", bucketName),
					resource.TestCheckResourceAttr(name, "queue_id", queueID),
				),
			},
			{
				Config: testAccCheckCloudflareR2EventNotificationUpdate(rnd, accountID, bucketName, queueID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "bucket_name", bucketName),
					resource.TestCheckResourceAttr(name, "queue_id", queueID),
				),
			},
			{
				Config: testAccCheckCloudflareR2EventNotificationUpdate2(rnd, accountID, bucketName, queueID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "bucket_name", bucketName),
					resource.TestCheckResourceAttr(name, "queue_id", queueID),
				),
			},
		},
	})
}

func testAccCheckCloudflareR2EventNotificationInitial(rnd, accountID, bucketName, queueID string) string {
	return acctest.LoadTestCase("r2eventnotificationinitial.tf", rnd, accountID, bucketName, queueID)
}

func testAccCheckCloudflareR2EventNotificationUpdate(rnd, accountID, bucketName, queueID string) string {
	return acctest.LoadTestCase("r2eventnotificationupdate.tf", rnd, accountID, bucketName, queueID)
}

func testAccCheckCloudflareR2EventNotificationUpdate2(rnd, accountID, bucketName, queueID string) string {
	return acctest.LoadTestCase("r2eventnotificationupdate2.tf", rnd, accountID, bucketName, queueID)
}

func testAccCheckCloudflareR2EventNotificationDestroy(s *terraform.State) error {
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare-extended_r2_event_notification" {
			continue
		}

		client := acctest.SharedClient()
		event_notifications, _ := client.EventNotifications.R2.Configuration.Get(
			context.Background(),
			rs.Primary.ID,
			event_notifications.R2ConfigurationGetParams{
				AccountID: cloudflare.F(accountID),
			},
		)

		if event_notifications != nil {
			return fmt.Errorf("r2 event notification with id %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
