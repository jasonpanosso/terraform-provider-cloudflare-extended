package workers_script

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/apijson"
	"github.com/jasonpanosso/terraform-provider-cloudflare-extended/internal/customfield"
)

type WorkersScriptSettingResponseEnvelope struct {
	Result WorkersScriptMetadataModel `json:"result"`
}

type WorkersScriptModel struct {
	ID                 types.String                                                 `tfsdk:"id" path:"id,computed"`
	ScriptName         types.String                                                 `tfsdk:"script_name" path:"script_name,required"`
	AccountID          types.String                                                 `tfsdk:"account_id" path:"account_id,required"`
	Parts              customfield.NestedObjectMap[WorkersScriptPartModel]          `tfsdk:"parts" path:"parts,required"`
	Bindings           customfield.NestedObjectSet[WorkersScriptBindingsModel]      `tfsdk:"bindings" json:"bindings,optional"`
	CompatibilityDate  types.String                                                 `tfsdk:"compatibility_date" json:"compatibility_date,optional"`
	CompatibilityFlags customfield.Set[types.String]                                `tfsdk:"compatibility_flags" json:"compatibility_flags,optional"`
	KeepBindings       customfield.Set[types.String]                                `tfsdk:"keep_bindings" json:"keep_bindings,optional"`
	MainModule         types.String                                                 `tfsdk:"main_module" json:"main_module,optional"`
	BodyPart           types.String                                                 `tfsdk:"body_part" json:"body_part,optional"`
	Migrations         customfield.NestedObject[WorkersScriptMigrationsModel]       `tfsdk:"migrations" json:"migrations,optional"`
	Tags               customfield.Set[types.String]                                `tfsdk:"tags" json:"tags,optional"`
	VersionTags        map[string]types.String                                      `tfsdk:"version_tags" json:"version_tags,optional"`
	Message            types.String                                                 `tfsdk:"message" json:"message,optional"`
	Logpush            types.Bool                                                   `tfsdk:"logpush" json:"logpush,computed_optional"`
	PlacementMode      types.String                                                 `tfsdk:"placement_mode" json:"placement_mode,computed_optional"`
	UsageModel         types.String                                                 `tfsdk:"usage_model" json:"usage_model,computed_optional"`
	TailConsumers      customfield.NestedObjectSet[WorkersScriptTailConsumersModel] `tfsdk:"tail_consumers" json:"tail_consumers,computed_optional"`
	StartupTimeMs      types.Int64                                                  `tfsdk:"startup_time_ms" json:"startup_time_ms,computed"`
	CreatedOn          timetypes.RFC3339                                            `tfsdk:"created_on" json:"created_on,computed" format:"date-time"`
	ModifiedOn         timetypes.RFC3339                                            `tfsdk:"modified_on" json:"modified_on,computed" format:"date-time"`
	Etag               types.String                                                 `tfsdk:"etag" json:"etag,computed"`
}

type WorkersScriptPartModel struct {
	Part   types.String `tfsdk:"part" path:"part,required"`
	Module types.Bool   `tfsdk:"module" path:"module,optional"`
}

