import * as pulumi from "@pulumi/pulumi";
import * as azure from "@pulumi/azure-native";

// Get configuration
const config = new pulumi.Config();
const azureConfig = new pulumi.Config("azure");
const environment = pulumi.getStack();

// Read configuration values from ESC
const location = azureConfig.require("location");
const appName = config.require("appName");


// Create resource group
const resourceGroup = new azure.resources.ResourceGroup(`${appName}-${environment}-rg`, {
    location: location,
    tags: {
        Environment: environment,
        ManagedBy: "Pulumi",
        Application: "PersonalBlog",
    },
});

// Create app service plan
const appServicePlan = new azure.web.AppServicePlan(`${appName}-${environment}-plan`, {
    resourceGroupName: resourceGroup.name,
    location: resourceGroup.location,
    kind: "Linux",
    reserved: true,
    sku: {
        name: environment === "prod" ? "B1" : "F1",
        tier: environment === "prod" ? "Basic" : "Free",
    },
    tags: {
        Environment: environment,
    },
});

// Create App Service
const appService = new  azure.web.WebApp(`${appName}-${environment}-app`, {
    resourceGroupName: resourceGroup.name,
    location: resourceGroup.location,
    serverFarmId: appServicePlan.id,
    kind: "app,linux,container",
    siteConfig: {
        linuxFxVersion: "DOCKER|smankenb/personal-blog:latest",
        alwaysOn: environment === "prod",
        http20Enabled: true,
        minTlsVersion: "1.2",
        ftpsState: "Disabled",
        healthCheckPath: "/health",
        appSettings: [
            {
                name: "WEBSITES_ENABLE_APP_SERVICE_STORAGE",
                value: "false",
            },
            {
                name: "GIN_MODE",
                value: "release",
            },
            {
                name: "CONTENT_DIR",
                value: "/content/posts",
            },
            // OpenTelemetry configuration for Grafana Cloud
            {
                name: "OTEL_EXPORTER_OTLP_ENDPOINT",
                value: config.get("otlpEndpoint") || "",
            },
            {
                name: "OTEL_EXPORTER_OTLP_HEADERS",
                value: config.getSecret("otlpHeaders") || "",
            },
            {
                name: "OTEL_SERVICE_NAME",
                value: `${appName}-${environment}`,
            },
            {
                name: "OTEL_RESOURCE_ATTRIBUTES",
                value: `service.name=${appName}-${environment},deployment.environment=${environment}`,
            },
        ],
    },
    httpsOnly: true,
    tags: {
        Environment: environment,
    },
});

// Configure custom domain (prod only)
if (environment === "prod") {
    const customDomain = new azure.web.WebAppHostNameBinding(`${appName}-${environment}-domain`, {
        resourceGroupName: resourceGroup.name,
        name: appService.name,
        hostName: "blog.seanankenbruck.com",
    });
}

// Export important values
export const resourceGroupName = resourceGroup.name;
export const appServiceName = appService.name;
export const appServiceUrl = pulumi.interpolate`https://${appService.defaultHostName}`;