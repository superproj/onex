Welcome to the Onex! We’re thrilled to have a diverse group of developers contribute to our project. To ensure a smooth collaboration process, we kindly ask you to review this guide before submitting an issue or a pull request.

### Reporting Bugs or Suggesting Fixes

For bug reporting and fixes, we utilize GitHub issues. Before making a submission, please:

- Conduct a search for existing issues and pull requests.
- Review our [FAQ](https://onex.com/docs/intro/faq) to see if your question has already been addressed.

When reporting a bug, follow the issue template provided to detail the encountered issue and steps to reproduce it. If possible, include a minimal repository that replicates the problem.

### Proposing New Features

To ensure that new features reflect the broader community’s needs, we follow a proposal process. This process comprises three stages: proposal, feature discussion, and pull request (PR), with each stage taking the form of a GitHub issue for clarity and traceability.

1. **Proposal Submission**: Describe the feature’s functional requirements and any relevant references or literature in detail.
2. **Feature Discussion**: Once the community shows support for a proposal, a detailed feature issue will be opened to discuss implementation methods and demonstrations.
3. **Pull Request**: Following agreement on the feature, a PR will be initiated to implement the function, linking back to both the proposal and feature issues.

After successful merge, all related issues will be closed.

### How to Submit Code

New to GitHub? Follow these steps to contribute your code:

1. Fork the repository to your GitHub account.
2. Create a new feature branch from the main branch, naming it appropriately (e.g., `feature/log`).
3. Develop your feature.
4. Push your code to the branch.
5. Submit a PR on GitHub.
6. Await review and potential merge into the main branch.

**Note**: Ensure your code adheres to our coding standards and includes complete test cases. Linking relevant issues in your PR description can significantly aid the review process.

### Conventional Commits

We adhere to the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/#summary) standard, which structures commit messages as follows:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### Commit Types

- **Main Types**: `fix`, `feat`, `deps`, `break`
- **Other Types**: `docs`, `refactor`, `style`, `test`, `chore`, `ci`

#### Scopes

Our project defines scopes such as `uc`, `apiserver`, and `gw`, among others, to categorize changes more specifically.

#### Descriptions and Bodies

Commit messages should be clear and to the point, using present tense and avoiding capitalization and punctuation at the end.

#### Footers

Use the footer to note any breaking changes or to reference issues that the commit addresses.

### Release Notes

Use [git-chglog](https://github.com/git-chglog/git-chglog) to generate changelogs, capturing categories like Breaking Changes, Dependencies, Bug Fixes, and more.