func (r WorkersScriptModel) MarshalMultipart() (data []byte, contentType string, err error) {
	buf := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(buf)
	bindings, _ := r.Bindings.AsStructSliceT(context.Background())
	tc, _ := r.TailConsumers.AsStructSliceT(context.Background())

	metadata := WorkersScriptMetadataModel{
		Migrations:        r.Migrations,
		VersionTags:       r.VersionTags,
		CompatibilityDate: r.CompatibilityDate,
		MainModule:        r.MainModule,
		BodyPart:          r.BodyPart,
		UsageModel:        r.UsageModel,
		// Logpush:            r.Logpush,
		Bindings:           customfield.NewObjectListMust(context.Background(), bindings),
		TailConsumers:      customfield.NewObjectListMust(context.Background(), tc),
		CompatibilityFlags: customfield.NewListMust[basetypes.StringValue](context.Background(), r.CompatibilityFlags.Elements()),
		KeepBindings:       customfield.NewListMust[basetypes.StringValue](context.Background(), r.KeepBindings.Elements()),
		Tags:               customfield.NewListMust[basetypes.StringValue](context.Background(), r.Tags.Elements()),
		Placement:          customfield.NewObjectMust(context.TODO(), &WorkersScriptMetadataPlacementModel{Mode: r.PlacementMode}),
	}
	json, err := metadata.MarshalJSON()
	if err != nil {
		return nil, "", err
	}

	err = writer.WriteField("metadata", string(json))
	if err != nil {
		writer.Close()
		return nil, "", err
	}

	parts := make(map[string]WorkersScriptPartModel)
	diags := r.Parts.ElementsAs(context.TODO(), &parts, false)
	if diags.HasError() {
		for _, err := range diags.Errors() {
			return nil, "", fmt.Errorf(err.Detail())
		}
	}

	for k, v := range parts {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes(k), escapeQuotes(k)))
		if v.Module.ValueBool() {
			h.Set("Content-Type", "text/javascript+module")
		} else {
			h.Set("Content-Type", "text/javascript")
		}

		scriptWriter, err := writer.CreatePart(h)
		if err != nil {
			writer.Close()
			return nil, "", err
		}

		_, err = scriptWriter.Write([]byte(v.Part.ValueString()))
		if err != nil {
			writer.Close()
			return nil, "", err
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, "", err
	}

	return buf.Bytes(), writer.FormDataContentType(), nil
}

type WorkersScriptMetadataModel struct {
	Bindings           customfield.NestedObjectList[WorkersScriptBindingsModel]      `tfsdk:"bindings" json:"bindings,optional"`
	BodyPart           types.String                                                  `tfsdk:"body_part" json:"body_part,optional"`
	CompatibilityDate  types.String                                                  `tfsdk:"compatibility_date" json:"compatibility_date,optional"`
	CompatibilityFlags customfield.List[types.String]                                `tfsdk:"compatibility_flags" json:"compatibility_flags,optional"`
	KeepBindings       customfield.List[types.String]                                `tfsdk:"keep_bindings" json:"keep_bindings,optional"`
	MainModule         types.String                                                  `tfsdk:"main_module" json:"main_module,optional"`
	Migrations         customfield.NestedObject[WorkersScriptMigrationsModel]        `tfsdk:"migrations" json:"migrations,optional"`
	Placement          customfield.NestedObject[WorkersScriptMetadataPlacementModel] `tfsdk:"placement" json:"placement,computed_optional"`
	Tags               customfield.List[types.String]                                `tfsdk:"tags" json:"tags,optional"`
	TailConsumers      customfield.NestedObjectList[WorkersScriptTailConsumersModel] `tfsdk:"tail_consumers" json:"tail_consumers,optional"`
	UsageModel         types.String                                                  `tfsdk:"usage_model" json:"usage_model,optional"`
	VersionTags        map[string]types.String                                       `tfsdk:"version_tags" json:"version_tags,optional"`
	Logpush            types.Bool                                                    `tfsdk:"logpush" json:"logpush,optional"`
}

func (m WorkersScriptMetadataModel) MarshalJSON() (data []byte, err error) {
	return apijson.MarshalRoot(m)
}

type WorkersScriptBindingsModel struct {
	Name          types.String `tfsdk:"name" json:"name,optional"`
	Type          types.String `tfsdk:"type" json:"type,optional"`
	BucketName    types.String `tfsdk:"bucket_name" json:"bucket_name,optional"`
	Service       types.String `tfsdk:"service" json:"service,optional"`
	Environment   types.String `tfsdk:"environment" json:"environment,optional"`
	ClassName     types.String `tfsdk:"class_name" json:"class_name,optional"`
	ScriptName    types.String `tfsdk:"script_name" json:"script_name,optional"`
	QueueName     types.String `tfsdk:"queue_name" json:"queue_name,optional"`
	ID            types.String `tfsdk:"id" json:"id,optional"`
	CertificateID types.String `tfsdk:"certificate_id" json:"certificate_id,optional"`
}

