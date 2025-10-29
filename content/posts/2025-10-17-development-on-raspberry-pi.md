---
title: "Building a Development Environment on Raspberry Pi"
date: 2025-12-01T14:30:00Z
slug: "development-on-raspberry-pi"
tags: ["technical", "development", "raspberry pi", "kubernetes"]
description: "Building an environment for development on Raspberry Pi"
published: false
---

# Developing on Raspberry Pi

I have always wanted to find a usecase for learning how to use Raspberry Pis and figured that this blog application would be a great motivator to really get my hands dirty and see what they are capable of. What is a Raspberry Pi? It is essentially a low-cost, no frills computer that is small enough to fit in your pocket, but powerful enough to provide computing power to an array of different types of tasks. I decided that for this project, I wanted to connect several raspberry pi units together so that I could configure a kubernetes cluster and deploy containerized applications in the same way that you would deploy them to Azure or AWS (without the additional costs). 

## Benefits

- **Cost-Effective Learning**: A multi-node Raspberry Pi cluster costs a fraction of running equivalent infrastructure on AKS or EKS. No ongoing cloud bills or surprise charges for compute, storage, or data transfer.
- **Physical Hardware Control**: Complete control over the hardware means you can experiment with networking, storage configurations, and cluster architecture without worrying about cloud provider limitations or quotas.
- **Hands-On Experience**: Working with bare metal gives you deeper understanding of Kubernetes networking, storage, and node management that can be abstracted away in managed services.
- **Always-On Development Environment**: Your cluster is available 24/7 in your home without needing to manage cloud resources or worry about accidentally leaving expensive instances running.
- **Energy Efficient**: Raspberry Pis consume minimal power (typically 5-15W per unit) compared to traditional servers or cloud infrastructure, making them environmentally friendly and cheap to run continuously.
- **Portable and Compact**: The entire cluster can fit in a small space or even a backpack, making it ideal for demos, education, or development on the go.

## Disadvantages

- **Limited Resources**: Raspberry Pis have constrained CPU, memory, and storage compared to cloud instances. This limits the complexity and scale of applications you can run effectively.
- **ARM Architecture**: Many container images are built for x86/AMD64 architecture and may not work on ARM-based Raspberry Pis without rebuilding or finding ARM-compatible alternatives.
- **Manual Maintenance**: Unlike managed Kubernetes services (AKS/EKS), you're responsible for all cluster updates, security patches, backups, and hardware failures.
- **Network Reliability**: Your cluster depends on your home network and power supply. Cloud providers offer enterprise-grade networking, uptime SLAs, and DDoS protection.
- **Scaling Limitations**: Adding nodes means purchasing and configuring physical hardware. Cloud services let you scale up or down instantly with API calls.
- **Storage Performance**: SD cards and USB storage used with Raspberry Pis are significantly slower than cloud SSD/NVMe storage, which can impact application performance.

The benefits in this case definitely outweigh the disadvantages from a development standpoint. I can spend as much time as I need writing the application code and don't have to worry about the costs incurred from running a development environment in the cloud before I am ready to actually use it.

## Technical Details

The rest of this post will outline the resources required and steps that I went through to build this local development environment. I won't get too into the details for any one component as that was already done in [this post](https://anthonynsimon.com/blog/kubernetes-cluster-raspberry-pi/).

### The Stack

### Networking

### The Cluster

### Storage

### Deployment Scripts

## Conclusion