package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

type appConfig struct {
	targetAdd string
	runOnce   bool
	webServer bool
}

var fileHandler http.Handler
var shouldExit bool = false
var config appConfig = appConfig{}

var (
	logger        *log.Logger
	verboseLogger *log.Logger
)

func ws(w http.ResponseWriter, r *http.Request) {
	if shouldExit {
		return
	}

	if config.webServer {
		if header := r.Header["Connection"]; header == nil || !strings.Contains(strings.ToLower(header[0]), "upgrade") {
			verboseLogger.Println("Serving file ", r.URL)
			fileHandler.ServeHTTP(w, r)
			return
		}
	}
	// Upgrade connection
	if config.runOnce {
		shouldExit = true
		defer log.Println("Run once! so good bye")
		defer os.Exit(0)
	}
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error in accepting websokcet", err)
		return
	}
	verboseLogger.Printf("Rec connection from %s", conn.RemoteAddr())
	defer verboseLogger.Printf("Close connection from %s", conn.RemoteAddr())
	defer conn.Close()
	tcpconn, err := net.Dial("tcp", config.targetAdd)
	if err != nil {
		log.Println("error in connecting to ", err)
		return
	}
	defer tcpconn.Close()
	go func() {
		defer verboseLogger.Printf("Close connection from %s", conn.RemoteAddr())
		defer conn.Close()
		defer tcpconn.Close()
		for {
			buffer := make([]byte, 1024)
			n, err := tcpconn.Read(buffer)
			if err != nil || n == 0 {
				return
			}
			conn.WriteMessage(websocket.BinaryMessage, buffer[:n])
		}
	}()
	// Read messages from socket
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if msgType != websocket.BinaryMessage {
			log.Print("Non binary message recieved")
		}
		n, err := tcpconn.Write(msg)
		if err != nil || n == 0 {
			return
		}
	}
}

func main() {
	helpFalg := flag.Bool("h", false, "Print Help")
	verboseFlag := flag.Bool("v", false, "Verbose")
	cert := flag.String("cert", "", "SSL certificate file")
	key := flag.String("key", "", "SSL key file")
	webdir := flag.String("web", "", "Serve files from DIR.")
	runOnceFlag := flag.Bool("run-once", false, "handle a single WebSocket connection and exit")
	flag.Parse()
	if *helpFalg {
		flag.PrintDefaults()
		return
	}
	logger = log.New(os.Stdout, "\n", log.Ldate|log.Ltime)
	if *verboseFlag {
		verboseLogger = log.New(os.Stdout, "\n", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		verboseLogger = log.New(ioutil.Discard, "", log.Ldate)
	}
	config.runOnce = *runOnceFlag

	listenadd := flag.Arg(0)
	config.targetAdd = flag.Arg(1)
	ssllog := " - No SSL/TLS support (no cert file)\n"
	if len(*cert) > 0 {
		ssllog = " - SSL/TLS support\n"
	}
	logger.Printf("WebSocket server settings:\n"+
		" - Listen on %s\n"+
		ssllog+
		" - proxying %s\n", listenadd, config.targetAdd)
	//http.Handle("/z/", http.FileServer(http.Dir("./")))
	if len(*webdir) > 0 {
		config.webServer = true
		fileHandler = http.FileServer(http.Dir(*webdir))
	}
	http.HandleFunc("/", ws)
	if len(*cert) > 0 {
		if err := http.ListenAndServeTLS(listenadd, *cert, *key, nil); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := http.ListenAndServe(listenadd, nil); err != nil {
			log.Fatal(err)
		}
	}
}
