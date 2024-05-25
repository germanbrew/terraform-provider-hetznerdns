# Releasing and Publishing Provider

A release of this provider includes publishing the new version
at the Terraform registry. Therefore, a file containing hashes
of the binaries must be singed with my private GPG key.

Create a new tag and push it to GitHub. This starts a workflow that creates a release.

```bash
$ git tag v1.0.17
$ git push --tags
```

