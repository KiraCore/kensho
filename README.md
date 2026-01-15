# KENSHO

### Building
1. **via Dockerfile**:
   ```
   docker build -t kensho-builder . &&  docker run --rm -v ./tmp/:/output kensho-builder
   ```

2. **build natively**
   ```
   go build main.go
   ```

3. **with fyne-cross**
   ```
   fyne-cross linux
   ```

[For more information how to build with fyne-cross check official documentation](https://docs.fyne.io/started/cross-compiling.html)

### Preparing Your Pull Request

1. **Fork and Clone**: First, fork the main repository and clone it locally.
   ```bash
   git clone https://github.com/your-username/repository-name.git
   cd repository-name


2. Create a New Branch: Always create a branch from an up-to-date master branch for your changes.

```bash
git checkout master
git pull
git checkout -b your-feature-branch
```

3. Make Changes: Make the necessary modifications or additions to the codebase and commit your changes locally. Use meaningful commit messages that clearly describe the changes.

### Using Semantic Commit Messages
To automate the version management of our project, we use mathieudutour/github-tag-action which relies on semantic commit messages to determine the type of version bump (major, minor, or patch). Here's how to format your commit messages:

* Patch update (fix:): Includes bug fixes, minor changes, or any non-breaking changes that fix incorrect behavior.

```bash
git commit -m "fix: resolve an issue where..."
```

* Minor update (feat:): Adds new features or updates that do not break backward compatibility.

```bash
git commit -m "feat: add support for..."
```

* Major update (BREAKING CHANGE): Introduces changes that break backward compatibility with the previous versions.

```bash
git commit -m "feat: change database schema"
-m "BREAKING CHANGE: database migration required for..."
```

1. Push Your Changes: Once your changes are committed, push the branch to your fork.

```bash
git push origin your-feature-branch
```

2. Create a Pull Request: Go to the repository on GitHub. You'll often find a button to create a pull request from your recently pushed branches.

3. Describe Your Changes: Provide a clear, detailed description of your changes when opening your pull request.
