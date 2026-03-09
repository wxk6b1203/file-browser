commit message should be in the format of:

```
<type>: <subject>
```

Where `<type>` is one of the following:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, missing semicolons, etc.)
- `refactor`: Code refactoring without changing functionality
- `test`: Adding or updating tests
- `chore`: Changes to the build process or auxiliary tools and libraries
- `perf`: Performance improvements
- `ci`: Changes to CI configuration files and scripts
- `build`: Changes that affect the build system or external dependencies
- `revert`: Reverts a previous commit
- `wip`: Work in progress (not ready for review or merge)
- `release`: Commits related to release preparation (version bump, changelog update, etc.)

The `<subject>` should be a concise description of the change, ideally less than 100 characters. It should be written in the imperative mood (e.g., "Add feature" instead of "Added feature" or "Adds feature").