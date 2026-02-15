## Apinto Gateway - Enhanced for AWS Bedrock

[![Go Report Card](https://goreportcard.com/badge/github.com/eolinker/apinto)](https://goreportcard.com/report/github.com/eolinker/apinto) [![Releases](https://img.shields.io/github/release/eolinker/apinto/all.svg?style=flat-square)](https://github.com/eolinker/apinto/releases) [![LICENSE](https://img.shields.io/github/license/eolinker/Apinto.svg?style=flat-square)](https://github.com/eolinker/apinto/blob/main/LICENSE)![](https://shields.io/github/downloads/eolinker/apinto/total)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](CODE_OF_CONDUCT.md)

> **This is an enhanced fork** of the original [Apinto Gateway](https://github.com/eolinker/apinto) with additional support for AWS Bedrock tool calling and improved AI provider integrations.

### What's New in This Fork

- **AWS Bedrock Tool Calling Support**: Full support for OpenAI-compatible tool calling (function calling) via AWS Bedrock
  - Automatic conversion of OpenAI tool definitions to Bedrock format
  - Support for tool result handling and multi-turn conversations
  - Comprehensive test coverage for tool calling workflows
- **Enhanced Message Handling**: Improved message format conversion between OpenAI and Bedrock APIs
- **Production-Ready**: Includes comprehensive unit tests and benchmarks
- **Kubernetes Deployment**: Enhanced Helm charts and deployment guides

---

## About Apinto

Apinto is a high-performance microservice gateway developed in Golang. It provides HTTP API forwarding, multi-tenant management, API access control, and a powerful plugin system for enterprise API management.

**Note**: The **main** branch is under active development. For stable releases, see [releases](https://github.com/sagarkolli249/apinto/releases).

### Summary

- [Why Apinto](#WhyApinto "Why Apinto")
- [Feature](#Feature)
- [Benchmark](#Benchmark)
- [Deployment](#Deployment)
- [GetStart](#GetStart "Get Start")
- [Contact](#Contact)
- [About](#About)

### Why Apinto

Apinto API gateway is a microservice gateway running on the service boundary of enterprise system. When you build websites, apps, iots and even open API transactions, Apinto API gateway can help you extract duplicate components from your internal system and run them on Apinto gateway, such as user authorization, access control, firewall, data conversion, etc; Moreover, Apinto provides the function of service arrangement, so that enterprises can quickly obtain the required data from various services and realize rapid response to business.

Apinto API gateway has the following advantages:

- Completely open source: the Apinto project is initiated and maintained by eolinker for a long time. We hope to work with global developers to build the infrastructure of micro service ecology.
- Excellent performance: under the same environment, Apinto is about 50% faster than nginx, Kong and other products, and its stability is also optimized.
- Rich functions: Apinto provides all the functions of a standard gateway, and you can quickly connect your micro services and manage network traffic.
- Extremely low use and maintenance cost: Apinto is an open source gateway developed in pure go language. It has no cumbersome deployment and no external product dependence. It only needs to download and run, which is extremely simple.
- Good scalability: most of Apinto's functions are modular, so you can easily expand its capabilities.

In a word, Apinto API gateway enables the business development team to focus more on business implementation.

### Star History

[![Star History Chart](https://api.star-history.com/svg?repos=eolinker/apinto&type=Date)](https://star-history.com/#eolinker/apinto&Date)


### Feture

| Feture                | Description                                                  |
| --------------------- | ------------------------------------------------------------ |
| Dynamic router        | Match the corresponding service by setting parameters such as location, query, header, host and method |
| Service discovery     | Support such as Eureka, Nacos and Consul                     |
| Load Balance          | Support polling weight algorithm                             |
| Authentication        | Anonymous, basic, apikey, JWT, AK / SK authentication        |
| SSL certificate       | Manage multiple certificates                                 |
| Access Domain         | The access domain can be set for the gateway                 |
| Health check          | Support health check of load nodes to ensure service robustness |
| Protocol              | HTTP/HTTPS、Webservice                                       |
| Plugin                | The process is plug-in, and the required modules are loaded on demand |
| OPEN API              | Gateway configuration using open API is supported            |
| Log                   | Provide the operation log of the node, and set the level output of the log |
| Multiple log output   | The node's request log can be output to different log receivers, such as file, NSQ, Kafka,etc |
| Cli                   | The gateway is controlled by cli command. The plug-in installation, download, opening and closing of the gateway can be controlled by one click command |
| Black and white list  | Support setting black-and-white list IP to intercept illegal IP |
| Parameter mapping     | Mapping the request parameters of the client to the forwarding request, you can change the location and name of the parameters as needed |
| Additional parameters | When forwarding the request, add back-end verification parameters, such as apikey, etc |
| Proxy rewrite         | It supports rewriting of 'scheme', 'URI', 'host', and adding or deleting the value of the request header of the forwarding request |
| flow control          | Intercept abnormal traffic                                   |

#### RoadMap

- **UI**： The gateway configuration can be operated through the UI interface, and different UI interfaces (Themes) can be customized by loading as required
- **Multi protocol**：Support a variety of protocols, including but not limited to grpc, websocket, TCP / UDP and Dubbo
- **Plugin Market**：Because Apinto mainly loads the required modules through plug-in loading, users can compile the required functions into plug-ins, or download and update the plug-ins developed by contributors from the plug-in market for one click installation
- **Service Orchestration**：An orchestration API corresponds to multiple backends. The input parameters of backends support client input and parameter transfer between backends; The returned data of backend supports filtering, deleting, moving, renaming, unpacking and packaging of fields; The orchestration API can set the exception return when the orchestration call fails
- **Monitor**：Capture the gateway request data and export it to Promethus and graphite for analysis
- .....

#### RoadMap  for 2022

![roadmap_en](https://user-images.githubusercontent.com/14105999/170408557-478830d5-3725-4fbe-a6f6-0ff0dd91d90e.jpeg)


### Benchmark

![image](https://user-images.githubusercontent.com/25589530/149748340-dc544f79-a8f9-46f5-903d-a3af4fb8b16e.png)



### Deployment

* Direct Deployment：[Deployment Tutorial](https://help.apinto.com/docs/apinto/quick/arrange.html)
* [Quick Start Tutorial](https://help.apinto.com/docs/apinto/quick/quick_course.html)
* [Source Code Compilation Tutorial](https://help.apinto.com/docs/apinto/quick/arrange.html)
* [Docker](https://hub.docker.com/r/eolinker/apinto-gateway)
* Kubernetes：Follow up support

### Get start

1. Download and unzip the installation package (here is an example of the installation package of version v0.12.1)

```
wget https://github.com/eolinker/apinto/releases/download/v0.12.1/apinto_v0.12.1_linux_amd64.tar.gz && tar -zxvf apinto_v0.12.1_linux_amd64.tar.gz && cd apinto
```
Apinto supports running on the arm64 and amd64 architectures. 

Please download the installation package of the corresponding architecture and system as required. [Click](https://github.com/eolinker/apinto/releases/) to jump to download the installation package.

2. Install the gateway:

```shell

./install.sh install

```

Executing this step will generate configuration files'/etc/apinto/apinto. yml 'and'/etc/apinto/config. yml ', which can be modified as needed.

3. Start gateway：

```
apinto start
```

3.To configure the gateway through the visual interface, click [apinto dashboard](https://github.com/eolinker/apinto-dashboard)

### AWS Bedrock Enhancements

This fork includes significant enhancements for AWS Bedrock integration:

#### Tool Calling Support
- **OpenAI-Compatible Tool Definitions**: Pass OpenAI-style tool/function definitions
- **Automatic Format Conversion**: Converts between OpenAI and Bedrock tool formats
- **Tool Name Sanitization**: Ensures tool names comply with Bedrock requirements
- **Multi-Turn Conversations**: Full support for tool use and tool result messages

#### Implementation Details
- `drivers/ai-provider/bedrock/bedrock.go`: Enhanced request conversion with tool support
- `drivers/ai-provider/bedrock/message.go`: Improved message handling for tool calls
- `drivers/ai-provider/bedrock/tools.go`: OpenAI to Bedrock tool conversion logic
- Comprehensive test coverage in `*_test.go` files

### Container Images

Pre-built container images are available on AWS ECR:
- **Repository**: `public.ecr.aws/e5v3y2z9/apinto`
- **Architectures**: amd64, arm64
- **Latest**: `1.0.0`

### Deployment

See our enhanced deployment guides:
- [Kubernetes Deployment Guide](./KUBERNETES_DEPLOYMENT_GUIDE.md)
- [Quick Start with ECR](./QUICK_START.md)
- [Build Verification](./BUILD_VERIFICATION_SUMMARY.md)
- [ECR Migration](./ECR_MIGRATION_SUMMARY.md)

### Contact & Resources

- **Original Project**: [Eolinker Apinto](https://github.com/eolinker/apinto)
- **Help Documentation**: [https://help.apinto.com](https://help.apinto.com/docs)
- **Official Website**: [https://www.apinto.com](https://www.apinto.com)
- **Community Forum**: [https://community.apinto.com](https://community.apinto.com)

### About

This enhanced version is maintained for enterprise use with AWS Bedrock. The original Apinto project is developed and maintained by [Eolink](https://www.eolink.com), a leading API management service provider.

For questions about the Bedrock enhancements, please open an issue in this repository.
