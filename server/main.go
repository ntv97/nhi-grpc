package main

import (
 "fmt"
 //"context"
 uploadpb "github.com/nhi-grpc/filetransfer/proto"
 "google.golang.org/grpc"
 "log"
 "net"
 "bytes"
 "os"
 "path/filepath"
 "io"
)

type FileServiceServer struct {
	uploadpb.UnimplementedFileServiceServer
}

type File struct {
	FilePath   string
	buffer     *bytes.Buffer
	OutputFile *os.File
}

func NewFile() *File {
	return &File{
		buffer: &bytes.Buffer{},
	}
}

func (f *File) SetFile(fileName, path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	f.FilePath = filepath.Join(path, fileName)
	file, err := os.Create(f.FilePath)
	if err != nil {
		return err
	}
	f.OutputFile = file
	return nil
}

func (f *File) Write(chunk []byte) error {
	if f.OutputFile == nil {
		return nil
	}
	_, err := f.OutputFile.Write(chunk)
	return err
}

func (f *File) Close() error {
	return f.OutputFile.Close()
}

func (g *FileServiceServer) Upload(stream uploadpb.FileService_UploadServer) error {
	file := NewFile()
	var fileSize uint32
	fileSize = 0
	defer func() {
		if err := file.OutputFile.Close(); err != nil {
			fmt.Println("Error 0")
			return
		}
	}()
	for {
		req, err := stream.Recv()
		if file.FilePath == "" {
			file.SetFile(req.GetFileName(), "./videos")
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error 1")
			return nil
			
		}
		chunk := req.GetChunk()
		fileSize += uint32(len(chunk))
		if err := file.Write(chunk); err != nil {
			fmt.Println("Error 2")
			return nil
		}
	}

	fileName := filepath.Base(file.FilePath)
	return stream.SendAndClose(&uploadpb.FileUploadResponse{FileName: fileName, Size: fileSize})
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		fmt.Println("failed to listen on port 50051: %v", err)
		return
	}
	fmt.Println("Listening on :50051")
	g := grpc.NewServer()
	uploadpb.RegisterFileServiceServer(g, &FileServiceServer{})
	if err := g.Serve(lis); err != nil {
		fmt.Println("Serve error")
		return
	}

}
