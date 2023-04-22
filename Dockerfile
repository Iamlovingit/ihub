FROM golang:1.18 AS build
WORKDIR /
COPY hfproxy /project/hfproxy/
RUN cd /project/hfproxy && go build -mod vender -o hfproxy -ldflags '-linkmode "external" -extldflags "-static"'

FROM scratch AS final
COPY --from=build /project/hfproxy/hfproxy .
COPY --from=build /project/hfproxy/hfproxy-config.yaml .
CMD ["./hfproxy"]