
## Contribute

Read this file and feel free :-)

### Dockerized

With docker, you do not need to install `go`  and `mongodb` on your local machine, it's easy to setup a development environment for this repository. Thanks to @centsent who helped to dockerize this project.

### Prerequisites

Make sure you already installed below programs on your local machine:

* `git`
* `docker`
* `docker-compose`

### dep

[dep](https://github.com/golang/dep) is a prototype dependency management tool for Go.
To use `dep` in the container, prefix `make` for all `dep` commands, for example:

```bash
$ make dep "ensure -add github.com/some/repos"
```

Beware that the double quotes are required after `make dep` command.

### Making Changes

Puppet has some good and simple [contribution rules](https://github.com/puppetlabs/puppet/blob/master/CONTRIBUTING.md#making-changes) so lets adopt them in an adjusted way.

* Create a topic branch from where you want to base your work.
  * This is usually the master branch.
  * To quickly create a topic branch based on master, run `git checkout -b
    fix/master/my_contribution master`. Please avoid working directly on the
    `master` branch.
* Make commits of logical and atomic units.
* Check for unnecessary whitespace with `git diff --check` before committing.
* Make sure your commit messages are in a proper format:
  ```
  chore: add Oyster build script
  docs: explain hat wobble
  feat: add beta sequence
  fix: remove broken confirmation message
  refactor: share logic between 4d3d3d3 and flarhgunnstow
  style: convert tabs to spaces
  test: ensure Tayne retains clothing
  ```
   [(Source)](https://gist.githubusercontent.com/mutewinter/9648651/raw/77f8abd031d02f822543992a86bf4c1fc50ad760/commit_format_examples.txt)
* Make sure you have added the necessary tests for your changes.
* Run _all_ the tests with `make test` to assure nothing else was accidentally broken (it will build the docker container and run `go test`)