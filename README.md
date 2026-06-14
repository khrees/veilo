# 🥷 Veilo: The Invisible Email Shield

> **"Send, receive, and reply like a ghost."**

Veilo is a self-hosted, lightweight, lightning-fast email alias forwarding and replying engine. It allows you to create instant email aliases, automatically sets up DNS on Cloudflare, registers domains in Resend, forwards incoming messages to your real inbox, and lets you reply to them invisibly without exposing your real email address.

---

## 🚀 Key Features

*   **🎨 Creative Auto-Aliases**: Make custom aliases or let Veilo generate creative, GitHub-like repository names under 25 characters (e.g. `bouncy-valley-919@yourdomain.com`).
*   **☁️ Cloudflare Auto-DNS**: Automatic configuration of MX, TXT (SPF), and verification records for your registered domains.
*   **📨 Ghost Forwarding**: Inbound emails sent to your aliases are instantly forwarded to your real email inbox.
*   **💬 Ghost Replying**: Reply directly to any forwarded email from your personal inbox. Veilo rewrites headers and forwards it back to the original sender utilizing secure reply tokens.
*   **🕵️ Background Verification Worker**: A built-in ticker checks unverified domains against Resend's API and activates them as soon as DNS propagates.
*   **🛡️ Tracker Stripper**: Automatically strips tracking pixels (1x1 tracking images or known tracker domains like Mailchimp and Sendgrid) from incoming email HTML bodies, showing a "trackers blocked" count.
*   **⏳ Self-Destructing Aliases**: Create temporary aliases that automatically disable themselves after a friendly duration (e.g., `24h`, `7d`) or after forwarding a maximum number of emails.
*   **⚡ Built with Go & Fiber**: Blazing fast performance with minimal overhead.

---

## 🗺️ How it Works (Architecture)

```mermaid
sequenceDiagram
    autonumber
    actor Sender as External Sender (sender@example.com)
    participant Veilo as Veilo App (Fiber)
    participant Resend as Resend (Email Provider)
    participant Cloudflare as Cloudflare (DNS)
    actor Recipient as Real Email (you@gmail.com)

    Note over Veilo, Cloudflare: 1. Setup Phase
    Veilo->>Resend: Register Domain
    Resend-->>Veilo: Verification DNS Records
    Veilo->>Cloudflare: Auto-Configure DNS Records
    Note over Veilo: Background verification worker polls status...
    Resend-->>Veilo: Domain Verified!

    Note over Sender, Recipient: 2. Inbound Email Flow
    Sender->>Resend: Sends email to alias@yourdomain.com
    Resend->>Veilo: Triggers inbound webhook (Svix signature)
    Veilo->>Veilo: Validates signature & matches database alias
    Veilo->>Resend: Fetches full email content
    Veilo->>Veilo: Generates Reply Token & rewrites headers
    Veilo->>Resend: Sends rewritten email to you@gmail.com
    Resend->>Recipient: Delivers email (Reply-To: reply+token@yourdomain.com)

    Note over Recipient, Sender: 3. Outbound Reply Flow
    Recipient->>Resend: Replies to reply+token@yourdomain.com
    Resend->>Veilo: Triggers inbound webhook
    Veilo->>Veilo: Validates token expiration & finds original sender
    Veilo->>Resend: Sends reply from alias@yourdomain.com
    Resend->>Sender: Delivers reply to sender@example.com
```

---

## 🛠️ Getting Started

### 📋 Prerequisites