type WorkersScriptMigrationsModel struct {
	DeletedClasses     customfield.List[types.String]                                                       `tfsdk:"deleted_classes" json:"deleted_classes,optional"`
	NewClasses         customfield.List[types.String]                                                       `tfsdk:"new_classes" json:"new_classes,optional"`
	NewSqliteClasses   customfield.List[types.String]                                                       `tfsdk:"new_sqlite_classes" json:"new_sqlite_classes,optional"`
	NewTag             types.String                                                                         `tfsdk:"new_tag" json:"new_tag,optional"`
	OldTag             types.String                                                                         `tfsdk:"old_tag" json:"old_tag,optional"`
	RenamedClasses     customfield.NestedObjectList[WorkersScriptMetadataMigrationsRenamedClassesModel]     `tfsdk:"renamed_classes" json:"renamed_classes,computed_optional"`
	TransferredClasses customfield.NestedObjectList[WorkersScriptMetadataMigrationsTransferredClassesModel] `tfsdk:"transferred_classes" json:"transferred_classes,computed_optional"`
	Steps              customfield.NestedObjectList[WorkersScriptMetadataMigrationsStepsModel]              `tfsdk:"steps" json:"steps,computed_optional"`
}

type WorkersScriptMetadataMigrationsRenamedClassesModel struct {
	From types.String `tfsdk:"from" json:"from,optional"`
	To   types.String `tfsdk:"to" json:"to,optional"`
}

type WorkersScriptMetadataMigrationsTransferredClassesModel struct {
	From       types.String `tfsdk:"from" json:"from,optional"`
	FromScript types.String `tfsdk:"from_script" json:"from_script,optional"`
	To         types.String `tfsdk:"to" json:"to,optional"`
}

type WorkersScriptMetadataMigrationsStepsModel struct {
	DeletedClasses     customfield.List[types.String]                                                            `tfsdk:"deleted_classes" json:"deleted_classes,optional"`
	NewClasses         customfield.List[types.String]                                                            `tfsdk:"new_classes" json:"new_classes,optional"`
	NewSqliteClasses   customfield.List[types.String]                                                            `tfsdk:"new_sqlite_classes" json:"new_sqlite_classes,optional"`
	RenamedClasses     customfield.NestedObjectList[WorkersScriptMetadataMigrationsStepsRenamedClassesModel]     `tfsdk:"renamed_classes" json:"renamed_classes,computed_optional"`
	TransferredClasses customfield.NestedObjectList[WorkersScriptMetadataMigrationsStepsTransferredClassesModel] `tfsdk:"transferred_classes" json:"transferred_classes,computed_optional"`
}

type WorkersScriptMetadataMigrationsStepsRenamedClassesModel struct {
	From types.String `tfsdk:"from" json:"from,optional"`
	To   types.String `tfsdk:"to" json:"to,optional"`
}

type WorkersScriptMetadataMigrationsStepsTransferredClassesModel struct {
	From       types.String `tfsdk:"from" json:"from,optional"`
	FromScript types.String `tfsdk:"from_script" json:"from_script,optional"`
	To         types.String `tfsdk:"to" json:"to,optional"`
}

type WorkersScriptMetadataPlacementModel struct {
	Mode types.String `tfsdk:"mode" json:"mode,optional"`
}

type WorkersScriptTailConsumersModel struct {
	Service     types.String `tfsdk:"service" json:"service,required"`
	Environment types.String `tfsdk:"environment" json:"environment,optional"`
	Namespace   types.String `tfsdk:"namespace" json:"namespace,optional"`
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}
