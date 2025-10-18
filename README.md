# Video Transcoder API

A high-performance video transcoding service built with Go, Gin, GORM, and FFmpeg. This service allows you to upload videos and automatically transcode them into multiple qualities (480p, 720p, 1080p) for adaptive streaming and various device compatibility.

## Features

- **Video Upload**: Upload videos in various formats (MP4, AVI, MOV, MKV, etc.)
- **Multiple Quality Transcoding**: Automatically transcode to 480p, 720p, and 1080p
- **Background Processing**: Non-blocking video processing with worker queues
- **Progress Tracking**: Real-time status updates and progress monitoring
- **RESTful API**: Clean REST endpoints for integration
- **Database Storage**: PostgreSQL with GORM for robust data management
- **Docker Support**: Containerized deployment with Docker Compose

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Upload API    │    │   Status API    │    │  Download API   │
│   (Handler)     │    │   (Handler)     │    │   (Handler)     │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          ▼                      ▼                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Controller Layer                             │
│               (Business Logic & Orchestration)                  │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Service Layer                                │
│              (FFmpeg Integration & Processing)                  │
└─────────────────────┬───────────────────────────────────────────┘
                      │
          ┌───────────┼───────────┐
          ▼           ▼           ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│   Worker    │ │  Database   │ │File Storage │
│  (Queue)    │ │  (GORM)     │ │  (Local)    │
└─────────────┘ └─────────────┘ └─────────────┘
```

## Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- FFmpeg installed on your system
- Docker & Docker Compose (optional)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd video_transcoder
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up PostgreSQL database**
   ```bash
   createdb video_transcoder
   ```

4. **Configure environment variables** (optional)
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=saikumar
   export DB_PASSWORD=saikumar
   export DB_NAME=video_transcoder
   ```

5. **Run the application**
   ```bash
   go run main.go
   ```

### Docker Setup (Recommended)

1. **Start with Docker Compose**
   ```bash
   docker-compose up -d
   ```

This will start:
- PostgreSQL database on port 5432
- Video Transcoder API on port 8080

## API Endpoints

### Base URL
```
http://localhost:8080/api/v1
```

### 1. Upload Video
```http
POST /upload
Content-Type: multipart/form-data

Form data:
- video: (file) The video file to upload
```

**Response:**
```json
{
    "id": "uuid-string",
    "message": "Video uploaded successfully. Transcoding started."
}
```

### 2. Get Video Status
```http
GET /videos/:id/status
```

**Response:**
```json
{
    "id": "uuid-string",
    "status": "completed",
    "original_path": "/path/to/original.mp4",
    "transcoded_paths": {
        "480p": "/path/to/480p.mp4",
        "720p": "/path/to/720p.mp4",
        "1080p": "/path/to/1080p.mp4"
    },
    "duration": 120.5,
    "file_size": 52428800,
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:05:00Z",
    "jobs": [
        {
            "id": 1,
            "quality": "720p",
            "status": "completed",
            "progress": 100
        }
    ]
}
```

### 3. Download Video
```http
GET /videos/:id/download?quality=720p
```

Query parameters:
- `quality`: Optional. Options: `original`, `480p`, `720p`, `1080p`
- If not specified, returns the original file

### 4. List Videos
```http
GET /videos?page=1&limit=10
```

**Response:**
```json
{
    "videos": [...],
    "page": "1",
    "limit": "10"
}
```

## Status Values

- `pending`: Video uploaded, waiting to start transcoding
- `processing`: Transcoding in progress
- `completed`: All transcoding jobs completed successfully
- `failed`: Transcoding failed

## Supported Video Formats

**Input formats:**
- MP4, AVI, MOV, MKV, WMV, FLV, WebM, M4V, 3GP, MPG, MPEG

**Output format:**
- MP4 with H.264 video codec and AAC audio codec

## Quality Presets

| Quality | Resolution | Video Bitrate | Audio Bitrate |
|---------|------------|---------------|---------------|
| 480p    | 854x480    | 1000k         | 128k          |
| 720p    | 1280x720   | 2500k         | 128k          |
| 1080p   | 1920x1080  | 5000k         | 128k          |

## Development

### Project Structure

```
video_transcoder/
├── main.go                 # Application entry point
├── config/
│   └── database.go         # Database configuration
├── models/
│   └── transcoder.go       # Data models (Video, TranscodeJob)
├── handlers/
│   └── transcoder.go       # HTTP request handlers
├── controller/
│   └── transcoder.go       # Business logic controller
├── services/
│   └── transcoder.go       # FFmpeg integration service
├── worker/
│   └── worker.go           # Background job processor
├── utils/
│   └── helpers.go          # Utility functions
├── Dockerfile              # Docker container definition
├── docker-compose.yml      # Multi-container setup
└── README.md
```

### Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

### Building

```bash
# Build for current platform
go build -o video-transcoder

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o video-transcoder-linux

# Build Docker image
docker build -t video-transcoder .
```

## Configuration

### Environment Variables

| Variable    | Default          | Description                |
|-------------|------------------|----------------------------|
| DB_HOST     | localhost        | PostgreSQL host            |
| DB_PORT     | 5432             | PostgreSQL port            |
| DB_USER     | <db username>    | PostgreSQL username        |
| DB_PASSWORD | <db password>    | PostgreSQL password        |
| DB_NAME     | video_transcoder | PostgreSQL database name   |
| DB_SSLMODE  | disable          | PostgreSQL SSL mode        |
| GIN_MODE    | debug            | Gin framework mode         |

## Monitoring & Logs

The application provides comprehensive logging for:
- Video upload events
- Transcoding job progress
- Error handling and debugging
- Database operations

Logs are output to stdout and can be collected by Docker logging drivers.

## Performance Considerations

- **Concurrent Processing**: Worker processes multiple jobs simultaneously
- **Connection Pooling**: Database connections are pooled for efficiency
- **Background Processing**: Video transcoding doesn't block API responses
- **Storage Optimization**: Configurable quality presets to balance quality vs. file size

## Troubleshooting

### Common Issues

1. **FFmpeg not found**
   ```bash
   # Install FFmpeg on macOS
   brew install ffmpeg
   
   # Install FFmpeg on Ubuntu
   sudo apt update && sudo apt install ffmpeg
   ```

2. **Database connection failed**
   - Verify PostgreSQL is running
   - Check connection credentials
   - Ensure database exists

3. **Out of disk space**
   - Monitor available storage
   - Implement cleanup policies for old files
   - Use external storage solutions for production

### Health Check

```http
GET /
```

Returns server status and basic information.

## Production Deployment

### Recommendations

1. **Use external PostgreSQL** (AWS RDS, Google Cloud SQL, etc.)
2. **Implement file storage** (AWS S3, Google Cloud Storage, etc.)
3. **Add monitoring** (Prometheus, Grafana)
4. **Set up load balancing** for high availability
5. **Configure SSL/TLS** for secure connections
6. **Implement rate limiting** to prevent abuse

### Scaling

- **Horizontal scaling**: Deploy multiple instances behind a load balancer
- **Worker scaling**: Increase worker count for faster processing
- **Database scaling**: Use read replicas for better performance
- **Storage scaling**: Use CDN for video delivery

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues and questions:
- Create an issue on GitHub
- Check the troubleshooting section
- Review the logs for error details