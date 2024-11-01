package vectorize_test

import (
	"context"
	"testing"

	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/services/vectorize"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/test_helpers"
)

func TestQueueConsumerModelSchemaParity(t *testing.T) {
	t.Parallel()
	model := (*vectorize.VectorizeModel)(nil)
	schema := vectorize.ResourceSchema(context.TODO())
	errs := test_helpers.ValidateResourceModelSchemaIntegrity(model, schema)
	errs.Report(t)
}
