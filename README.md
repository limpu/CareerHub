# 🚀 Career & LinkedIn Intelligence Platform

[![Next.js](https://img.shields.io/badge/Next.js-15-black?style=flat&logo=next.js)](https://nextjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.8-blue?style=flat&logo=typescript)](https://www.typescriptlang.org/)
[![Tailwind CSS](https://img.shields.io/badge/TailwindCSS-v4-38bdf8?style=flat&logo=tailwind-css)](https://tailwindcss.com/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](./LICENSE)
[![Security: Human-in-the-Loop](https://img.shields.io/badge/Security-Strict%20Human--in--the--Loop-emerald)](./SECURITY.md)

An open-source, privacy-first career management and authentic LinkedIn networking platform. Designed for modern job seekers, engineers, and professionals to streamline job discovery, tailor ATS-optimized resumes with guaranteed single-page layout, and scale relationship-first networking without unauthorized scraping or risk.

---

## ✨ Key Features

### 📄 1. Canonical Profile & Master Resumes
- **Heuristic Fact Extraction**: Upload existing resumes (PDF/DOCX) or enter facts manually. Facts undergo quarantined human confirmation with zero LLM hallucination (`AT-003`).
- **ATS-Optimized Vector PDF & DOCX Generation**: Built-in vector document engines producing clean, single-page, ATS-readable PDF and DOCX files.
- **Privacy Quarantine**: Candidate PII is protected and owner-isolated (`AT-011`).

### 🎯 2. Real-Time Resume Tailoring & Job Alignment
- **Position-Specific Tuning**: Align profile experience bullets and keywords with target job postings in seconds.
- **Diagnostics & Parse-Back Verification**: Instant ATS score breakdown, keyword match metrics, and layout preview.
- **Cover Letter Generation**: Matched cover letters available for immediate download in PDF and DOCX formats.

### 💼 3. Multi-Board Job Discovery
- **Aggregated Job Feed**: Integrates opportunities across Greenhouse, Lever, Ashby, and Indeed.
- **One-Click Tailor Handoff**: Clicking *"Tailor Resume for this Job"* seamlessly routes to the tailor studio with target title, company, skills, and requirements auto-populated.

### 🛡️ 4. Safe LinkedIn Strategy & Networking Studio
- **Consumer OAuth Protocol**: Official API and assisted-manual flows with zero unauthorized DOM scraping or account-ban risk (`AT-010`).
- **Outreach & Recruiter Workspace**: Manage prospective connections, recruiter leads, and authentic messaging queues.
- **Content Studio & Voice Humanizer**: Story bank, comment sweep, post ideation, and conversational voice tuning.
- **Safety Gatekeeper**: Enforces strict daily limits and circuit-breaker quotas (`REQ-009`).

### 🔄 5. Tri-Source Consistency Checker
- Cross-references your **Master Resume**, **Master Profile**, and **LinkedIn Snapshot** to highlight discrepancies in job titles, employment dates, or skill omissions.

---

## 🏗️ Repository Architecture

```text
career-platform-monorepo/
├── doc/               # Central architecture documentation & specifications
├── scripts/           # Platform bootstrap and verification scripts
└── career/            # Career & LinkedIn application
    ├── apps/
    │   └── web/       # Next.js 15 App Router frontend & studios
    ├── packages/
    │   ├── contracts/ # Canonical TypeScript schemas & interfaces
    │   └── ui/        # Shared design tokens and components
    └── services/
        └── core/      # Core Go backend services & pure vector PDF generator
```

---

## ⚡ Quick Start

### Prerequisites
- [Node.js](https://nodejs.org/) (v20+ recommended)
- [pnpm](https://pnpm.io/) (`npm install -g pnpm`)

### Installation & Run

1. **Clone the repository:**
   ```bash
   git clone https://github.com/<your-username>/<repo-name>.git
   cd <repo-name>
   ```

2. **Install monorepo dependencies:**
   ```bash
   pnpm install
   ```

3. **Start the Web Application:**
   ```bash
   cd career
   pnpm --filter web dev
   ```

4. **Open in browser:**
   Visit **[http://localhost:3000](http://localhost:3000)** to explore the platform.

---

## ⚙️ Configuration

Copy the example environment configuration:
```bash
cp .env.example .env.local
```

| Variable | Description | Default |
|---|---|---|
| `PORT` | API server port | `8080` |
| `NEXT_PUBLIC_API_URL` | Frontend API base URL | `http://localhost:8080` |
| `SAFE_MODE` | Deterministic approval mode | `true` |

---

## 🤝 Contributing

Contributions are warmly welcomed! Please read our **[CONTRIBUTING.md](./CONTRIBUTING.md)** and **[CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md)** before opening a Pull Request.

---

## 🔒 Security & Privacy

For security disclosures or privacy concerns, please refer to our **[SECURITY.md](./SECURITY.md)**.

---

## 👤 Author & Maintainer

- **Developer**: **[Atique Ullah](https://www.linkedin.com/in/atiqueullahlimon)**

---

## 📄 License

This project is licensed under the **[MIT License](./LICENSE)**.
