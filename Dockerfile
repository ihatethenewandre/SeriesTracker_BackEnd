# Utiliza la imagen de Go basada en Alpine
FROM golang:1.20-alpine

WORKDIR /app

# Copia los archivos de módulos y descarga las dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copia el resto del código fuente
COPY . .

# Compila la aplicación
RUN go build -o main .

# Expone el puerto 8080
EXPOSE 8080

# Ejecuta la aplicación
CMD ["./main"]