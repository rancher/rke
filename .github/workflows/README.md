# Description of GitHub Actions in this repository

## Stale (`stale.yml`)

This workflow will label and eventually close issues and pull requests based on the configuration in this file.

## Go Generate (`git-actions-go-generate.yml`)

In RKE, `go generate` is used to:

* Generate deepcopy functions
* Download the latest published `data.json` from rancher/kontainer-driver-metadata (see `codegen/codegen.go`)
* Converts the downloaded `data.json` in to Go source code using go-bindata

See the comments on top of `main.go` to see what is being executed when you run `go generate`.

This was done manually every time a PR was merged in rancher/kontainer-driver-metadata, and it was also usually executed on the machine of the user. The first step to tackle this was to create a make target for `go generate`, called `go-generate`, so everyone can use `make go-generate` to consistently run the command with the same environment.

Second improvement was to wrap this `make go-generate` into a GitHub Actions workflow (this one), so it can be;

* Triggered via the UI (no more local machine needed)
* Automatically triggered via API

In rancher/kontainer-driver-metadata, we added a `dispatch` step in `.drone.yml`, which will execute this GitHub Actions workflow after a PR is merged. The user that merged the PR will also be mentioned and assigned in the PR that will be created when this workflow has finished.

## Update README (`update-readme.yml`)

This workflow will update the README with latest versions retrieved from tags in the GitHub repository. It uses the file `README-template.md` as source, and will add the latest versions into this file to end up with the actual `README.md` file.
