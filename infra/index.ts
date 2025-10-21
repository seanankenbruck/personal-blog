import * as pulumi from "@pulumi/pulumi";
import * as azure from "@pulumi/azure-native";

// Get configuration
const config = new pulumi.Config();
const environment = pulumi.getStack();
const location = config.get("location");
const appName = config.require("appName");
