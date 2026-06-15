# Security Policy

## Reporting a Vulnerability

If you believe you have found a security vulnerability in `betterado`, please
report it privately rather than opening a public issue.

- Preferred: open a [private security advisory](https://github.com/parsoFish/terraform-provider-betterado/security/advisories/new)
  on the repository.
- Include a description of the issue, the affected resource(s) or data source(s),
  and step-by-step instructions to reproduce.

You will receive an acknowledgement once the report is triaged. Please allow a
reasonable amount of time for the issue to be addressed before any public
disclosure. This project follows the principle of
[Coordinated Vulnerability Disclosure](https://en.wikipedia.org/wiki/Coordinated_vulnerability_disclosure).

## Scope

`betterado` authenticates to the Azure DevOps REST API using a Personal Access
Token (or service-principal / OIDC / managed-identity credentials). Report issues
in the provider here. Vulnerabilities in **Azure DevOps itself** should be reported
to Microsoft via the [Microsoft Security Response Center](https://msrc.microsoft.com/),
not to this repository.
