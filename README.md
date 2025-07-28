# policy-bot tests

This tool can be used to write tests against a [policy-bot](https://github.com/palantir/policy-bot) policy file to validate intent.

## Example

Given this [.policy.yml](./tests/.policy.yml) and this [.policy-tests.yml](./tests/.policy-tests.yml)
you'll get the following test output:

```sh
❯ policy-bot-tests
Found 4 test case(s)
--- Running Test: Pass policy when team alpha files change and team alpha approves ---
  - Evaluation Tree:
    - ✅ policy: All rules are approved
      - ✅ approval: All rules are approved
        - ✅ team-alpha-review: Approved by alpha-bob
        - 💤 team-beta-review: No changed files match the required patterns
      - 💤 disapproval: No disapproval policy is specified or the policy is empty
PASS
--- Running Test: Fail policy when team alpha files change and only PR author approves ---
  - Evaluation Tree:
    - 🟡 policy: 0/1 rules approved
      - 🟡 approval: 0/1 rules approved
        - 🟡 team-alpha-review: 0/1 required approvals. Ignored 1 approval from disqualified users
        - 💤 team-beta-review: No changed files match the required patterns
      - 💤 disapproval: No disapproval policy is specified or the policy is empty
PASS
--- Running Test: Pass policy when multiple team files are changing with multiple team approvals ---
  - Evaluation Tree:
    - ✅ policy: All rules are approved
      - ✅ approval: All rules are approved
        - ✅ team-alpha-review: Approved by alpha-bob
        - ✅ team-beta-review: Approved by beta-charlie
      - 💤 disapproval: No disapproval policy is specified or the policy is empty
PASS
--- Running Test: Fail policy when multiple team files are changing with review missing from beta team ---
  - Evaluation Tree:
    - 🟡 policy: 1/2 rules approved
      - 🟡 approval: 1/2 rules approved
        - ✅ team-alpha-review: Approved by alpha-bob
        - 🟡 team-beta-review: 0/1 required approvals. Ignored 1 approval from disqualified users
      - 💤 disapproval: No disapproval policy is specified or the policy is empty
PASS

--- Summary ---
4 / 4 tests passed.
```

## Installation

### Manual Installation

Download the latest release from [GitHub Releases](https://github.com/reegnz/policy-bot-tests/releases) and extract the binary for your platform.

## Development

### Building from source

```bash
git clone https://github.com/reegnz/policy-bot-tests.git
cd policy-bot-tests
go build -o policy-bot-tests .
```

### Running tests

```bash
go test -v ./...
```

### Local GoReleaser testing

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Test build locally
goreleaser build --snapshot --clean

# Test release locally (dry run)
goreleaser release --snapshot --clean
```
