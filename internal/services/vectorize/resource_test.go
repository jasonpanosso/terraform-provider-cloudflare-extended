package vectorize_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/vectorize"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/acctest"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/utils"
)

const (
	dimensions = 1068
	metric     = "cosine"
)

func TestAccCloudflareVectorize_Create(t *testing.T) {
	t.Parallel()

	rnd := utils.GenerateRandomResourceName()
	name := "cloudflare-extended_vectorize_index." + rnd
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.TestAccPreCheck(t)
			acctest.TestAccPreCheck_AccountID(t)
		},
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCloudflareVectorizeDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareVectorizeIndexInitial(rnd, accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", rnd),
					resource.TestCheckResourceAttr(name, "id", rnd),
					testAccCheckCloudflareVectorizeIndexExists(rnd),
				),
			},
		},
	})
}

func testAccCheckCloudflareVectorizeIndexInitial(rnd, accountID string) string {
	return acctest.LoadTestCase("vectorizeindexinitial.tf", rnd, accountID, dimensions, metric)
}

func testAccCheckCloudflareVectorizeIndexExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acctest.SharedClient()

		rs, ok := s.RootModule().Resources["cloudflare-extended_vectorize_index."+name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		accountID := rs.Primary.Attributes["account_id"]
		_, err := client.Vectorize.Indexes.Get(
			context.Background(),
			name,
			vectorize.IndexGetParams{
				AccountID: cloudflare.String(accountID),
			})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckCloudflareVectorizeDatabaseDestroy(s *terraform.State) error {
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare-extended_vectorize_index" {
			continue
		}

		client := acctest.SharedClient()
		index, _ := client.Vectorize.Indexes.Get(
			context.Background(),
			rs.Primary.ID,
			vectorize.IndexGetParams{
				AccountID: cloudflare.F(accountID),
			},
		)

		if index != nil {
			return fmt.Errorf("vectorize index with id %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
