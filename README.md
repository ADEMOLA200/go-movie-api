Sure, here's a basic README.md file based on the provided code:

---

# Movies API

The Movies API is a Go application that interacts with the Star Wars API (SWAPI) to fetch movie data and store it in a Redis cache. It also provides endpoints to add comments to movies and fetch movie details along with associated comments.

## Features

- **Ping**: Check if the server is alive and listening.
- **AddComment**: Add comments to movies.
- **FetchMovies**: Fetch a list of movies along with associated comments.
- **FetchMovie**: Fetch details of a single movie along with associated comments.
- **FetchMovieCharacters**: Fetch characters for a specific movie.

## Prerequisites

- Go installed on your local machine
- Redis server running locally or accessible via network
- PostgreSQL database set up (if using the comment feature)

## Setup

1. Clone the repository:

   ```bash
   https://github.com/ADEMOLA200/go-movie-api.git
   ```

2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. Set up environment variables:

   - Ensure that the following environment variables are set:
     - `REDIS_HOST`: Hostname or IP address of the Redis server
     - `REDIS_PORT`: Port on which Redis is running
     - `POSTGRES_DSN`: Data Source Name (DSN) for connecting to the PostgreSQL database (if using the comment feature)

4. Run the application:

   ```bash
   go run main.go
   ```

## API Endpoints

- **Ping**:
  - Endpoint: `/ping`
  - Method: `GET`
  - Description: Check if the server is alive and listening.

- **AddComment**:
  - Endpoint: `/movies/{movie_id}/comments`
  - Method: `POST`
  - Description: Add a comment to a specific movie.

- **FetchMovies**:
  - Endpoint: `/movies`
  - Method: `GET`
  - Description: Fetch a list of movies along with associated comments.

- **FetchMovie**:
  - Endpoint: `/movies/{movie_id}`
  - Method: `GET`
  - Description: Fetch details of a single movie along with associated comments.

- **FetchMovieCharacters**:
  - Endpoint: `/movies/{movie_id}/characters`
  - Method: `GET`
  - Description: Fetch characters for a specific movie.

## Contributing

Contributions are welcome! Please feel free to fork the repository and submit pull requests to suggest improvements or new features.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
