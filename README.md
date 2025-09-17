# Crush

<p align="center">
    <a href="https://stuff.charm.sh/crush/charm-crush.png"><img width="450" alt="Charm Crush Logo" src="https://github.com/user-attachments/assets/adc1a6f4-b284-4603-836c-59038caa2e8b" /></a><br />
    <a href="https://github.com/charmbracelet/crush/releases"><img src="https://img.shields.io/github/release/charmbracelet/crush" alt="Latest Release"></a>
    <a href="https://github.com/charmbracelet/crush/actions"><img src="https://github.com/charmbracelet/crush/actions/workflows/build.yml/badge.svg" alt="Build Status"></a>
</p>

<p align="center">Your new coding bestie, now available in your favourite terminal.<br />Your tools, your code, and your workflows, wired into your LLM of choice.</p>

<p align="center"><img width="800" alt="Crush Demo" src="https://github.com/user-attachments/assets/58280caf-851b-470a-b6f7-d5c4ea8a1968" /></p>

## Features

- **Multi-Model:** choose from a wide range of LLMs or add your own via OpenAI- or Anthropic-compatible APIs
- **Flexible:** switch LLMs mid-session while preserving context
- **Session-Based:** maintain multiple work sessions and contexts per project
- **LSP-Enhanced:** Crush uses LSPs for additional context, just like you do
- **Extensible:** add capabilities via MCPs (`http`, `stdio`, and `sse`)
- **Works Everywhere:** first-class support in every terminal on macOS, Linux, Windows (PowerShell and WSL), FreeBSD, OpenBSD, and NetBSD

## Installation

Use a package manager:

```bash
# Homebrew
brew install charmbracelet/tap/crush

# NPM
npm install -g @charmland/crush

# Arch Linux (btw)
yay -S crush-bin

# Nix
nix run github:numtide/nix-ai-tools#crush
```

Windows users:

```bash
# Winget
winget install charmbracelet.crush

# Scoop
scoop bucket add charm https://github.com/charmbracelet/scoop-bucket.git
scoop install crush
```

<details>
<summary><strong>Nix (NUR)</strong></summary>

