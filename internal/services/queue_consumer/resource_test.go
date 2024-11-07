package queue_consumer_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/queues"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/acctest"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/utils"
)

func TestAccCloudflareQueueConsumer_ScriptEnt(t *testing.T) {
	t.Parallel()

	rnd := utils.GenerateRandomResourceName()
	name := "cloudflare-extended_queue_consumer." + rnd
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	queueID := os.Getenv("CLOUDFLARE_QUEUE_ID")
	workerName := os.Getenv("CLOUDFLARE_WORKER_NAME")

	if queueID == "" {
		t.Fatal("CLOUDFLARE_QUEUE_ID must be set for this acceptance test")
	} else if workerName == "" {
		t.Fatal("CLOUDFLARE_WORKER_NAME must be set for this acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.TestAccPreCheck(t)
			acctest.TestAccPreCheck_AccountID(t)
		},
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCloudflareQueueConsumerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareQueueConsumerConfigInitial(rnd, accountID, queueID, workerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "script_name", workerName),
				),
			},
			{
				Config: testAccCheckCloudflareQueueConsumerConfigUpdate(rnd, accountID, queueID, workerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "type", "http_pull"),
				),
			},
		},
	})
}

func testAccCheckCloudflareQueueConsumerConfigInitial(rnd, accountID, queueID, scriptName string) string {
	return acctest.LoadTestCase("queueconsumerinitial.tf", rnd, accountID, queueID, scriptName)
}

func testAccCheckCloudflareQueueConsumerConfigUpdate(rnd, accountID, queueID, scriptName string) string {
	return acctest.LoadTestCase("queueconsumerupdate.tf", rnd, accountID, queueID, scriptName)
}

func testAccCheckCloudflareQueueConsumerDestroy(s *terraform.State) error {
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	client := cloudflare.NewClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare-extended_queue_consumer" {
			continue
		}

		consumers, err := client.Queues.Consumers.Get(
			context.Background(),
			rs.Primary.Attributes["queue_id"],
			queues.ConsumerGetParams{
				AccountID: cloudflare.F(accountID),
			})
		if err != nil {
			return err
		}

		for _, consumer := range *consumers {
			if consumer.Service == rs.Primary.Attributes["script_name"] {
				return fmt.Errorf("queue consumer for service %s still exists", rs.Primary.Attributes["script_name"])
			}
		}
	}

	return nil
}
