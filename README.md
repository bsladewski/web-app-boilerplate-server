# Web App Boilerplate Server

![go workflow](https://github.com/bsladewski/web-app-boilerplate-server/workflows/Go/badge.svg)

The web app boilerplate provides the basic structure of a web app that uses Go for the API server and VueJS for the front-end. The purpose of this repository is to speed up the development of prototypes by implementing some of the basic functionality that is common to many different types of web apps. This includes packages for working with the environment, database, application router, etc. as well as functionality for managing user accounts, authentication, and permissions.

For an example project utilizing this boilerplate check out the [Mojito Repository](https://github.com/bsladewski/mojito).

Feel free to [fork](https://github.com/bsladewski/web-app-boilerplate-server/fork), copy, or borrow from this repository and start building!

# Dependencies

This project uses the [Go programming language](https://golang.org/dl/).

Additional dependencies are managed through [Go Modules](https://blog.golang.org/using-go-modules).

# Usage

## Installation

To get started, retrieve the package using the `go get` command:

`$ go get github.com/bsladewski/web-app-boilerplate-server`

Alternatively you may clone the repository directly:

`$ git clone https://github.com/bsladewski/web-app-boilerplate-server`

## Running Without Docker

Build the application by running the `go build` command in the root project directory:

`$ go build`

This will produce an executable binary:

`$ ./web-app-boilerplate-server`

The application is configured through the environement. To stand up an API server ensure that all required environment variables are set to appropriate values. You can find a sample configuration in the `.env.sample` file. Documentation for the environment variables are found in their respective package documentation.

## Running With Docker

To begin, copy the `.env.sample` file to `.env`. You may use this file to configure the API server.

Build the docker image by running the following command:

`$ docker build --tag webapp:1.0 .`

Once the docker image is built, run the application using the `docker run` command:

`$ docker run --publish 8080:8080 --env-file .env --name webapp webapp:1.0`

We pass the `.env` file into the `docker run` command to configure the API server.

To stop the API server use the `docker stop` command:

`$ docker stop webapp`

Finally, you may remove the container with the following command:

`$ docker rm --force webapp`

# Contributing

1. [Fork it!](https://github.com/bsladewski/web-app-boilerplate-server/fork)
2. Create your feature branch: `git checkout -b feature/my-new-feature`
3. Commit your changes: `git commit -am 'Implemented my cool new feature'`
4. Push to the branch: `git push origin feature/my-new-feature`
5. Submit a new Pull Request

# License

The MIT License (MIT)

Copyright (c) 2021 Benjamin Sladewski

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
