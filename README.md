# eflomal-service

A simple service for running eflomal word alignment in a containerized environment.

## Building the Docker Image

To build the Docker image, run:

```sh
docker build -t eflomal-service .
```

## Running the Service

To start the service, run:

```sh
docker run -p 8000:8000 eflomal-service
```

This will start the service and expose it on port 8000.

## Querying the Alignment Endpoint

You can query the `/align` endpoint using `curl` as follows:

```sh
curl -X POST "http://localhost:8000/align" \
  -F "src=@niv.vref.txt" \
  -F "tgt=@urt.vref.txt"
```

- Replace `niv.vref.txt` and `urt.vref.txt` with your source and target text files.
- The endpoint expects a POST request with two files: `src` (source) and `tgt` (target).

## License

MIT License
