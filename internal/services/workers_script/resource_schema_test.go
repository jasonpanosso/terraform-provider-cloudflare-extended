package workers_script_test

import (
	"context"
	"testing"

	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/services/workers_script"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/test_helpers"
)

func TestWorkersScriptModelSchemaParity(t *testing.T) {
	t.Parallel()
	model := (*workers_script.WorkersScriptModel)(nil)
	schema := workers_script.ResourceSchema(context.TODO())
	errs := test_helpers.ValidateResourceModelSchemaIntegrity(model, schema)
	errs.Report(t)
}
