# Documentation Index

Welcome to the Go Modular Monolith documentation! This directory contains comprehensive guides for developing and maintaining the project.

## 📚 Documentation Files

### [TECHNICAL_DOCUMENTATION.md](TECHNICAL_DOCUMENTATION.md)
**Comprehensive architecture and design guide**

Complete guide to the project's architecture, including:
- Project structure and organization
- Clean architecture layers
- Module isolation pattern
- Domain-per-module pattern
- Shared kernel design
- Anti-Corruption Layer (ACL) pattern
- Event-driven architecture
- Repository pattern
- Dependency rules and isolation
- Feature flags system
- Configuration management
- Database support (SQL and MongoDB)
- Microservices readiness assessment
- Testing guidelines
- Contributing guidelines

**Best for:** Understanding the architecture, implementing new features, maintaining code quality

---

### [SCRIPTS.md](SCRIPTS.md) ⭐ NEW
**Developer scripts and workflow automation guide**

Complete guide to the developer experience scripts, including:
- Overview of all 8 scripts
- Detailed usage for each script
- Common development workflows
- Git workflow automation
- Module scaffolding
- Project health checks
- Database management
- Pre-commit automation
- Best practices
- Troubleshooting

**Best for:** Daily development, setting up workflows, automating tasks

---

### [UNIT_TESTS.md](UNIT_TESTS.md)
**Unit testing guidelines and practices**

Guide to testing in the project, including:
- Testing strategy and structure
- Mock generation
- Testing patterns
- Test organization
- Coverage requirements
- Running tests
- Best practices

**Best for:** Writing tests, understanding testing patterns, setting up test infrastructure

---

### [MOCKS.md](MOCKS.md)
**Mock generation and management guide**

Guide to generating and using mocks in the project, including:
- Mock generation setup
- Using mockgen
- Mock organization
- Interface design for mockability
- Using mocks in tests

**Best for:** Setting up mocks, understanding mock patterns, test infrastructure

---

## 🚀 Quick Start Guides

### First Time Setup
1. Read: [TECHNICAL_DOCUMENTATION.md](TECHNICAL_DOCUMENTATION.md) - Architecture Overview
2. Run: `./scripts/setup.sh`
3. Read: [SCRIPTS.md](SCRIPTS.md) - Available Commands

### Starting Daily Development
1. Check: [SCRIPTS.md](SCRIPTS.md) - Quick command reference
2. Run: `./scripts/dev.sh help`
3. Use: `./scripts/git-flow.sh start feature-name`

### Creating a New Module
1. Reference: [TECHNICAL_DOCUMENTATION.md](TECHNICAL_DOCUMENTATION.md) - Module Structure
2. Run: `./scripts/new-module.sh module-name`
3. Read: [SCRIPTS.md](SCRIPTS.md) - Module Development Workflow

### Writing Tests
1. Read: [UNIT_TESTS.md](UNIT_TESTS.md) - Testing Guidelines
2. Reference: [MOCKS.md](MOCKS.md) - Mock Setup
3. Run: `./scripts/dev.sh mocks`

---

## 🎯 By Role

### Backend Developer
1. Start: [SCRIPTS.md](SCRIPTS.md) - Setup and daily workflow
2. Reference: [TECHNICAL_DOCUMENTATION.md](TECHNICAL_DOCUMENTATION.md) - Architecture
3. Deep Dive: [UNIT_TESTS.md](UNIT_TESTS.md) - Testing

### New Team Member
1. Start: [SCRIPTS.md](SCRIPTS.md) - Quick setup guide
2. Learn: [TECHNICAL_DOCUMENTATION.md](TECHNICAL_DOCUMENTATION.md) - Full architecture
3. Practice: Create a new module using `./scripts/new-module.sh`

### DevOps/Infrastructure
1. Reference: [TECHNICAL_DOCUMENTATION.md](TECHNICAL_DOCUMENTATION.md) - Infrastructure Layer
2. Reference: [SCRIPTS.md](SCRIPTS.md) - Database Management
3. Deploy using configurations in `config/` directory

### QA/Testing
1. Start: [UNIT_TESTS.md](UNIT_TESTS.md) - Testing Guidelines
2. Reference: [MOCKS.md](MOCKS.md) - Mock Usage
3. Verify: `./scripts/dev.sh test:cover`

---

## 📋 Documentation Structure

```
docs/
├── TECHNICAL_DOCUMENTATION.md   # Architecture, patterns, design
├── SCRIPTS.md                   # Developer scripts guide
├── UNIT_TESTS.md               # Testing guidelines
├── MOCKS.md                    # Mock generation guide
└── INDEX.md                    # This file
```

