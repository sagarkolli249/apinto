# Commit Message Conventions

This project uses **Conventional Commits** for automatic semantic versioning and changelog generation.

## Quick Reference

### Format

```
<type>(<scope>): <subject>

[optional body]

[optional footer]
```

### Common Examples

#### Features (Minor Version Bump)
```bash
git commit -m "feat: add AWS Bedrock tools support"
git commit -m "feat(auth): implement OAuth2 authentication"
git commit -m "feat(api): add rate limiting for AI providers"
```

#### Bug Fixes (Patch Version Bump)
```bash
git commit -m "fix: resolve memory leak in cache strategy"
git commit -m "fix(bedrock): handle null responses correctly"
git commit -m "fix(grey-strategy): correct routing weight calculation"
```

#### Breaking Changes (Major Version Bump)
```bash
git commit -m "feat!: redesign strategy interface

BREAKING CHANGE: All strategy implementations must implement new Execute method"

git commit -m "refactor!: change API response format

BREAKING CHANGE: API now returns JSON instead of XML"
```

#### Other Types (No Version Bump)
```bash
# Documentation
git commit -m "docs: update ECR deployment guide"

# Code refactoring (Patch bump)
git commit -m "refactor: simplify error handling in actuators"

# Performance improvements (Patch bump)
git commit -m "perf: optimize database query in cache strategy"

# Tests
git commit -m "test: add unit tests for bedrock tools"

# Chores/Maintenance
git commit -m "chore: update dependencies"
git commit -m "chore: configure semantic-release"

# CI/CD
git commit -m "ci: add ECR build workflow"
git commit -m "ci: fix docker build script permissions"
```

## Commit Types

| Type | Description | Version Impact | Example |
|------|-------------|----------------|---------|
| `feat` | New feature | Minor (0.1.0 → 0.2.0) | `feat: add new cache backend` |
| `fix` | Bug fix | Patch (0.1.0 → 0.1.1) | `fix: resolve connection timeout` |
| `perf` | Performance improvement | Patch | `perf: improve query speed` |
| `refactor` | Code restructuring | Patch | `refactor: extract common logic` |
| `docs` | Documentation | None | `docs: update README` |
| `style` | Formatting, missing semicolons | None | `style: format with gofmt` |
| `test` | Adding tests | None | `test: add integration tests` |
| `build` | Build system changes | None | `build: update go.mod` |
| `ci` | CI configuration | None | `ci: add GitHub Actions` |
| `chore` | Other maintenance | None | `chore: update gitignore` |

## Scopes

Use scopes to specify which part of the codebase is affected:

### Common Scopes
- `auth` - Authentication/Authorization
- `api` - API endpoints
- `bedrock` - AWS Bedrock provider
- `cache` - Cache strategy
- `grey` - Grey strategy
- `limiting` - Rate limiting
- `fuse` - Circuit breaker
- `docs` - Documentation
- `ci` - CI/CD pipelines
- `deps` - Dependencies

### Examples with Scopes
```bash
git commit -m "feat(bedrock): add streaming support"
git commit -m "fix(cache): resolve race condition"
git commit -m "docs(api): add authentication examples"
git commit -m "refactor(limiting): simplify rate limit logic"
```

## Multi-line Commits

For complex changes, use multi-line commit messages:

```bash
git commit -m "feat(api): add GraphQL support

- Implement GraphQL schema
- Add query resolvers
- Add mutation resolvers
- Update documentation

Closes #123"
```

## Breaking Changes

Mark breaking changes with `!` or in the footer:

### Method 1: Using `!`
```bash
git commit -m "feat!: change configuration format"
```

### Method 2: In Footer
```bash
git commit -m "feat: update API endpoints

BREAKING CHANGE: /api/v1/* endpoints moved to /api/v2/*"
```

### Method 3: Both
```bash
git commit -m "refactor!: redesign plugin system

BREAKING CHANGE: Plugins must now implement PluginV2 interface instead of Plugin.
Migration guide: https://docs.example.com/migration"
```

## Version Bump Examples

### From 1.0.0 to 1.0.1 (Patch)
```bash
git commit -m "fix: correct validation logic"
# or
git commit -m "perf: optimize memory usage"
```

### From 1.0.0 to 1.1.0 (Minor)
```bash
git commit -m "feat: add webhook support"
```

### From 1.0.0 to 2.0.0 (Major)
```bash
git commit -m "feat!: redesign authentication system

BREAKING CHANGE: JWT tokens now use different format"
```

### No Version Bump
```bash
git commit -m "docs: fix typo in README"
git commit -m "test: add missing test cases"
git commit -m "chore: update dependencies"
```

## Pre-release Versions

On the `develop` branch, commits create pre-release versions:

```bash
# On develop branch
git commit -m "feat: experimental feature"
# Creates: 1.1.0-beta.1

git commit -m "fix: bug in experimental feature"
# Creates: 1.1.0-beta.2
```

## Tips for Good Commit Messages

### DO ✅
- Use imperative mood: "add feature" not "added feature"
- Start with lowercase: "feat: add" not "feat: Add"
- Be specific and concise
- Reference issues: "Closes #123" or "Fixes #456"
- Explain *why*, not just *what*
- Keep subject line under 72 characters

### DON'T ❌
- Don't use generic messages: "update code", "fix stuff"
- Don't include multiple unrelated changes in one commit
- Don't forget the type prefix
- Don't use past tense: "added" or "fixed"
- Don't include implementation details in subject

## Good vs Bad Examples

### ❌ Bad
```bash
git commit -m "fixed it"
git commit -m "updates"
git commit -m "Added new feature for caching and fixed auth bug"
git commit -m "Fix: bug fix"  # Wrong case
```

### ✅ Good
```bash
git commit -m "fix(cache): resolve memory leak in Redis connection"
git commit -m "feat(auth): add OAuth2 support for Google"
git commit -m "perf(api): reduce response time by 50%"
git commit -m "docs: add deployment guide for Kubernetes"
```

## Commit Message Template

Create a commit message template:

```bash
# ~/.gitmessage
# <type>(<scope>): <subject>
#
# <body>
#
# <footer>

# Type: feat, fix, docs, style, refactor, perf, test, build, ci, chore
# Scope: auth, api, cache, bedrock, etc.
# Subject: imperative, lowercase, no period
#
# Body: explain what and why (not how)
#
# Footer: Breaking changes, issue references
```

Configure git to use the template:
```bash
git config --global commit.template ~/.gitmessage
```

## Verification

Test your commit message format before pushing:

```bash
# Install commitlint (optional)
npm install -g @commitlint/cli @commitlint/config-conventional

# Test a commit message
echo "feat: add new feature" | commitlint

# Set up git hook (optional)
npx husky add .husky/commit-msg 'npx --no -- commitlint --edit ${1}'
```

## Interactive Commit Helper

For easier commits, use commitizen:

```bash
# Install commitizen
npm install -g commitizen cz-conventional-changelog

# Configure project
commitizen init cz-conventional-changelog --save-dev --save-exact

# Make commits interactively
git cz
```

## Resources

- [Conventional Commits Specification](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
- [Angular Commit Guidelines](https://github.com/angular/angular/blob/main/CONTRIBUTING.md#commit)
- [Semantic Release Documentation](https://semantic-release.gitbook.io/)

## Questions?

If you're unsure about the commit message format:
1. Check recent commits: `git log --oneline`
2. Refer to this guide
3. Ask in pull request reviews
4. When in doubt, use: `feat:` for new things, `fix:` for bug fixes
