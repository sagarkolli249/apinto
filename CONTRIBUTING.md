### Apinto Project
- apinto
  - description：The main project of apinto open source gateway is responsible for the core content of the gateway, including forwarding, security protection, flow control, monitoring and other functions.
  - address：https://github.com/eolinker/apinto
- apinto-ingress-controller
  - description：Apinto is connected to the derivative project of kubernetes. Users can modify the configuration of apinto by loading update files on kubernetes
  - address：https://github.com/eolinker/apinto-ingress-controller
- eosc
  - description：Apinto open source gateway relies on the underlying framework. The characteristics and process operations of the open source gateway project are defined and implemented here.
  - address：https://github.com/eolinker/eosc
- apinto-docs
  - description：Apinto document project, which includes the use tutorials, development documents and other related contents of apinto. Users can cooperate synchronously and submit the supplements and changes of the tutorials through PR
  - address：https://github.com/eolinker/apinto-docs
- apinto-dashboard（In the design）
  - description：Apinto visual UI simplifies user operation and reduces user learning and use costs.
  - address：https://github.com/eolinker/apinto-dashboard
- apinto-plugin-market（In the design）
  - description：Apinto plug-in market to enhance the expansibility of apinto. Users can download and install plug-ins from the plug-in market through simple apinto cli instructions, so as to enrich the functional features of apinto open source gateway.
  - address：https://github.com/eolinker/apinto-plugin-market
- apinto-custom-plugin
  - description：Customize plugins. Developers can develop extensions according to plug-in development documents. When the plug-ins are submitted to the plug-in market, they can also get the logo of contributors.
  - course：[Plugin Development Tutorial](https://help.apinto.com/?path=/plugins/plugin_build)
### Contributor
Anyone who has one PR merged for any project in the Apinto Project is a Contributor. New Contributors should be welcomed to the community by existing members, helped with PR workflow, and directed to relevant documentation and communication channels. A contributor becomes a member of the SIG of the contribution automatically.

### How to become a Contributor
- merged at least 1 PR for any project in the Apinto Project.
You are also encouraged to participate in the projects in the following ways:
- Actively answer technical questions asked by community users in GitHub issues.
- Help test the projects
- Help review the pull requests (PRs) submitted by others
- Help improve technical documents
- Submit valuable issues
- Report or fix known and unknown bugs
- Write articles about source code analysis and usage cases for a project.

---

## Development Guidelines

### Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/apinto.git
   cd apinto
   ```

3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/sagarkolli249/apinto.git
   ```

4. **Create a branch** for your changes:
   ```bash
   git checkout -b feat/your-feature-name
   ```

### Development Setup

**Prerequisites:**
- Go 1.23.6 or later
- Docker (for building images)
- Make (optional)

**Install Dependencies:**
```bash
go mod download
go mod tidy
```

**Build Locally:**
```bash
# Build for current platform
./build/cmd/build.sh

# Build Docker image
./build/cmd/docker_build.sh 0.0.0-dev username amd64
```

### Commit Message Conventions

This project uses **Conventional Commits** for automatic semantic versioning.

**Format:**
```
<type>(<scope>): <subject>
```

**Examples:**
```bash
feat(bedrock): add streaming support
fix(cache): resolve memory leak
docs: update README
```

For detailed guidelines, see [docs/COMMIT_CONVENTIONS.md](docs/COMMIT_CONVENTIONS.md)

### CI/CD Pipeline

The project uses automated CI/CD for building and deploying:
- Automatic semantic versioning
- Multi-architecture Docker builds (amd64, arm64)
- Automated push to AWS ECR
- GitHub releases with changelogs

For setup instructions, see [docs/ECR_CI_CD_SETUP.md](docs/ECR_CI_CD_SETUP.md)

### Pull Request Process

1. Follow commit message conventions
2. Update documentation for functionality changes
3. Add tests for new features
4. Ensure all tests pass: `go test ./...`
5. Create PR with clear description

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...
```

## Documentation

- `README.md`: Project overview
- `docs/`: Detailed documentation
- `docs/COMMIT_CONVENTIONS.md`: Commit message guidelines
- `docs/ECR_CI_CD_SETUP.md`: CI/CD setup guide
