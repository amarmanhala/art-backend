# Art Backend

Simple Go backend with a Spring Boot-like structure:

- `cmd/server`: application entry point
- `internal/config`: environment and database setup
- `internal/controller`: HTTP handlers
- `internal/service`: business logic
- `internal/repository`: database queries
- `internal/model`: request, response, and database models
- `internal/response`: common API success and error format

## Run With Docker For Development

```bash
docker compose up --build
```

This uses Air for live reload. After the first build, normal Go code changes restart the API automatically.

If the containers are already running, you usually do not need to rebuild after editing `.go` files.

If you already started Postgres before the user/address tables were added, recreate the local database volume so Docker runs the new migrations:

```bash
docker compose down -v
docker compose up --build
```

API base URL:

```text
http://localhost:8080/api
```

## Products API

Public APIs:

| Method | Endpoint | Description |
| --- | --- | --- |
| GET | `/api/products` | Get all products with pagination and filters |
| POST | `/api/products` | Create a new product |
| PUT/PATCH | `/api/products/{id-or-slug}` | Edit a product by ID or slug |
| DELETE | `/api/products/{id-or-slug}` | Delete a product by ID or slug |
| GET | `/api/products/{slug}` | Get a single product by slug |
| GET | `/api/products/featured` | Get featured products for homepage |
| GET | `/api/products/search?q=` | Search products by keyword |
| GET | `/api/products/categories` | Get all product categories |
| GET | `/api/products/styles` | Get all art styles |
| GET | `/api/products/themes` | Get all art themes |

Create product:

```bash
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -d '{"title":"Canvas Art","slug":"canvas-art","description":"Large cotton canvas","price":19.99,"currency":"USD","category":"Canvas","style":"Abstract","theme":"Modern","orientation":"Landscape","size":"24x36 in","image_url":"https://example.com/canvas.jpg","thumbnail_url":"https://example.com/canvas-thumb.jpg","original_url":"https://example.com/canvas-original.jpg","stock_quantity":10,"is_active":true}'
```

List products:

```bash
curl "http://localhost:8080/api/products?page=0&size=10&category=Canvas&style=Abstract"
```

Get product:

```bash
curl http://localhost:8080/api/products/golden-abstract-canvas
```

Edit product by ID or slug:

```bash
curl -X PATCH http://localhost:8080/api/products/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated Canvas Art","price":24.99,"stock_quantity":8}'

curl -X PATCH http://localhost:8080/api/products/canvas-art \
  -H "Content-Type: application/json" \
  -d '{"is_active":false}'
```

Delete product by ID or slug:

```bash
curl -X DELETE http://localhost:8080/api/products/1
curl -X DELETE http://localhost:8080/api/products/golden-abstract-canvas
```

Search products:

```bash
curl "http://localhost:8080/api/products/search?q=ocean&page=0&size=10"
```

Get featured products:

```bash
curl http://localhost:8080/api/products/featured
```

Get categories, styles, and themes:

```bash
curl http://localhost:8080/api/products/categories
curl http://localhost:8080/api/products/styles
curl http://localhost:8080/api/products/themes
```

## Contact API

Submit contact form:

```bash
curl -X POST http://localhost:8080/api/contact \
  -H "Content-Type: application/json" \
  -d '{"name":"Amar Singh","email":"amar@example.com","message":"I have a question about an artwork."}'
```

## User APIs

Register:

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Amar","last_name":"Singh","email":"amar@example.com","password":"secret123"}'
```

Login:

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"amar@example.com","password":"secret123"}'
```

Use the returned token for protected APIs:

```bash
curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Update profile:

```bash
curl -X PUT http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Amar","last_name":"Singh"}'
```

Logout:

```bash
curl -X POST http://localhost:8080/api/auth/logout \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Address APIs

List addresses:

```bash
curl http://localhost:8080/api/addresses \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Create address:

```bash
curl -X POST http://localhost:8080/api/addresses \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"full_name":"Amar Singh","phone":"1234567890","address_line_1":"123 Main St","address_line_2":"Apt 4","city":"Toronto","province":"Ontario","postal_code":"M5V 1A1","country":"Canada","is_default":true}'
```

Update address:

```bash
curl -X PUT http://localhost:8080/api/addresses/1 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"full_name":"Amar Singh","phone":"1234567890","address_line_1":"456 Queen St","address_line_2":"","city":"Toronto","province":"Ontario","postal_code":"M5V 2A2","country":"Canada","is_default":true}'
```

Delete address:

```bash
curl -X DELETE http://localhost:8080/api/addresses/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Cart APIs

All cart APIs require:

```text
Authorization: Bearer YOUR_TOKEN
```

Get cart:

```bash
curl http://localhost:8080/api/cart \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Add item to cart:

```bash
curl -X POST http://localhost:8080/api/cart/items \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"product_id":1,"quantity":2}'
```

Update cart item quantity:

```bash
curl -X PUT http://localhost:8080/api/cart/items/1 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"quantity":3}'
```

Delete cart item:

```bash
curl -X DELETE http://localhost:8080/api/cart/items/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Clear cart:

```bash
curl -X DELETE http://localhost:8080/api/cart \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Response Format

Success:

```json
{
  "success": true,
  "message": "products fetched successfully",
  "data": {
    "content": [],
    "page": 0,
    "size": 10,
    "total_elements": 0,
    "total_pages": 0,
    "first": true,
    "last": true
  }
}
```

Error:

```json
{
  "success": false,
  "message": "validation failed",
  "error": {
    "code": "BAD_REQUEST",
    "details": "title is required"
  }
}
```
