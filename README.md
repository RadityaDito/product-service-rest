# Product Service

A scalable Golang CRUD application for managing products using PostgreSQL, SQLx, and Echo framework.

## Features

- CRUD operations for products
- Bulk product generation
- Dockerized deployment
- PostgreSQL database
- Pagination support
- Random product generation

## Prerequisites

- Go 1.22+
- Docker
- Docker Compose

## Local Development Setup

1. Clone the repository

```bash
git clone https://your-repo/product-service.git
cd product-service
```

2. Install dependencies

```bash
go mod download
```

3. Run local development

```bash
make run
```

## Docker Deployment

1. Build and start services

```bash
make docker-up
```

2. Stop services

```bash
make docker-down
```

## API Endpoints

### Product Operations

- `POST /products` - Create a new product
- `GET /products` - List products (with pagination)
- `GET /products/:id` - Get a specific product
- `PUT /products/:id` - Update a product
- `DELETE /products/:id` - Delete a product

### Bulk Operations

- `POST /products/bulk/generate?count=1000` - Generate random products
- `DELETE /products/bulk` - Delete all products

### Health Check

- `GET /health` - Check application health

## Environment Variables

| Variable      | Default       | Description         |
| ------------- | ------------- | ------------------- |
| `DB_HOST`     | `localhost`   | Database host       |
| `DB_PORT`     | `5432`        | Database port       |
| `DB_USER`     | `productuser` | Database username   |
| `DB_PASSWORD` | `productpass` | Database password   |
| `DB_NAME`     | `productdb`   | Database name       |
| `DB_SSLMODE`  | `disable`     | PostgreSQL SSL mode |
| `PORT`        | `8080`        | Application port    |

## Testing

Run tests:

```bash
make test
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

MIT License
