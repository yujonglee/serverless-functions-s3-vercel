package handler

import (
	"compress/flate"
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	guuid "github.com/google/uuid"
	"github.com/mholt/archiver"
)

type S3GetObjectAclAPI interface {
	GetObjectAcl(ctx context.Context,
		params *s3.GetObjectAclInput,
		optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error)
}

type S3PutObjectAPI interface {
	PutObject(ctx context.Context,
		params *s3.PutObjectInput,
		optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

func GetZipped(w http.ResponseWriter, r *http.Request) {
	BUCKET_NAME := os.Getenv("BUCKET_NAME")

	var (
		cfg      = getConfig()
		client   = s3.NewFromConfig(cfg)
		psClient = s3.NewPresignClient(client)

		downloader = manager.NewDownloader(client, func(d *manager.Downloader) {
			d.PartSize = 64 * 1024 * 1024 // 64MB per part
			// More options available
		})
		uploader = manager.NewUploader(client, func(u *manager.Uploader) {
			u.PartSize = 64 * 1024 * 1024 // 64MB per part
			// More options available
		})
	)

	var (
		infos = getFileInfos(r)
		files = make([]string, 0)
	)

	for _, info := range infos {
		key := info["userId"] + "." + info["fileName"]
		filePath := "/tmp/" + info["fileName"]

		getInput := &s3.GetObjectInput{
			Bucket: &BUCKET_NAME,
			Key:    &key,
		}

		file, err := os.Create(filePath)
		defer file.Close()

		if err != nil {
			fmt.Println("Error creating file.")
			panic(err)
		}

		fmt.Printf("Downloading s3://%s/%s to %s...\n", BUCKET_NAME, key, filePath)
		_, err = downloader.Download(context.TODO(), file, getInput)

		if err != nil {
			fmt.Println("Error downloading file.")
			panic(err)
		}

		files = append(files, filePath)
	}

	zip := archiver.Zip{
		CompressionLevel:       flate.BestCompression,
		MkdirAll:               true,
		SelectiveCompression:   true,
		ContinueOnError:        false,
		OverwriteExisting:      true,
		ImplicitTopLevelFolder: false,
	}

	zippedFileName := guuid.New().String() + ".zip"
	zippedFilePath := "/tmp/" + guuid.New().String() + ".zip"

	err := zip.Archive(files, zippedFilePath)

	if err != nil {
		fmt.Println("Error zipping files.")
		panic(err)
	}

	zippedFile, err := os.Open(zippedFilePath)
	defer zippedFile.Close()

	if err != nil {
		fmt.Println("Error zipping files.")
		panic(err)
	}

	putInput := &s3.PutObjectInput{
		Bucket: &BUCKET_NAME,
		Key:    &zippedFileName,
		Body:   zippedFile,
	}

	result, err := uploader.Upload(context.TODO(), putInput)
	if err != nil {
		fmt.Println("Error uploading zipped files.")
		panic(err)
	}

	fmt.Println(result)

	presignedInput := &s3.GetObjectInput{
		Bucket: &BUCKET_NAME,
		Key:    &zippedFileName,
	}

	resp, err := getPresignedURL(context.TODO(), psClient, presignedInput)
	if err != nil {
		fmt.Println("Error retrieving presigned url.")
		panic(err)
	}

	fmt.Fprint(w, resp.URL)

	for _, f := range append(files, zippedFilePath) {
		fmt.Println(f)
		err := os.Remove(f)

		if err != nil {
			fmt.Printf("Error deleting %s.\n", f)
		}
	}
}
