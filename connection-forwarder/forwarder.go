package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

func init() {
	log.SetFlags(log.LUTC | log.Llongfile)
}

func main() {
	lr, err := net.Listen("tcp", ":3389")
	if err != nil {
		log.Fatalf("%v", err)
	}

	log.Println("listener is running at 3389")
	for {
		srcConn, err := lr.Accept()
		if err != nil {
			log.Printf("error occurred %v", err)
		}
		targetConn, err := net.Dial("tcp", "xrdp:3389")
		if err != nil {
			log.Printf("error occurred %v", err)
		}

		sr := &SessionRecorder{
			once:      sync.Once{},
			sessionID: "",
			path:      os.Getenv("STORAGE_LOCATION"),
		}
		go func(srcConn, targetConn net.Conn, sr *SessionRecorder) {
			handleConnection(srcConn, targetConn, sr)
		}(srcConn, targetConn, sr)
	}
}

func handleConnection(srcConn, targetConn net.Conn, sr *SessionRecorder) (err error) {
	var wg sync.WaitGroup
	defer func() {
		_ = srcConn.Close()
		_ = targetConn.Close()
		_ = sr.Close()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		n, err := io.Copy(io.MultiWriter(sr, targetConn), srcConn)
		if err != nil {
			log.Printf("forward: connection error %v", err)
			return
		}
		log.Printf("forward: wrote %d bytes", n)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		n, err := io.Copy(io.MultiWriter(sr, srcConn), targetConn)
		if err != nil {
			log.Printf("reverse: connection error %v", err)
			return
		}
		log.Printf("reverse: wrote %d bytes", n)
	}()
	wg.Wait()

	return
}

type SessionRecorder struct {
	once      sync.Once
	sessionID string
	path      string
	storage   *os.File
}

func (sr *SessionRecorder) Write(b []byte) (n int, err error) {
	sr.once.Do(func() {
		if err := os.MkdirAll(sr.path, 0700); err != nil {
			log.Printf("%v", err)
			return
		}
		sr.storage, err = os.Create(fmt.Sprintf("path/%v", sr.sessionID))
		if err != nil {
			return
		}
	})
	return sr.storage.Write(b)
}

func (sr *SessionRecorder) Read(b []byte) (n int, err error) {
	return sr.storage.Read(b)
}

func (sr *SessionRecorder) Close() (err error) {
	if sr.storage != nil {
		err = sr.storage.Close()
	}
	return
}
