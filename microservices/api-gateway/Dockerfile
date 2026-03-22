FROM maven:3.9.9-eclipse-temurin-17 AS build

WORKDIR /workspace

COPY . .

RUN mvn -B -ntp -pl microservices/api-gateway -am clean package -DskipTests

FROM eclipse-temurin:17-jre-jammy

WORKDIR /app

RUN useradd --create-home --shell /bin/bash spring

COPY --from=build /workspace/microservices/api-gateway/target/api-gateway-0.0.1-SNAPSHOT.jar app.jar
COPY docker/java-service-entrypoint.sh docker-entrypoint.sh

RUN chmod +x /app/docker-entrypoint.sh

ENV SERVICE_NAME=api-gateway

EXPOSE 8080

USER spring

ENTRYPOINT ["/app/docker-entrypoint.sh"]
