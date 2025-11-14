import * as pulumi from "@pulumi/pulumi";
import * as azure from "@pulumi/azure-native";

// Get configuration
const config = new pulumi.Config();
const azureConfig = new pulumi.Config("azure");
const environment = pulumi.getStack();

// Read configuration values from ESC
const appName = config.require("appName");
const appHost = config.require("appHost");
const dnsResourceGroup = config.require("dnsResourceGroup");
const location = azureConfig.require("location");
const appServiceSku = azureConfig.require("appServiceSku");
const appServiceSkuTier = azureConfig.require("appServiceSkuTier");


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
        name: appServiceSku,
        tier: appServiceSkuTier,
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
    // Get the custom domain verification ID from the App Service
    const verificationId = appService.customDomainVerificationId;

    // Extract zone name and subdomain
    const domainParts = appHost.split('.');
    const zoneName = domainParts.slice(-2).join('.');
    const subdomain = domainParts.slice(0, -2).join('.');

    // Create TXT record for domain verification
    const verifyTxt = new azure.dns.RecordSet(`${appName}-verify-txt`, {
        resourceGroupName: dnsResourceGroup,
        zoneName: zoneName,
        recordType: "TXT",
        relativeRecordSetName: `asuid.${subdomain}`,
        ttl: 3600,
        txtRecords: verificationId.apply(id => [{ value: [id!] }]),
    });

    // Create CNAME record for the custom domain
    const cname = new azure.dns.RecordSet(`${appName}-cname`, {
        resourceGroupName: dnsResourceGroup,
        zoneName: zoneName,
        recordType: "CNAME",
        relativeRecordSetName: subdomain,
        ttl: 3600,
        cnameRecord: {
            cname: appService.defaultHostName,
        },
    });

    // Get app service managed certificate details (created manually in portal)
    const managedCert = pulumi.all([appHost, resourceGroup.name, appService.name]).apply(([host, rgName, appName]) =>
        azure.web.getCertificateOutput({
            resourceGroupName: rgName,
            name: `${host}-${appName}`
        })
    );

    // Create the custom domain binding (depends on TXT record)
    const binding = new azure.web.WebAppHostNameBinding(`${appName}-${environment}-binding`, {
        resourceGroupName: resourceGroup.name,
        name: appService.name,
        hostName: appHost,
        azureResourceName: "Website",
        customHostNameDnsRecordType: "CName",
        sslState: "SniEnabled",
        thumbprint: managedCert.apply(cert => cert.thumbprint),
    }, { dependsOn: [verifyTxt, cname] });
}

// Export important values
export const resourceGroupName = resourceGroup.name;
export const appServiceName = appService.name;
export const appServiceUrl = pulumi.interpolate`https://${appService.defaultHostName}`;