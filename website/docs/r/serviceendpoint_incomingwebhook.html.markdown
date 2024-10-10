---
layout: "azuredevops"
page_title: "AzureDevops: azuredevops_serviceendpoint_incomingwebhook"
description: |-
  Manages an incoming webhook service endpoint that can be used as a trigger Azure DevOps pipelines.
---

# azuredevops_serviceendpoint_incomingwebhook

Manages an incoming webhook service endpoint that can be used as a trigger Azure DevOps pipelines.

## Example Usage

```hcl
resource "azuredevops_project" "project" {
  name       = "Sample Project"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "azuredevops_serviceendpoint_incomingwebhook" "serviceendpoint" {
  project_id            = azuredevops_project.project.id
  webhook_name          = "My Incoming Webhook"
  secret                = "secretTextForPayloadChecksum"
  service_endpoint_name = "Sample Incoming Webhook"
  description           = "Managed by Terraform"
}
```

## Argument Reference

The following arguments are supported:

- `project_id` - (Required) The project ID or project name.
- `service_endpoint_name` - (Required) The service endpoint name.
- `webhook_name` - (Required) The name of the webhook being created.
- `secret` - (Optional) Secret for the webhook. WebHook service will use this secret to calculate the payload checksum.
- `http_header` - (Optional) Http header name on which checksum will be sent.
- `description` - (Optional) The Service Endpoint description. Defaults to `Managed by Terraform`.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the service endpoint.
- `project_id` - The ID of the project associated with the service endpoint.
- `service_endpoint_name` - The name of the service endpoint.

## Relevant Links

- [Azure DevOps Service REST API 6.0](https://docs.microsoft.com/en-us/rest/api/azure/devops/serviceendpoint/endpoints?view=azure-devops-rest-6.0)
- [Azure DevOps Webhooks Integration](https://docs.microsoft.com/en-us/azure/devops/service-hooks/services/webhooks?view=azure-devops)

## Import

Azure DevOps Service Endpoint Incoming Webhook can be imported using **projectID/serviceEndpointID** or
**projectName/serviceEndpointID**

```sh
$ terraform import azuredevops_serviceendpoint_incomingwebhook.serviceendpoint 00000000-0000-0000-0000-000000000000/00000000-0000-0000-0000-000000000000
```