Crush is available via [NUR](https://github.com/nix-community/NUR) in `nur.repos.charmbracelet.crush`.

You can also try out Crush via `nix-shell`:

```bash
# Add the NUR channel.
nix-channel --add https://github.com/nix-community/NUR/archive/main.tar.gz nur
nix-channel --update

# Get Crush in a Nix shell.
nix-shell -p '(import <nur> { pkgs = import <nixpkgs> {}; }).repos.charmbracelet.crush'
```

</details>

<details>
<summary><strong>Debian/Ubuntu</strong></summary>

```bash
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://repo.charm.sh/apt/gpg.key | sudo gpg --dearmor -o /etc/apt/keyrings/charm.gpg
echo "deb [signed-by=/etc/apt/keyrings/charm.gpg] https://repo.charm.sh/apt/ * *" | sudo tee /etc/apt/sources.list.d/charm.list
sudo apt update && sudo apt install crush
```

</details>

<details>
<summary><strong>Fedora/RHEL</strong></summary>

```bash
echo '[charm]
name=Charm
baseurl=https://repo.charm.sh/yum/
enabled=1
gpgcheck=1
gpgkey=https://repo.charm.sh/yum/gpg.key' | sudo tee /etc/yum.repos.d/charm.repo
sudo yum install crush
```

</details>

Or, download it:

- [Packages][releases] are available in Debian and RPM formats
- [Binaries][releases] are available for Linux, macOS, Windows, FreeBSD, OpenBSD, and NetBSD

[releases]: https://github.com/charmbracelet/crush/releases

Or just install it with Go:

```
go install github.com/charmbracelet/crush@latest
```

> [!WARNING]
> Productivity may increase when using Crush and you may find yourself nerd
> sniped when first using the application. If the symptoms persist, join the
> [Discord][discord] and nerd snipe the rest of us.

## Getting Started

The quickest way to get started is to grab an API key for your preferred
provider such as Anthropic, OpenAI, Groq, or OpenRouter and just start
Crush. You'll be prompted to enter your API key.

That said, you can also set environment variables for preferred providers.

| Environment Variable       | Provider                                           |
| -------------------------- | -------------------------------------------------- |
| `ANTHROPIC_API_KEY`        | Anthropic                                          |
| `OPENAI_API_KEY`           | OpenAI                                             |
| `OPENROUTER_API_KEY`       | OpenRouter                                         |
| `CEREBRAS_API_KEY`         | Cerebras                                           |
| `GEMINI_API_KEY`           | Google Gemini                                      |
| `VERTEXAI_PROJECT`         | Google Cloud VertexAI (Gemini)                     |
| `VERTEXAI_LOCATION`        | Google Cloud VertexAI (Gemini)                     |
| `GROQ_API_KEY`             | Groq                                               |
| `AWS_ACCESS_KEY_ID`        | AWS Bedrock (Claude)                               |
| `AWS_SECRET_ACCESS_KEY`    | AWS Bedrock (Claude)                               |
| `AWS_REGION`               | AWS Bedrock (Claude)                               |
| `AZURE_OPENAI_ENDPOINT`    | Azure OpenAI models                                |
| `AZURE_OPENAI_API_KEY`     | Azure OpenAI models (optional when using Entra ID) |
| `AZURE_OPENAI_API_VERSION` | Azure OpenAI models                                |

### By the Way

Is there a provider you’d like to see in Crush? Is there an existing model that needs an update?

Crush’s default model listing is managed in [Catwalk](https://github.com/charmbracelet/catwalk), a community-supported, open source repository of Crush-compatible models, and you’re welcome to contribute.

<a href="https://github.com/charmbracelet/catwalk"><img width="174" height="174" alt="Catwalk Badge" src="https://github.com/user-attachments/assets/95b49515-fe82-4409-b10d-5beb0873787d" /></a>

## Configuration

Crush runs great with no configuration. That said, if you do need or want to
customize Crush, configuration can be added either local to the project itself,
or globally, with the following priority:

1. `.crush.json`
2. `crush.json`
3. `$HOME/.config/crush/crush.json` (Windows: `%USERPROFILE%\AppData\Local\crush\crush.json`)

Configuration itself is stored as a JSON object:

```json
{
  "this-setting": { "this": "that" },
  "that-setting": ["ceci", "cela"]
}
```

As an additional note, Crush also stores ephemeral data, such as application state, in one additional location:

```bash
# Unix
$HOME/.local/share/crush/crush.json

# Windows
%LOCALAPPDATA%\crush\crush.json
```

### LSPs

Crush can use LSPs for additional context to help inform its decisions, just
like you would. LSPs can be added manually like so:

```json
{
  "$schema": "https://charm.land/crush.json",
  "lsp": {
    "go": {
      "command": "gopls",
      "env": {
        "GOTOOLCHAIN": "go1.24.5"
      }
    },
    "typescript": {
      "command": "typescript-language-server",
      "args": ["--stdio"]
    },
    "nix": {
      "command": "nil"
    }
  }
}
```

### MCPs

Crush also supports Model Context Protocol (MCP) servers through three
transport types: `stdio` for command-line servers, `http` for HTTP endpoints,
and `sse` for Server-Sent Events. Environment variable expansion is supported
using `$(echo $VAR)` syntax.

```json
{
  "$schema": "https://charm.land/crush.json",
  "mcp": {
    "filesystem": {
      "type": "stdio",
      "command": "node",
      "args": ["/path/to/mcp-server.js"],
      "env": {
        "NODE_ENV": "production"
      }
    },
    "github": {
      "type": "http",
      "url": "https://example.com/mcp/",
      "headers": {
        "Authorization": "$(echo Bearer $EXAMPLE_MCP_TOKEN)"
      }
    },
    "streaming-service": {
      "type": "sse",
      "url": "https://example.com/mcp/sse",
      "headers": {
        "API-Key": "$(echo $API_KEY)"
      }
    }
  }
}
```

### Ignoring Files

Crush respects `.gitignore` files by default, but you can also create a
`.crushignore` file to specify additional files and directories that Crush
should ignore. This is useful for excluding files that you want in version
control but don't want Crush to consider when providing context.

The `.crushignore` file uses the same syntax as `.gitignore` and can be placed
in the root of your project or in subdirectories.

### Allowing Tools

By default, Crush will ask you for permission before running tool calls. If
you'd like, you can allow tools to be executed without prompting you for
permissions. Use this with care.

```json
{
  "$schema": "https://charm.land/crush.json",
  "permissions": {
    "allowed_tools": [
      "view",
      "ls",
      "grep",
      "edit",
      "mcp_context7_get-library-doc"
    ]
  }
}
```

You can also skip all permission prompts entirely by running Crush with the
`--yolo` flag. Be very, very careful with this feature.

### Local Models

Local models can also be configured via OpenAI-compatible API. Here are two common examples:

#### Ollama

```json
{
  "providers": {
    "ollama": {
      "name": "Ollama",
      "base_url": "http://localhost:11434/v1/",
      "type": "openai",
      "models": [
        {
          "name": "Qwen 3 30B",
          "id": "qwen3:30b",
          "context_window": 256000,
          "default_max_tokens": 20000
        }
      ]
    }
  }
}
```

#### LM Studio

```json
{
  "providers": {
    "lmstudio": {
      "name": "LM Studio",
      "base_url": "http://localhost:1234/v1/",
      "type": "openai",
      "models": [
        {
          "name": "Qwen 3 30B",
          "id": "qwen/qwen3-30b-a3b-2507",
          "context_window": 256000,
          "default_max_tokens": 20000
        }
      ]
    }
  }
}
```

#### llama.cpp

```json
{
  "providers": {
    "llamacpp": {
      "name": "llama.cpp Direct",
      "base_url": "http://localhost:8080/v1/",
      "type": "openai",
      "disable_stream": true,
      "models": [
        {
          "name": "Qwen2.5-14B-Instruct",
          "id": "qwen2.5-14b-instruct",
          "context_window": 8192,
          "default_max_tokens": 2048
        }
      ]
    }
  }
}
```

> llama.cpp's OpenAI-compatible server does not currently support streaming responses. Setting `disable_stream` tells Crush to fall back to a non-streaming flow for this provider.

### Custom Providers

Crush supports custom provider configurations for both OpenAI-compatible and
Anthropic-compatible APIs.

#### OpenAI-Compatible APIs

Here’s an example configuration for Deepseek, which uses an OpenAI-compatible
API. Don't forget to set `DEEPSEEK_API_KEY` in your environment.

```json
{
  "$schema": "https://charm.land/crush.json",
  "providers": {
    "deepseek": {
      "type": "openai",
      "base_url": "https://api.deepseek.com/v1",
      "api_key": "$DEEPSEEK_API_KEY",
      "models": [
        {
          "id": "deepseek-chat",
          "name": "Deepseek V3",
          "cost_per_1m_in": 0.27,
          "cost_per_1m_out": 1.1,
          "cost_per_1m_in_cached": 0.07,
          "cost_per_1m_out_cached": 1.1,
          "context_window": 64000,
          "default_max_tokens": 5000
        }
      ]
    }
  }
}
```

##### Qwen DashScope / ModelScope / OpenRouter

You can route Crush through any OpenAI-compatible provider. Below are common presets
for Qwen backends (pick one, and set the appropriate API key):

DashScope (Mainland China):

```json
{
  "$schema": "https://charm.land/crush.json",
  "providers": {
    "dashscope-cn": {
      "type": "openai",
      "name": "DashScope (CN)",
      "base_url": "https://dashscope.aliyuncs.com/compatible-mode/v1",
      "api_key": "$DASHSCOPE_API_KEY",
      "models": [
        { "id": "qwen3-coder-plus", "name": "Qwen3 Coder Plus", "context_window": 131072, "default_max_tokens": 8192 }
      ]
    }
  }
}
```

DashScope (International):

```json
{
  "$schema": "https://charm.land/crush.json",
  "providers": {
    "dashscope-intl": {
      "type": "openai",
      "name": "DashScope (Intl)",
      "base_url": "https://dashscope-intl.aliyuncs.com/compatible-mode/v1",
      "api_key": "$DASHSCOPE_API_KEY",
      "models": [
        { "id": "qwen3-coder-plus", "name": "Qwen3 Coder Plus", "context_window": 131072, "default_max_tokens": 8192 }
      ]
    }
  }
}
```

ModelScope (CN Free Tier):

```json
{
  "$schema": "https://charm.land/crush.json",
  "providers": {
    "modelscope-cn": {
      "type": "openai",
      "name": "ModelScope (CN)",
      "base_url": "https://api-inference.modelscope.cn/v1",
      "api_key": "$MODELSCOPE_API_KEY",
      "models": [
        { "id": "Qwen/Qwen3-Coder-480B-A35B-Instruct", "name": "Qwen3 Coder 480B Instruct", "context_window": 131072, "default_max_tokens": 8192 }
      ]
    }
  }
}
```

OpenRouter (Global):

```json
{
  "$schema": "https://charm.land/crush.json",
  "providers": {
    "openrouter": {
      "type": "openai",
      "name": "OpenRouter",
      "base_url": "https://openrouter.ai/api/v1",
      "api_key": "$OPENROUTER_API_KEY",
      "models": [
        { "id": "qwen/qwen3-coder:free", "name": "Qwen3 Coder (Free)", "context_window": 131072, "default_max_tokens": 8192 }
      ]
    }
  }
}
```

Tip: you can switch the preferred model quickly with `crush models use -t large <provider> <model>`.

If your OpenAI-compatible backend does not implement streaming (SSE), set `disable_stream: true` on the provider. Crush will automatically fall back to a non‑streaming interaction for that provider.

##### vLLM (Local GPU, OpenAI-compatible)

Install vLLM in your Python environment and run the OpenAI-compatible API server:

```
python -m venv .venv
source .venv/bin/activate
pip install --upgrade pip
pip install vllm==0.5.4.post1
python -m vllm.entrypoints.openai.api_server \
  --model meta-llama/Meta-Llama-3-8B-Instruct \
  --port 8000 --host 0.0.0.0 \
  --dtype float16
```

- Optional: set `HF_TOKEN` if the model is gated.
- To persist downloads, export `HF_HOME=$HOME/.cache/huggingface` before launching.

Then add a provider (OpenAI-compatible) to your project `.crush.json` (or use the preset included):

```json
{
  "$schema": "https://charm.land/crush.json",
  "providers": {
    "vllm": {
      "type": "openai",
      "name": "vLLM (local)",
      "base_url": "http://127.0.0.1:8000/v1/",
      "startup_command": "python -m vllm.entrypoints.openai.api_server --model meta-llama/Meta-Llama-3-8B-Instruct --port 8000 --host 0.0.0.0 --dtype float16",
      "startup_timeout_seconds": 90,
      "models": [
        { "id": "meta-llama/Meta-Llama-3-8B-Instruct", "name": "Llama 3 8B Instruct", "context_window": 8192, "default_max_tokens": 1024 }
      ]
    }
  }
}
```

Tip: Switch quickly with `crush models use -t large vllm meta-llama/Meta-Llama-3-8B-Instruct`.

Environment variable `CRUSH_SKIP_PROVIDER_STARTUP=1` disables automatic startup checks if you prefer to manage services yourself.

##### Groq (Hosted, ultra-low latency)

```json
{
  "$schema": "https://charm.land/crush.json",
  "providers": {
    "groq": {
      "type": "openai",
      "name": "Groq",
      "base_url": "https://api.groq.com/openai/v1/",
      "api_key": "$GROQ_API_KEY",
      "models": [
        { "id": "llama3-8b-8192", "name": "Llama3 8B 8192", "context_window": 8192, "default_max_tokens": 4096 }
      ]
    }
  }
}
```

Tip: `crush models use -t large groq llama3-8b-8192`.

##### Together AI (Hosted aggregation)

```json
{
  "$schema": "https://charm.land/crush.json",
  "providers": {
    "together": {
      "type": "openai",
      "name": "Together AI",
      "base_url": "https://api.together.xyz/v1/",
      "api_key": "$TOGETHER_API_KEY",
      "models": [
        { "id": "meta-llama/Llama-3-70B-Instruct-Turbo", "name": "Llama 3 70B Turbo", "context_window": 8000, "default_max_tokens": 4096 }
      ]
    }
  }
}
```

##### Fireworks AI (Hosted, high-throughput)

```json
{
  "$schema": "https://charm.land/crush.json",
  "providers": {
    "fireworks": {
      "type": "openai",
      "name": "Fireworks AI",
      "base_url": "https://api.fireworks.ai/inference/v1/",
      "api_key": "$FIREWORKS_API_KEY",
      "models": [
        { "id": "accounts/fireworks/models/firellava-13b", "name": "FireLLaVa 13B", "context_window": 8192, "default_max_tokens": 2048 }
      ]
    }
  }
}
```

##### OpenAI (Official)

```json
{
  "$schema": "https://charm.land/crush.json",
  "providers": {
    "openai": {
      "type": "openai",
      "name": "OpenAI",
      "base_url": "https://api.openai.com/v1/",
      "api_key": "$OPENAI_API_KEY",
      "models": [
        { "id": "gpt-4o", "name": "GPT-4o", "context_window": 128000, "default_max_tokens": 4096 }
      ]
    }
  }
}
```

##### Azure OpenAI

Azure’s OpenAI-compatible endpoint follows this pattern:

```
https://<resource>.openai.azure.com/openai/deployments/<deployment>/
```

```json
{
  "$schema": "https://charm.land/crush.json",
  "providers": {
    "azure-openai": {
      "type": "openai",
      "name": "Azure OpenAI",
      "base_url": "$AZURE_OPENAI_BASE_URL",
      "api_key": "$AZURE_OPENAI_KEY",
      "extra_headers": {
        "api-version": "2024-05-01-preview"
      },
      "models": [
        { "id": "gpt-4o", "name": "GPT-4o Deployment", "context_window": 128000, "default_max_tokens": 4096 }
      ]
    }
  }
}
```

Set `AZURE_OPENAI_BASE_URL` to your deployment URL (for example `https://my-resource.openai.azure.com/openai/deployments/gpt-4o/`).

#### Anthropic-Compatible APIs

Custom Anthropic-compatible providers follow this format:

```json
{
  "$schema": "https://charm.land/crush.json",
  "providers": {
    "custom-anthropic": {
      "type": "anthropic",
      "base_url": "https://api.anthropic.com/v1",
      "api_key": "$ANTHROPIC_API_KEY",
      "extra_headers": {
        "anthropic-version": "2023-06-01"
      },
      "models": [
        {
          "id": "claude-sonnet-4-20250514",
          "name": "Claude Sonnet 4",
          "cost_per_1m_in": 3,
          "cost_per_1m_out": 15,
          "cost_per_1m_in_cached": 3.75,
          "cost_per_1m_out_cached": 0.3,
          "context_window": 200000,
          "default_max_tokens": 50000,
          "can_reason": true,
          "supports_attachments": true
        }
      ]
    }
  }
}
```

### Amazon Bedrock

Crush currently supports running Anthropic models through Bedrock, with caching disabled.

- A Bedrock provider will appear once you have AWS configured, i.e. `aws configure`
- Crush also expects the `AWS_REGION` or `AWS_DEFAULT_REGION` to be set
- To use a specific AWS profile set `AWS_PROFILE` in your environment, i.e. `AWS_PROFILE=myprofile crush`

### Vertex AI Platform

Vertex AI will appear in the list of available providers when `VERTEXAI_PROJECT` and `VERTEXAI_LOCATION` are set. You will also need to be authenticated:

```bash
gcloud auth application-default login
```

To add specific models to the configuration, configure as such:

```json
{
  "$schema": "https://charm.land/crush.json",
  "providers": {
    "vertexai": {
      "models": [
        {
          "id": "claude-sonnet-4@20250514",
          "name": "VertexAI Sonnet 4",
          "cost_per_1m_in": 3,
          "cost_per_1m_out": 15,
          "cost_per_1m_in_cached": 3.75,
          "cost_per_1m_out_cached": 0.3,
          "context_window": 200000,
          "default_max_tokens": 50000,
          "can_reason": true,
          "supports_attachments": true
        }
      ]
    }
  }
}
```

## Logging

Sometimes you need to look at logs. Luckily, Crush logs all sorts of
stuff. Logs are stored in `./.crush/logs/crush.log` relative to the project.

The CLI also contains some helper commands to make perusing recent logs easier:

```bash
# Print the last 1000 lines
crush logs

# Print the last 500 lines
crush logs --tail 500

# Follow logs in real time
crush logs --follow
```

Want more logging? Run `crush` with the `--debug` flag, or enable it in the
config:

```json
{
  "$schema": "https://charm.land/crush.json",
  "options": {
    "debug": true,
    "debug_lsp": true
  }
}
```

## Disabling Provider Auto-Updates

By default, Crush automatically checks for the latest and greatest list of
providers and models from [Catwalk](https://github.com/charmbracelet/catwalk),
the open source Crush provider database. This means that when new providers and
models are available, or when model metadata changes, Crush automatically
updates your local configuration.

For those with restricted internet access, or those who prefer to work in
air-gapped environments, this might not be want you want, and this feature can
be disabled.

To disable automatic provider updates, set `disable_provider_auto_update` into
your `crush.json` config:

```json
{
  "$schema": "https://charm.land/crush.json",
  "options": {
    "disable_provider_auto_update": true
  }
}
```

Or set the `CRUSH_DISABLE_PROVIDER_AUTO_UPDATE` environment variable:

```bash
export CRUSH_DISABLE_PROVIDER_AUTO_UPDATE=1
```

### Manually updating providers

Manually updating providers is possible with the `crush update-providers`
command:

```bash
# Update providers remotely from Catwalk.
crush update-providers

# Update providers from a custom Catwalk base URL.
crush update-providers https://example.com/

# Update providers from a local file.
crush update-providers /path/to/local-providers.json

# Reset providers to the embedded version, embedded at crush at build time.
crush update-providers embedded

# For more info:
crush update-providers --help
```

## A Note on Claude Max and GitHub Copilot

Crush only supports model providers through official, compliant APIs. We do not
support or endorse any methods that rely on personal Claude Max and GitHub
Copilot accounts or OAuth workarounds, which violate Anthropic and
Microsoft’s Terms of Service.

We’re committed to building sustainable, trusted integrations with model
providers. If you’re a provider interested in working with us,
[reach out](mailto:vt100@charm.sh).

## Whatcha think?

We’d love to hear your thoughts on this project. Need help? We gotchu. You can find us on:

- [Twitter](https://twitter.com/charmcli)
- [Discord][discord]
- [Slack](https://charm.land/slack)
- [The Fediverse](https://mastodon.social/@charmcli)
- [Bluesky](https://bsky.app/profile/charm.land)

[discord]: https://charm.land/discord

## License

[FSL-1.1-MIT](https://github.com/charmbracelet/crush/raw/main/LICENSE.md)

---

Part of [Charm](https://charm.land).

<a href="https://charm.land/"><img alt="The Charm logo" width="400" src="https://stuff.charm.sh/charm-banner-next.jpg" /></a>

<!--prettier-ignore-->
Charm热爱开源 • Charm loves open source
