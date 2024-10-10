package serviceendpoint

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v6/serviceendpoint"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal/utils/converter"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal/utils/tfhelper"
)

// ResourceServiceEndpointGeneric schema and implementation for generic service endpoint resource
func ResourceServiceEndpointIncomingWebhook() *schema.Resource {
	r := genBaseServiceEndpointResource(flattenServiceEndpointIncomingWebhook, expandServiceEndpointIncomingWebhook)
	r.Schema["webhook_name"] = &schema.Schema{
		Type:         schema.TypeString,
		ValidateFunc: validation.StringIsNotEmpty,
		Required:     true,
		Description:  "The name of the webhook being created.",
	}
	r.Schema["http_header"] = &schema.Schema{
		Type:         schema.TypeString,
		ValidateFunc: validation.StringIsNotEmpty,
		Optional:     true,
		Description:  "Http header name on which checksum will be sent.",
	}
	r.Schema["secret"] = &schema.Schema{
		Type:             schema.TypeString,
		Description:      "Optional secret for the webhook. WebHook service will use this secret to calculate the payload checksum.",
		Sensitive:        true,
		Optional:         true,
		DiffSuppressFunc: tfhelper.DiffFuncSuppressSecretChanged,
	}
	secretHashKey, secretHashSchema := tfhelper.GenerateSecreteMemoSchema("secret")
	r.Schema[secretHashKey] = secretHashSchema
	return r
}

func expandServiceEndpointIncomingWebhook(d *schema.ResourceData) (*serviceendpoint.ServiceEndpoint, *uuid.UUID, error) {
	serviceEndpoint, projectID := doBaseExpansion(d)
	serviceEndpoint.Type = converter.String("incomingwebhook")
	serviceEndpoint.Url = converter.String("https://dev.azure.com")
	serviceEndpoint.Authorization = &serviceendpoint.EndpointAuthorization{
		Parameters: &map[string]string{
			"webhookname": d.Get("webhook_name").(string),
			"secret":      d.Get("secret").(string),
			"header":      d.Get("http_header").(string),
		},
		Scheme: converter.String("None"),
	}
	return serviceEndpoint, projectID, nil
}

func flattenServiceEndpointIncomingWebhook(d *schema.ResourceData, serviceEndpoint *serviceendpoint.ServiceEndpoint, projectID *uuid.UUID) {
	doBaseFlattening(d, serviceEndpoint, projectID)
	d.Set("webhook_name", (*serviceEndpoint.Authorization.Parameters)["webhookname"])
	tfhelper.HelpFlattenSecret(d, "secret")
	d.Set("secret", (*serviceEndpoint.Authorization.Parameters)["secret"])
	d.Set("http_header", (*serviceEndpoint.Authorization.Parameters)["header"])
}
