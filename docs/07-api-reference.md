# API Reference

The REST API server is planned for a future release.

## Status

**Not yet implemented** - Coming soon!

## Planned Features

- RESTful API for all Javinizer operations
- Scrape metadata via HTTP endpoints
- Manage genre replacements
- Query database
- Organize files remotely
- WebSocket support for real-time updates

## Configuration

The API server configuration exists in `config.yaml`:

```yaml
server:
  host: localhost
  port: 8080
```

## Planned Endpoints

### Scraping

```
GET  /api/v1/scrape/:id          # Scrape metadata
GET  /api/v1/scrapers             # List scrapers
POST /api/v1/scrape              # Batch scrape
```

### Movies

```
GET    /api/v1/movies            # List movies
GET    /api/v1/movies/:id        # Get movie
DELETE /api/v1/movies/:id        # Delete movie
POST   /api/v1/movies/:id/nfo    # Generate NFO
```

### Genres

```
GET    /api/v1/genres                        # List genres
GET    /api/v1/genres/replacements           # List replacements
POST   /api/v1/genres/replacements           # Add replacement
DELETE /api/v1/genres/replacements/:original # Remove replacement
```

### File Operations

```
POST /api/v1/organize            # Organize files
POST /api/v1/organize/plan       # Preview organization
```

## Stay Updated

Watch the GitHub repository for API implementation updates:
https://github.com/javinizer/javinizer-go

---

**Next**: [Migration Guide](./08-migration-guide.md)
