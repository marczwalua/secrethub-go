# Go SecretHub

[![GoDoc](http://img.shields.io/badge/godoc-reference-blue.svg)][godoc]
[![GolangCI](https://golangci.com/badges/github.com/secrethub/secrethub-go.svg)][golang-ci]

The official [SecretHub][secrethub] Go client library.

> SecretHub is a developer tool to help you keep database passwords, API tokens, and other secrets out of IT automation scripts. 

## Installation

Install secrethub-go with:

```sh
go get -u github.com/secrethub/secrethub-go
```

Or install a specific version with:

```sh
go get -u github.com/secrethub/secrethub-go@vX.Y.Z
```

Then, import it using:

``` go
import (
    "github.com/secrethub/secrethub-go/pkg/secrethub"
)
```

## Documentation

For details on all functionality of this library, see the [GoDoc][godoc] documentation.

Below are a few simple examples:

```go
import (
	"github.com/secrethub/secrethub-go/pkg/randchar"
	"github.com/secrethub/secrethub-go/pkg/secrethub"
)

// Setup
credential, err := secrethub.NewCredential("<your credential>", "<passphrase>")
client := secrethub.NewClient(credential, nil)

// Write
secret, err := client.Secrets().Write("path/to/secret", []byte("password123"))

// Read
secret, err = client.Secrets().Versions().GetWithData("path/to/secret:latest")
fmt.Println(secret.Data) // prints password123

// Generate a slice of 32 alphanumeric characters.
data, err := randchar.NewGenerator(false).Generate(32) 
secret, err = client.Secrets().Write("path/to/secret", data)
```

Note that only packages inside the `/pkg` directory should be considered library code that you can use in your projects. All other code is not guaranteed to be backwards compatible and may change in the future.  

## Development

Pull requests from the community are welcome.
If you'd like to contribute, please checkout [the contributing guidelines](./CONTRIBUTING.md).

## Test

Run all tests:

    make test

Run tests for one package:

    go test ./pkg/secrethub

Run a single test:

    go test ./pkg/secrethub -run TestSignup

For any requests, bug or comments, please [open an issue][issues] or [submit a
pull request][pulls].

[secrethub]: https://secrethub.io
[issues]: https://github.com/secrethub/secrethub-go/issues/new
[pulls]: https://github.com/secrethub/secrethub-go/pulls
[godoc]: http://godoc.org/github.com/secrethub/secrethub-go
[golang-ci]: https://golangci.com/r/github.com/secrethub/secrethub-go
