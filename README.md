# HTTP Loggable

A fast log ingestion and viewing service with a server written in Go and a modern UI built with HTML, JavaScript, and [Shoelace components](https://shoelace.style).

## Features

### Log Ingestion Server
- Fast HTTP server for log ingestion
- Automatic log file creation with ISO timestamps
- Support for JSON and raw request body logging
- Base64 encoding option for request bodies
- Request metadata logging (method, URL, headers)

### Log Viewer UI
- Modern, responsive interface using Shoelace.style components
- Real-time log file listing with pretty date formatting
- Advanced search functionality with debouncing
- Pagination support with configurable page size
- Copy to clipboard functionality
- Dialog view for large log entries
- AM/PM time format display

## Prerequisites

- Go 1.16 or higher
- Modern web browser (Chrome, Firefox, Safari, Edge)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/http-loggable.git
cd http-loggable
```

2. Build the project:
```bash
make build
```

This will create the following structure in the `build` directory:
```
build/
├── bin/
│   ├── read    # Log viewer server
│   └── write   # Log ingestion server
└── public/     # Static files for the UI
```

## Usage

### Starting the Servers

1. Start the log ingestion server:
```bash
make write
```
The server will start on port 8081 and create log files in the `logs` directory.

2. Start the log viewer server:
```bash
make read
```
The viewer will be available at http://localhost:8082

### Environment

- `INCLUDE_BODY_IN_JSON`: When set to true, body will be included in the json log written to log file. Defaults to true.
- `BODY_AS_BASE64`: When set to true, body bytes will be written as base64 string. Defaults to true.

### API Endpoints

#### Log Ingestion Server (Port 8081)
- `POST /` - Log a request
  - Request with url, method, body will be logged

#### Log Viewer Server (Port 8082)
- `GET /api/logs` - List all log files
- `GET /api/logs/search` - Search log entries
  - Query parameters:
    - `file`: Log file name (required)
    - `q`: Search term (optional)
    - `page`: Page number (default: 1)
    - `page_size`: Entries per page (default: 10)

### UI Features

1. **File Selection**
   - Dropdown shows log files with formatted dates
   - Files are sorted by timestamp

2. **Search**
   - Real-time search with 300ms debounce
   - Searches through both request bodies and raw entries

3. **Pagination**
   - Configurable entries per page (10, 25, 50, 100)
   - Previous/Next navigation
   - Page count and entry range display

4. **Log Entry View**
   - Request number and method display
   - URL and headers information
   - Formatted request body
   - Copy to clipboard functionality
   - Dialog view for large entries (>250 characters)

## Development

### Project Structure
```
.
├── bin/
│   ├── read/      # Log viewer server
│   └── write/     # Log ingestion server
├── public/        # Static files
│   └── index.html # Main UI
├── logs/          # Log files directory
└── Makefile       # Build and run commands
```

### Available Make Commands
- `make build` - Build both servers and copy static files
- `make read` - Run the log viewer server
- `make write` - Run the log ingestion server
- `make clean` - Remove build directory
- `make delete` - Remove logs directory

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.


