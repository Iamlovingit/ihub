FROM golang:1.19 AS build
WORKDIR /
COPY ihub /project/ihub/
RUN cd /project/ihub && go build -mod vender -o ihub -ldflags '-linkmode "external" -extldflags "-static"'

FROM scratch AS final
COPY --from=build /project/ihub/ihub .
COPY --from=build /project/ihub/ihub-config.yaml .
CMD ["./ihub"]