*   **Go**: 1.21 or higher.
*   **PostgreSQL**: DB host (e.g. Supabase, RDS, local pg).
*   **Resend API Key**: To handle forwarding, receiving, and verification.
*   **Cloudflare API Token**: To configure DNS records automatically.
*   **Custom Domain**: You must own a domain name to receive and reply to emails (e.g. `yourdomain.com`).
    *   *Need a cheap domain?* You can buy one (like `.xyz`, `.top`, `.icu`, `.cfd`) for as cheap as $1 - $2/year on registrars like Porkbun, Namecheap, or Cloudflare Registrar.
    *   *Want a free option?* You can register a free subdomain with full DNS control at [EU.org](https://nic.eu.org/) (e.g. `yourname.eu.org`), which can then be set up on Cloudflare. If you are a student, check out the [GitHub Student Developer Pack](https://education.github.com/pack) for a free domain from Namecheap or Name.com.


### ⚙️ Environment Variables

Create a `.env` file in the root directory (based on `.env.example`):

```ini
# Server
PORT=8084

# Database Config
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=changeme
DB_NAME=veilo
DB_SSLMODE=disable

# Resend API Key
RESEND_API_KEY=re_your_resend_api_key

# Cloudflare Token (requires Zone.Zone Read, Zone.DNS Edit permissions)
CLOUDFLARE_API_TOKEN=cf_your_cloudflare_token

# Webhook URL of this server for automatically registering with Resend on startup
# e.g., https://smee.io/your-id or https://your-domain.com
WEBHOOK_URL=https://smee.io/your-unique-channel-id

# Webhook secret (Svix signing secret from Resend)
# Leave empty initially. If WEBHOOK_URL is set, Veilo will auto-create the webhook in Resend and print this secret in the logs.
WEBHOOK_SECRET=

# Global brand name used as suffix in forwarded emails (default: Veilo)
VIA_BRAND_NAME=Veilo

# Reply token TTL in days
REPLY_TOKEN_TTL_DAYS=90

# CORS & Limits
CORS_ORIGINS=*
RATE_LIMIT=60
```

### 🏃 Running locally

To run in development with hot-reload (using [Air](https://github.com/cosmtrek/air)):
```bash
air
```

To run directly with Go:
```bash
go run .
```

---

## 💻 Command Line Interface (CLI)

Veilo includes a unified CLI to manage your email shield directly from the terminal. The single compiled binary `veilo` handles both running the server and performing client administrative commands.

### Installation & Configuration

1. Compile the binary from source:
   ```bash
   go build -o veilo
   ```

2. Set your Veilo instance's API URL, credentials, and defaults:
   ```bash
   ./veilo config set api-url http://localhost:8084/v1
   ./veilo config set api-key your_api_key
   ./veilo config set default-domain yourdomain.com
   ./veilo config set default-email your-inbox@gmail.com
   ```

3. View your configuration settings:
   ```bash
   ./veilo config show
   ```

### CLI Commands

#### Create an Alias
Create an alias with auto-generated values, or customize its properties and self-destruct limits:
```bash
# Creative auto-generated creative slug alias
./veilo create

# Create with custom slug, domain, and real destination email
./veilo create --slug custom-slug --domain yourdomain.com --email me@gmail.com

# Create an alias that expires in 24 hours (supports durations e.g., 12h, 7d, 30d, or RFC3339 timestamp)
./veilo create --expires-at 24h

# Create an alias that self-destructs (auto-disables) after forwarding 5 emails
./veilo create --max-forwards 5
```

#### List & Manage Aliases
```bash
# List all registered aliases
./veilo list

# List only enabled aliases
./veilo list --enabled

# Get specific alias details (displays expires_at, max_forwards, forwarded count)
./veilo get <alias-address-or-id>

# Enable / Disable / Delete an alias
./veilo enable <alias-address>
./veilo disable <alias-address>
./veilo delete <alias-address>
```

#### View Statistics & Logs
```bash
# View global stats (shows Total Aliases, Total Forwarded, Total Blocked, Trackers Blocked)
./veilo stats

# View forward logs for an alias (shows sender, direction, status, and trackers blocked per email)
./veilo logs <alias-address>
```

---

## 🔌 API Documentation

All endpoints are prefixed with `/v1`. If `API_KEY` is configured in your env, requests must include the `Authorization: Bearer <key>` header.

### 🌐 Domains

#### Register a Domain
*   **POST** `/v1/domains`
*   **Body**:
    ```json
    {
      "domain": "yourdomain.com"
    }
    ```
*   **What it does**: Registers the domain on Resend, configures MX & TXT records on Cloudflare, and registers the record in the database.

#### List Registered Domains
*   **GET** `/v1/domains`

---

### 📧 Aliases

#### Create an Alias
*   **POST** `/v1/aliases`
*   **Body** (slug, address, and display_name are optional):
    ```json
    {
      "domain": "yourdomain.com",
      "real_email": "real-inbox@gmail.com",
      "display_name": "My Custom Brand",
      "label": "My Personal Label"
    }
    ```
*   **Response** (with generated creative slug):
    ```json
    {
      "success": true,
      "message": "Alias created successfully",
      "data": {
        "id": "eb7cef51-dc5f-4052-a7b8-47a56ce77f0c",
        "address": "bouncy-valley-919@yourdomain.com",
        "slug": "bouncy-valley-919",
        "domain": "yourdomain.com",
        "real_email": "real-inbox@gmail.com",
        "display_name": "My Custom Brand",
        "label": "My Personal Label",
        "enabled": true
      }
    }
    ```

---

## 🪝 Webhook & Local Setup (Automated)

Veilo automatically manages and configures webhooks on Resend for you:

1.  **Configure Webhook URL**: In your `.env` file, set `WEBHOOK_URL` to your Smee channel (e.g. `https://smee.io/your-unique-channel-id`) or your live domain.
2.  **Start your local tunnel (if testing locally)**:
    ```bash
    smee --url https://smee.io/your-unique-channel-id --port 8084
    ```
3.  **Run Veilo**: On startup, Veilo will detect your `WEBHOOK_URL`, check if the webhook is registered, and if not, register it as a Webhook Endpoint on Resend automatically. If it creates a new webhook, it will log the generated `WEBHOOK_SECRET` (Svix signing secret):
    ```
    [Warning] Automatically configured new Resend webhook pointing to https://smee.io/your-unique-channel-id. Please copy and paste this signing secret to your .env file: WEBHOOK_SECRET=whsec_...
    ```
4.  **Update WEBHOOK_SECRET**: Copy the printed `whsec_...` value and set it as `WEBHOOK_SECRET` in your `.env` to enable secure signature validation.

---

## 🤝 Contributing Guidelines

We love contributions! To help us keep Veilo clean and high-quality:

1.  **Fork** this repository.
2.  Create a feature branch: `git checkout -b feature/cool-new-stuff`.
3.  **Run Tests**: Ensure all unit/repository tests compile and pass before pushing code:
    ```bash
    go test -count=1 ./...
    ```
4.  Commit your changes: `git commit -m 'feat: add awesome new feature'`.
5.  Push to the branch: `git push origin feature/cool-new-stuff`.
6.  Create a **Pull Request**.

---

## 📜 License

Distributed under the MIT License. See `LICENSE` for more information.
