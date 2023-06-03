# go-cdn

> An experimental CDN project in Go.
> A lightweight server that allows you to upload and download files via HTTP. It provides optional authentication, input validation, error handling, and verbose logging. The server is implemented in Go and can be easily configured using a config.json file.

## Features

- Serve static files securely over HTTP
- Basic authentication support
- File upload & download functionality

### Error Handling

The server is equipped with error handling to handle various scenarios, including invalid requests, authentication failures, and payload size exceeding the maximum allowed limit. If an error occurs, the server will return an appropriate HTTP status code along with an error message.
Logging

The server uses the logrus library for logging. Verbose logging is enabled by default and displays detailed information about incoming requests, errors, and server restarts. The logs are printed to the console.
Customization

The server implementation provided is a starting point and may require customization based on your specific requirements. You can modify the handleGet and handlePost functions in the CDNHandler struct to implement your desired file upload and download logic.

## Possible use cases

The CDN can be used in various scenarios, such as:

- Hosting static assets (i.e images or Javascript files) for web applications
- Distributing files securely to authorized users
- Building a personal or small-scale CDN for content delivery

## Usage

### Prerequisites

- Go programming language (version 1.16 or later)

### Installation

1. Clone the repository:

```bash

git clone https://github.com/notashelf/go-cdn.git
```

2. Change to the project directory:

```bash
cd your-repo
```

3. Install the dependencies:

```bash
go get github.com/sirupsen/logrus
go get github.com/pkg/errors
```

### Configuration

The server can be configured using the config.json file. Create a file named config.json in the same directory as the main file (main.go) and configure the following properties:

- port (string): The port on which the server will listen for incoming connections.
- max_upload_size (integer): The maximum allowed size of file uploads, in bytes.
- heartbeat (string): The duration after which the server will automatically restart. Specify a value in the format "5m" for 5 minutes, "1h" for 1 hour, etc. Set to "0" to disable automatic restarts.
- require_auth (boolean): Whether to require authentication for file uploads and downloads.
- auth_username (string): The username for authentication (only applicable if require_auth is set to true).
- auth_password (string): The password for authentication (only applicable if require_auth is set to true).

Example config.json file:

```json
{
  "port": "8080",
  "max_upload_size": 10485760,
  "heartbeat": "1h",
  "require_auth": true,
  "auth_username": "your-username",
  "auth_password": "your-password"
}
```

### Usage

Start the CDN server by running the following command:

```bash
go run main.go -config config.json
```

The server will start and listen on the specified port. Verbose log messages will be displayed in the console.

#### Uploading a file:

Use the curl command to upload a file to the CDN server:

```bash
curl -X POST -F "file=@/path/to/file" http://localhost:8080
```

_Replace /path/to/file with the actual path to the file you want to upload. If authentication is enabled, provide the username and password when prompted._

#### Downloading a file:

Use the curl command to download a file from the CDN server:

```bash
curl -O http://localhost:8080/<filename>
```

_Replace <filename> with the name of the file you want to download._

### Authenticated Upload and Download

To perform an authenticated upload or download, you can use the following curl commands:

#### Uploading a File:

```bash
curl -X POST -F "file=@/path/to/file" -u "your-username:your-password" http://localhost:8080
```

_Replace /path/to/file with the actual path to the file you want to upload. The -u flag is used to provide the authentication credentials._

#### Downloading a File:

```bash
curl -O -u "your-username:your-password" http://localhost:8080/<filename>
```

Replace <filename> with the name of the file you want to download. The -u flag is used to provide the authentication credentials.

**Please note that these examples assume you're running the server on localhost with the specified port and authentication credentials. Make sure to adjust the hostname and port accordingly.**

# License

This project is licensed under the GPL3 License. See the LICENSE file for details.
