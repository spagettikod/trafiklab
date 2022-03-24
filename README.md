# trafiklab
List Trafiklab top 10 bus stop by number of stops per line.

## Running
Application requires Docker to run.

1. Get (or create) your an API key at https://www.trafiklab.se
2. Start the application:
    ```
    docker run -p 8080:8080 -e LAB_KEY=<you Trafiklab API key> spagettikod/trafiklab
    ```
3. Point your browser to http://localhost:8080

## Release build
```
docker buildx build --push --platform=linux/amd64,linux/arm64 -t spagettikod/trafiklab:latest .
```

## Development build - Mac OS ARM
```
docker buildx build --load --platform=linux/arm64 -t spagettikod/trafiklab:latest .
```