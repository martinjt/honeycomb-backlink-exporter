# Back link exporter for Honeycomb

This exporter will generate a SpanLink with the inverse SpanId/TraceId properties in Honeycomb. This allows navigation from each side of the SpanLink and therefore opens up some interesting usecases for modelling Span data for complex systems.

## Usecases

The main usecase this was developed for is in Messaging/Event streaming architectures whereby a producer creates multiple messages in a single "batch", therefore being a trace parent isn't viable, however seeing the link is useful.


## Methodology

In OpenTelemetry, Links are created only when you generate a span, and therefore there is no generic mechanism to add reverse links. In Honeycomb however, we have an API for receiving Events that can also receive SpanLinks.

The Exporter is added to the pipeline, and whenever it sees a SpanLink on a span, it will generate the inverse link and send that to the honeycomb API.

## Configuration


```yaml
processors:
  honeycombbacklinkexporter:
    apikey: "<key>"
```

## CAUTION

This has had limited testing. Although it will likely work fine, and not pose a risk to the rest of your telemetry pipeline, you have been warned.