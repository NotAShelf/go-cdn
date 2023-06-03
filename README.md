# go-cdn

An experimental CDN project in Go.

## Features

- Serve static files securely over HTTP
- Basic authentication support
- File upload functionality

## Possible use cases

The CDN can be used in various scenarios, such as:

- Hosting static assets (i.e images or Javascript files) for web applications
- Distributing files securely to authorized users
- Building a personal or small-scale CDN for content delivery

## Usage

### Starting the CDN

1. Build the program

```go
go build
```

2. Run the built binary

```go

```

### Using the CDN

To request a file from the CDN, use the following URL format:

```bash
http://your-cdn-server:port/file-path
```

> Replace your-cdn-server with the hostname or IP address of your CDN server and port with the port number on which the server is running. Append the desired file path after the hostname and port.

**Example:**

- `http://localhost:8080/images/file_you_have_uploaded.png`

  - If the file exists and the request is authorized, the file will be served by the CDN.
  - If the file doesn't exist, a 404 Not Found response will be returned.

#### Uploading Files to the CDN

> Send a POST request to the /upload endpoint of the CDN server.

Example using cURL:

```bash
curl -X POST -u username:password -F "file=@/path/to/file" http://your-cdn-server:port/upload
```

> Replace username and password with the authentication credentials you have set in the main.go file. Replace your-cdn-server with the hostname or IP address of your CDN server and port with the port number on which the server is running. Provide the file path after the @ symbol in the -F parameter.

**Example:**

```bash
curl -X POST -u admin:password -F "file=@/absolute/path/to/image.jpg" http://localhost:8080/upload
```

    - The uploaded file will be saved in the specified uploadPath directory.

> Note: The server responds with a success message if the file upload is successful.

### Security Considerations

It is highly recommended to use SSL/TLS encryption (HTTPS) for secure communication between clients and the CDN server.
Change the default username and password in the main.go file to strong and secure credentials.
Consider implementing additional security measures based on your specific requirements.

# License

This project is licensed under the GPL3 License. See the LICENSE file for details.
