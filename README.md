# stayathome

This repo was created during the livecoding workshops @[GopherCon Russia](https://www.gophercon-russia.ru/en)
(March/August 2020).

The application demonstrates how to start with observability in Go
(logging, metrics, tracing).

# Workshop materials in Russian

You can find the slides [here](http://bit.ly/observability-aug).
The recordings of the practical parts are available [here](https://www.youtube.com/playlist?list=PLyF2SpuGmalvFzvmGnMZGvxbxIZNnAN0S).

# How to run?

```
 PORT=8080 DIAG_PORT=8081 JAEGER_ENDPOINT=127.0.0.1:5775  go run main.go
```

# Tools

- Jaeger for tracing: the [all-in-one](https://www.jaegertracing.io/docs/1.18/getting-started/) container
- PMM for metrics: [how to run](https://github.com/AlekSi/pmm-workshop)

# Step-by-step

- Clean app without any observability: [clean](https://github.com/rumyantseva/stayathome/tree/clean)
- Adding zap: [logger](https://github.com/rumyantseva/stayathome/tree/logger)
- Adding tracer: [tracer](https://github.com/rumyantseva/stayathome/tree/tracer)
- Adding meter: [meter](https://github.com/rumyantseva/stayathome/tree/meter)
- Adding Jaeger and Prometheus: [tools](https://github.com/rumyantseva/stayathome/tree/tools)
