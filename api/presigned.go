package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func getConfig() aws.Config {
	var (
		ACCESS_KEY_ID = os.Getenv("ACCESS_KEY_ID")
		SECRET_KEY    = os.Getenv("SECRET_KEY")
		BUCKET_REGION = os.Getenv("BUCKET_REGION")
	)

	CREDENTIALS := credentials.NewStaticCredentialsProvider(ACCESS_KEY_ID, SECRET_KEY, "")

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(BUCKET_REGION),
		config.WithCredentialsProvider(CREDENTIALS))

	if err != nil {
		fmt.Println(err)
		panic("configuration error, " + err.Error())
	}

	return cfg
}

func getFileInfos(r *http.Request) []map[string]string {
	var infos []map[string]string

	bytesBody, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		fmt.Println(err)
		panic("reading body error, " + err.Error())
	}

	json.Unmarshal(bytesBody, &infos)

	return infos
}

type S3PresignGetObjectAPI interface {
	PresignGetObject(
		ctx context.Context,
		params *s3.GetObjectInput,
		optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

func getPresignedURL(c context.Context, api S3PresignGetObjectAPI, input *s3.GetObjectInput) (*v4.PresignedHTTPRequest, error) {
	return api.PresignGetObject(c, input)
}

func Main(w http.ResponseWriter, r *http.Request) {
	BUCKET_NAME := os.Getenv("BUCKET_NAME")

	var (
		cfg      = getConfig()
		client   = s3.NewFromConfig(cfg)
		psClient = s3.NewPresignClient(client)
	)

	var (
		infos = getFileInfos(r)
		urls  = make([]string, 0)
	)

	for _, info := range infos {
		key := info["userId"] + "." + info["fileName"]

		input := &s3.GetObjectInput{
			Bucket: &BUCKET_NAME,
			Key:    &key,
		}

		resp, err := getPresignedURL(context.TODO(), psClient, input)

		if err != nil {
			fmt.Println("Got an error retrieving pre-signed object:")
			fmt.Println(err)
			continue
		}

		urls = append(urls, resp.URL)
	}

	j, err := json.Marshal(urls)

	if err != nil {
		fmt.Println("Got an error converting slice to json:")
		fmt.Println(err)

		fmt.Fprint(w, "[]")

		return
	}

	fmt.Fprint(w, string(j))
}
