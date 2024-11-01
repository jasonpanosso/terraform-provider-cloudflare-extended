package workers_script_test

import (
	"context"
	"testing"

	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/services/workers_script"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/test_helpers"
)

func TestWorkersScriptDataSourceModelSchemaParity(t *testing.T) {
	t.Parallel()
	model := (*workers_script.WorkersScriptDataSourceModel)(nil)
	schema := workers_script.DataSourceSchema(context.TODO())
	errs := test_helpers.ValidateDataSourceModelSchemaIntegrity(model, schema)
	errs.Report(t)
}
