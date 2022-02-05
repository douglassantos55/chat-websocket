package main

func main() {
    server := NewServer()
    server.Listen("0.0.0.0:8080")
}
