package serviceendpoint

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v6/serviceendpoint"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal/utils/converter"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal/utils/tfhelper"
)

// ResourceServiceEndpointArgoCD schema and implementation for ArgoCD service endpoint resource
func ResourceServiceEndpointArgoCD() *schema.Resource {
	r := genBaseServiceEndpointResource(flattenServiceEndpointArgoCD, expandServiceEndpointArgoCD)

	r.Schema["url"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ValidateFunc: func(i interface{}, key string) (_ []string, errors []error) {
			url, ok := i.(string)
			if !ok {
				errors = append(errors, fmt.Errorf("expected type of %q to be string", key))
				return
			}
			if strings.HasSuffix(url, "/") {
				errors = append(errors, fmt.Errorf("%q should not end with slash, got %q.", key, url))
				return
			}
			return validation.IsURLWithHTTPorHTTPS(url, key)
		},
		Description: "Url for the ArgoCD Server",
	}

	patHashKey, patHashSchema := tfhelper.GenerateSecreteMemoSchema("token")
	at := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"token": {
				Description:      "The ArgoCD access token.",
				Type:             schema.TypeString,
				Required:         true,
				Sensitive:        true,
				DiffSuppressFunc: tfhelper.DiffFuncSuppressSecretChanged,
			},
			patHashKey: patHashSchema,
		},
	}

	patHashKeyU, patHashSchemaU := tfhelper.GenerateSecreteMemoSchema("username")
	patHashKeyP, patHashSchemaP := tfhelper.GenerateSecreteMemoSchema("password")
	aup := &schema.Resource{
		// Normally we don’t mark username as sensitive data, but author of the ArgoCD extension have declared this property as sensitive
		Schema: map[string]*schema.Schema{
			"username": {
				Description:      "The ArgoCD user name.",
				Type:             schema.TypeString,
				Required:         true,
				Sensitive:        true,
				DiffSuppressFunc: tfhelper.DiffFuncSuppressSecretChanged,
			},
			patHashKeyU: patHashSchemaU,
			"password": {
				Description:      "The ArgoCD password.",
				Type:             schema.TypeString,
				Required:         true,
				Sensitive:        true,
				DiffSuppressFunc: tfhelper.DiffFuncSuppressSecretChanged,
			},
			patHashKeyP: patHashSchemaP,
		},
	}

	r.Schema["authentication_token"] = &schema.Schema{
		Type:         schema.TypeList,
		Optional:     true,
		MinItems:     1,
		MaxItems:     1,
		Elem:         at,
		ExactlyOneOf: []string{"authentication_basic", "authentication_token"},
	}

	r.Schema["authentication_basic"] = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem:     aup,
	}

	return r
}

// Convert internal Terraform data structure to an AzDO data structure
func expandServiceEndpointArgoCD(d *schema.ResourceData) (*serviceendpoint.ServiceEndpoint, *uuid.UUID, error) {
	serviceEndpoint, projectID := doBaseExpansion(d)
	serviceEndpoint.Type = converter.String("argocd")
	serviceEndpoint.Url = converter.String(d.Get("url").(string))
	authScheme := "Token"

	authParams := make(map[string]string)

	if x, ok := d.GetOk("authentication_token"); ok {
		msi := x.([]interface{})[0].(map[string]interface{})
		authParams["apitoken"] = expandSecret(msi, "token")
	} else if x, ok := d.GetOk("authentication_basic"); ok {
		authScheme = "UsernamePassword"
		msi := x.([]interface{})[0].(map[string]interface{})
		authParams["username"] = expandSecret(msi, "username")
		authParams["password"] = expandSecret(msi, "password")
	}
	serviceEndpoint.Authorization = &serviceendpoint.EndpointAuthorization{
		Parameters: &authParams,
		Scheme:     &authScheme,
	}

	return serviceEndpoint, projectID, nil
}

// Convert AzDO data structure to internal Terraform data structure
func flattenServiceEndpointArgoCD(d *schema.ResourceData, serviceEndpoint *serviceendpoint.ServiceEndpoint, projectID *uuid.UUID) {
	doBaseFlattening(d, serviceEndpoint, projectID)
	if strings.EqualFold(*serviceEndpoint.Authorization.Scheme, "Token") {
		auth := make(map[string]interface{})
		if x, ok := d.GetOk("authentication_token"); ok {
			authList := x.([]interface{})[0].(map[string]interface{})
			if len(authList) > 0 {
				newHash, hashKey := tfhelper.HelpFlattenSecretNested(d, "authentication_token", authList, "token")
				auth[hashKey] = newHash
			}
		}
		if serviceEndpoint.Authorization != nil && serviceEndpoint.Authorization.Parameters != nil {
			auth["token"] = (*serviceEndpoint.Authorization.Parameters)["apitoken"]
		}
		d.Set("authentication_token", []interface{}{auth})
	} else if strings.EqualFold(*serviceEndpoint.Authorization.Scheme, "UsernamePassword") {
		auth := make(map[string]interface{})
		if old, ok := d.GetOk("authentication_basic"); ok {
			oldAuthList := old.([]interface{})[0].(map[string]interface{})
			if len(oldAuthList) > 0 {
				newHash, hashKey := tfhelper.HelpFlattenSecretNested(d, "authentication_basic", oldAuthList, "password")
				auth[hashKey] = newHash
				newHash, hashKey = tfhelper.HelpFlattenSecretNested(d, "authentication_basic", oldAuthList, "username")
				auth[hashKey] = newHash
			}
		}
		if serviceEndpoint.Authorization != nil && serviceEndpoint.Authorization.Parameters != nil {
			auth["password"] = (*serviceEndpoint.Authorization.Parameters)["password"]
			auth["username"] = (*serviceEndpoint.Authorization.Parameters)["username"]
		}
		d.Set("authentication_basic", []interface{}{auth})
	} else {
		panic(fmt.Errorf("inconsistent authorization scheme. Expected: (Token, UsernamePassword)  , but got %s", *serviceEndpoint.Authorization.Scheme))
	}

	d.Set("url", *serviceEndpoint.Url)
}
