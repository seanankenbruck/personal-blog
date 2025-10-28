## Infrastructure (Pulumi + Azure)

This directory contains Pulumi code that provisions Azure resources for the blog (resource group, storage and networking where applicable, and the Azure App Service used to host the containerized app).

### Prerequisites

- Node.js 20+
- Pulumi CLI
- Azure subscription with permissions
- GitHub repo/environment secrets configured for OIDC login (used by CI):
  - `AZURE_CLIENT_ID`
  - `AZURE_TENANT_ID`
  - `AZURE_SUBSCRIPTION_ID`
  - `PULUMI_ACCESS_TOKEN` (for CI runs)

Pulumi stacks available: `dev`, `prod`.

### Install dependencies

```bash
cd infra
npm install
```

### Login to Azure (local)

```bash
az login
az account set --subscription <your-subscription-id>
```

### Select or create a stack

```bash
pulumi stack select dev   # or prod
```

### Configure required config (if any)

If your program requires configuration values, set them via `pulumi config set <key> <value>`.

### Preview changes

```bash
pulumi preview
```

### Deploy

```bash
pulumi up
```

### Useful outputs

The Pulumi program exports outputs used by the application pipeline. After `pulumi up`, view them:

```bash
pulumi stack output
```

Specifically:

- `appServiceName` – Azure Web App name
- `appServiceUrl` – Public URL
- `resourceGroupName` – Resource group containing the app

You can copy these into GitHub Environment secrets for `dev`/`prod` as:

- `AZURE_APP_SERVICE_NAME`
- `AZURE_APP_SERVICE_RESOURCE_GROUP`

The app deployment workflow (`.github/workflows/deploy-app.yaml`) will read these environment-scoped secrets.

### Destroy (teardown)

```bash
pulumi destroy
```

### CI usage

The `.github/workflows/deploy-infra.yaml` workflow runs `pulumi preview` and `pulumi up` with OIDC-based Azure login and can surface outputs in the run summary. Use those values to update environment secrets if they change.
