package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

func main() {
	port := flag.Int("port", 8080, "the port to listen on; default is 8080")
	flag.Parse()

	l, err := net.Listen("tcp", ":"+strconv.Itoa(*port))

	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Listening on %d\n", *port)
	defer l.Close()

	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println(err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	var i int
	var rURI, rMethod string
	for scanner.Scan() {
		ln := scanner.Text()
		if i == 0 {
			rURI = strings.Fields(ln)[1]
			rMethod = strings.Fields(ln)[0]
			fmt.Println("URI:", rURI)
			fmt.Println("METHOD:", rMethod)
		}
		fmt.Println(ln)
		if ln == "" {
			fmt.Println("THIS IS THE END OF THE HTTP REQUEST HEADERS")
			break
		}
		i++
	}
	switch {
	case rURI == "/" && rMethod == "GET":
		handleIndex(conn)
	case rURI == "/apply" && rMethod == "GET":
		handleApply(conn)
	case rURI == "/apply" && rMethod == "POST":
		handleApplyPost(conn)
	default:
		handleDefault(conn)
	}
}

func handleIndex(conn net.Conn) {
	body := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>GET INDEX</title>
	</head>
	<body>
		<h1>"GET INDEX"</h1>
		<a href="/">index</a><br>
		<a href="/apply">apply</a><br>
	</body>
	</html>
`
	io.WriteString(conn, "HTTP/1.1 200 OK\r\n")
	fmt.Fprintf(conn, "Content-Length: %d\r\n", len(body))
	fmt.Fprint(conn, "Content-Type: text/html\r\n")
	io.WriteString(conn, "\r\n")
	io.WriteString(conn, body)
}

func handleApply(conn net.Conn) {
	body := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title>GET DOG</title>
		</head>
		<body>
			<h1>"GET APPLY"</h1>
			<a href="/">index</a><br>
			<a href="/apply">apply</a><br>
			<form action="/apply" method="POST">
			<input type="hidden" value="WOW">
			<input type="submit" value="submit">
			</form>
		</body>
		</html>
	`
	io.WriteString(conn, "HTTP/1.1 200 OK\r\n")
	fmt.Fprintf(conn, "Content-Length: %d\r\n", len(body))
	fmt.Fprint(conn, "Content-Type: text/html\r\n")
	io.WriteString(conn, "\r\n")
	io.WriteString(conn, body)
}

func handleApplyPost(conn net.Conn) {
	body := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title>POST APPLY</title>
		</head>
		<body>
			<h1>"POST APPLY"</h1>
			<a href="/">index</a><br>
			<a href="/apply">apply</a><br>
		</body>
	</html>
	`
	io.WriteString(conn, "HTTP/1.1 200 OK\r\n")
	fmt.Fprintf(conn, "Content-Length: %d\r\n", len(body))
	fmt.Fprint(conn, "Content-Type: text/html\r\n")
	io.WriteString(conn, "\r\n")
	io.WriteString(conn, body)
}

func handleDefault(conn net.Conn) {
	body := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title>default</title>
		</head>
		<body>
			<h1>"default"</h1>
		</body>
		</html>
	`
	io.WriteString(conn, "HTTP/1.1 200 OK\r\n")
	fmt.Fprintf(conn, "Content-Length: %d\r\n", len(body))
	fmt.Fprint(conn, "Content-Type: text/html\r\n")
	io.WriteString(conn, "\r\n")
	io.WriteString(conn, body)
}
