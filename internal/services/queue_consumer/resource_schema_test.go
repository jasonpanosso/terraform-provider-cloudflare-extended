package queue_consumer_test

import (
	"context"
	"testing"

	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/services/queue_consumer"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/test_helpers"
)

func TestQueueConsumerModelSchemaParity(t *testing.T) {
	t.Parallel()
	model := (*queue_consumer.QueueConsumerModel)(nil)
	schema := queue_consumer.ResourceSchema(context.TODO())
	errs := test_helpers.ValidateResourceModelSchemaIntegrity(model, schema)
	errs.Report(t)
}
