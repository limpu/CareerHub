# Contributing to Career & LinkedIn Platform

Thank you for your interest in contributing to **Career & LinkedIn Platform**! This project is open-source and welcomes contributions, bug reports, feature requests, and documentation improvements from the community.

---

## 🛠️ Development Setup

### Prerequisites
- **Node.js**: v20 or higher
- **pnpm**: v9 or higher
- **Go**: 1.22+ (optional, if working on core Go services)

### Quick Start
1. Fork and clone the repository:
   ```bash
   git clone https://github.com/<your-username>/<repo-name>.git
   cd <repo-name>
   ```

2. Install dependencies:
   ```bash
   pnpm install
   ```

3. Start the Next.js development server:
   ```bash
   cd career
   pnpm --filter web dev
   ```

4. Open [http://localhost:3000](http://localhost:3000) in your browser.

---

## 🌿 Branching Strategy & Workflow

1. Create a descriptive branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. Make your targeted, clean code changes.
3. Verify type safety before committing:
   ```bash
   cd career/apps/web
   npx tsc --noEmit
   ```

4. Commit with conventional commit messages:
   ```bash
   git commit -m "feat: add keyword extraction helper to resume tailoring"
   ```

5. Push to your fork and submit a Pull Request against `main`.

---

## 🔒 Security & Privacy Guidelines
- **Zero Secrets In Git**: Never commit API keys, `.env` files, or personal credentials.
- **Privacy-First**: Keep resume data parsing local or quarantined. Do not send unscrubbed candidate PII to unauthorized external endpoints.

---

## 📜 Code of Conduct
Please review and adhere to our [Code of Conduct](./CODE_OF_CONDUCT.md) during all interactions in our community.

---

## 👤 Maintainer
Developed & maintained by **[Atique Ullah](https://www.linkedin.com/in/atiqueullahlimon)**.
