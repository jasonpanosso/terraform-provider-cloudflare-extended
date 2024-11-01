package queue_consumer_test

import (
	"context"
	"testing"

	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/services/queue_consumer"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/test_helpers"
)

func TestQueueConsumerDataSourceModelSchemaParity(t *testing.T) {
	t.Parallel()
	model := (*queue_consumer.QueueConsumerDataSourceModel)(nil)
	schema := queue_consumer.DataSourceSchema(context.TODO())
	errs := test_helpers.ValidateDataSourceModelSchemaIntegrity(model, schema)
	errs.Report(t)
}