---

## 🔍 Finding Information

### Looking for...

**Project Architecture?**
→ [TECHNICAL_DOCUMENTATION.md](TECHNICAL_DOCUMENTATION.md) - Overview, Architecture, Design Patterns

**How to run commands?**
→ [SCRIPTS.md](SCRIPTS.md) - Script Reference, Workflows

**How to test code?**
→ [UNIT_TESTS.md](UNIT_TESTS.md) - Testing Guidelines

**How to create mocks?**
→ [MOCKS.md](MOCKS.md) - Mock Generation

**Module Isolation Rules?**
→ [TECHNICAL_DOCUMENTATION.md](TECHNICAL_DOCUMENTATION.md) - Dependency Rules

**Setting up development environment?**
→ [SCRIPTS.md](SCRIPTS.md) - Setup Section

**Creating a new module?**
→ [SCRIPTS.md](SCRIPTS.md) - new-module.sh section

**Git workflow?**
→ [SCRIPTS.md](SCRIPTS.md) - git-flow.sh section

**Database setup?**
→ [SCRIPTS.md](SCRIPTS.md) - db.sh section

**Microservices readiness?**
→ [TECHNICAL_DOCUMENTATION.md](TECHNICAL_DOCUMENTATION.md) - Microservices Readiness

---

## 🔗 Related Files

### In Root Directory
- **ROADMAP_CHECKLIST.md** - Project roadmap and completed items
- **SCRIPTS_QUICK_REFERENCE.md** - Command cheat sheet
- **SCRIPTS_SUMMARY.md** - Executive overview of scripts
- **SCRIPTS_COMPLETE_PACKAGE.md** - Detailed scripts overview
- **Makefile** - Alternative task runner

### In scripts/ Directory
- **scripts/README.md** - Complete scripts documentation
- All executable `.sh` files with help built-in

### In config/ Directory
- **config/config.yaml** - Application configuration
- **config/featureflags.yaml** - Feature flag configuration

---

## 📚 Documentation Conventions

### Notation
- 🎯 Goal/objective
- ✅ Completed/implemented
- 🔴 Not started
- 🟡 In progress
- ⚠️ Warning/caution
- 💡 Tip/best practice

### Code Examples
- Code blocks show practical examples
- Terminal commands prefixed with `./scripts/` or `go`
- Configuration examples in YAML format

### Links
- Internal links use markdown relative paths
- External links use full URLs

---

## 🤝 Contributing to Documentation

When updating documentation:
1. Keep descriptions clear and concise
2. Include practical examples
3. Add links to related documentation
4. Use consistent formatting
5. Test commands and workflows
6. Update this index if adding new docs

---

## 📞 Support

### Getting Help

**For scripting issues:**
```bash
./scripts/[script-name].sh help
```

**For architecture questions:**
- Read [TECHNICAL_DOCUMENTATION.md](TECHNICAL_DOCUMENTATION.md)
- Check related examples in the codebase

**For testing issues:**
- Read [UNIT_TESTS.md](UNIT_TESTS.md)
- Check [MOCKS.md](MOCKS.md)

**For workflow questions:**
- Check [SCRIPTS.md](SCRIPTS.md)
- Run `./scripts/[script-name].sh help`

---

## 📈 Documentation Roadmap

Planned documentation additions:
- [ ] Debugging guide
- [ ] Performance optimization guide
- [ ] Security best practices
- [ ] Deployment guide
- [ ] Kubernetes configuration
- [ ] API documentation (OpenAPI/Swagger)
- [ ] Architecture Decision Records (ADRs)
- [ ] Troubleshooting guide

---

## ✨ Quick Navigation

| Need | Go To |
|------|-------|
| Architecture overview | [TECHNICAL_DOCUMENTATION.md](TECHNICAL_DOCUMENTATION.md) |
| Run a command | [SCRIPTS.md](SCRIPTS.md) or `./scripts/dev.sh help` |
| Setup project | [SCRIPTS.md](SCRIPTS.md) - setup.sh section |
| Test code | [UNIT_TESTS.md](UNIT_TESTS.md) |
| Generate mocks | [MOCKS.md](MOCKS.md) or `./scripts/dev.sh mocks` |
| Create module | `./scripts/new-module.sh` or [SCRIPTS.md](SCRIPTS.md) |
| Git workflow | [SCRIPTS.md](SCRIPTS.md) - git-flow.sh section |
| Database help | [SCRIPTS.md](SCRIPTS.md) - db.sh section |
| Health check | `./scripts/health-check.sh` |

---

**Last Updated:** December 3, 2025  
**Version:** 2.0.0
