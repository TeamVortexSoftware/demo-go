# Vortex Go SDK Demo

A demo application showcasing the Vortex Go SDK integration with a Gin web server.

## Features

- 🔐 **Authentication System**: Session-based auth with JWT tokens
- ⚡ **Vortex Integration**: Full Vortex API integration for invitation management
- 🎯 **JWT Generation**: Generate Vortex JWTs for authenticated users
- 📧 **Invitation Management**: Get, accept, revoke, and reinvite functionality
- 👥 **Group Management**: Handle invitations by group type and ID
- 🌐 **Interactive Frontend**: Complete HTML interface to test all features

## Prerequisites

- Go 1.19 or later
- The Vortex Go SDK (automatically linked via workspace)

## Installation

1. Navigate to the demo directory:
   ```bash
   cd apps/demo-go
   ```

2. Download dependencies:
   ```bash
   go mod download
   ```

## Running the Demo

1. Set your Vortex API key (optional - defaults to demo key):
   ```bash
   export VORTEX_API_KEY=your-api-key-here
   ```

2. Set the port (optional - defaults to 3000):
   ```bash
   export PORT=3000
   ```

3. Run the server:
   ```bash
   go run src/server.go src/auth.go
   ```

4. Open your browser and visit: `http://localhost:3000`

## Demo Users

The demo includes two test users:

- **Admin User**: `admin@example.com` / `password123` (admin role)
- **Regular User**: `user@example.com` / `userpass` (user role)

## API Endpoints

### Authentication Routes
- `POST /api/auth/login` - Login with email/password
- `POST /api/auth/logout` - Logout (clears session cookie)
- `GET /api/auth/me` - Get current user info

### Demo Routes
- `GET /api/demo/users` - Get all demo users
- `GET /api/demo/protected` - Protected route (requires auth)

### Vortex API Routes
All Vortex routes require authentication:

- `POST /api/vortex/jwt` - Generate Vortex JWT
- `GET /api/vortex/invitations` - Get invitations by target
- `GET /api/vortex/invitations/:id` - Get specific invitation
- `DELETE /api/vortex/invitations/:id` - Revoke invitation
- `POST /api/vortex/invitations/accept` - Accept invitations
- `GET /api/vortex/invitations/by-group/:type/:id` - Get group invitations
- `DELETE /api/vortex/invitations/by-group/:type/:id` - Delete group invitations
- `POST /api/vortex/invitations/:id/reinvite` - Reinvite user

### Health Check
- `GET /health` - Server health status

## Configuration

The demo supports the following environment variables:

- `VORTEX_API_KEY`: Your Vortex API key (defaults to "demo-api-key")
- `PORT`: Server port (defaults to 3000)
- `VORTEX_API_BASE_URL`: Vortex API base URL (uses SDK default)

## Project Structure

```
apps/demo-go/
├── src/
│   ├── server.go      # Main server with routes
│   └── auth.go        # Authentication system
├── public/
│   └── index.html     # Frontend interface
├── go.mod             # Go module definition
└── README.md          # This file
```

## Development

The demo uses the Gin web framework for HTTP handling and includes:

- Session-based authentication with HTTP-only cookies
- CORS support for development
- Static file serving
- JSON request/response handling
- Error handling and validation

## Testing the Demo

1. **Login**: Use one of the demo users to authenticate
2. **Generate JWT**: Test Vortex JWT generation for the authenticated user
3. **Query Invitations**: Test invitation queries by target (email, username, etc.)
4. **Group Operations**: Test group-based invitation operations
5. **Health Check**: Verify the server status and configuration

The frontend provides an interactive interface to test all functionality without needing external tools.

## Integration Notes

This demo shows how to integrate the Vortex Go SDK with:

- Web frameworks (Gin)
- Authentication systems (JWT-based sessions)
- Frontend applications (static HTML/JS)
- Environment-based configuration

The same patterns can be applied to other Go web frameworks like Echo, Fiber, or the standard library.

## Troubleshooting

1. **Import errors**: Make sure you're running from the project root or the Go module path is correct
2. **Port conflicts**: Change the PORT environment variable if 3000 is in use
3. **API errors**: Check the Vortex API key and network connectivity
4. **Authentication issues**: Clear browser cookies and try logging in again