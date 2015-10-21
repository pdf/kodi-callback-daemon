# Contributing

To hack on this project:

1. Install as usual (`go get -u github.com/path/repo`)
2. Install dependencies using [glide](https://github.com/Masterminds/glide) (`glide up`)
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Commit your changes (`git commit -am 'Add some feature'`)
5. Build using go version 1.5+ with vendoring enabled (`GO15VENDOREXPERIMENT=1 go install`)

Contribute upstream:

1. Fork it on GitHub
2. Add your remote (`git remote add fork git@github.com:mycompany/repo.git`)
3. Push to the branch (`git push -u fork my-new-feature`)
4. Create a new Pull Request on GitHub

For other team members:

1. Install as usual (`go get -u github.com/path/repo`)
2. Add your remote (`git remote add fork git@github.com:mycompany/repo.git`)
3. Pull your revisions (`git fetch fork; git checkout -b my-new-feature fork/my-new-feature`)

Notice: Always use the original import path by installing with `go get`.
