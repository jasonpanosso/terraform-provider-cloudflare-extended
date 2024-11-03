package r2_event_notification_test

import (
	"context"
	"testing"

	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/services/r2_event_notification"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/test_helpers"
)

func TestQueueConsumerModelSchemaParity(t *testing.T) {
	t.Parallel()
	model := (*r2_event_notification.R2EventNotificationModel)(nil)
	schema := r2_event_notification.ResourceSchema(context.TODO())
	errs := test_helpers.ValidateResourceModelSchemaIntegrity(model, schema)
	errs.Report(t)
}
