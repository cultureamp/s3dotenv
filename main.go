package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

func main() {
	log.SetOutput(os.Stderr)
	envExec := os.Environ()

	if u := os.Getenv("S3DOTENV"); u != "" {
		var err error
		log.Println("loading env from", u)
		envExec, err = appendFromS3(envExec, u)
		if err != nil {
			log.Fatal(err)
		}
	}

	program, args, err := programAndArgs(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	execErr := syscall.Exec(program, args, envExec)
	if execErr != nil {
		log.Fatal(errors.Wrap(execErr, "exec "+program))
	}
}

func programAndArgs(argv []string) (string, []string, error) {
	argc := len(argv)
	if argc == 0 {
		panic("missing os.Args[0]")
	} else if argc == 1 {
		return "", nil, errors.New(argv[0] + " expected program as first argument")
	}
	program, err := exec.LookPath(argv[1])
	if err != nil {
		return "", nil, errors.Wrap(err, "searching PATH")
	}
	var args []string
	if argc >= 2 {
		args = argv[1:argc]
	}
	return program, args, nil
}

func appendFromS3(env []string, s3url string) ([]string, error) {
	// validate S3DOTENV URL
	u, err := url.Parse(s3url)
	if err != nil || u.Scheme != "s3" {
		return nil, errors.New("S3DOTENV expects s3://... env file URL")
	}

	// read env file using https://github.com/joho/godotenv
	s3env, err := readEnvFromS3(u)
	if err != nil {
		return nil, err
	}

	// add vars without overwriting existing ones
	for k, v := range s3env {
		if _, present := os.LookupEnv(k); present == false {
			env = append(env, k+"="+v)
		}
	}

	return env, nil
}

func readEnvFromS3(u *url.URL) (map[string]string, error) {
	tmpfile, err := ioutil.TempFile("", "s3dotenv")
	if err != nil {
		return nil, errors.Wrap(err, "create temp file")
	}
	defer os.Remove(tmpfile.Name())
	err = downloadFromS3(tmpfile, u)
	if err != nil {
		return nil, errors.Wrap(err, "download from S3")
	}
	return godotenv.Read(tmpfile.Name())
}

func downloadFromS3(file *os.File, u *url.URL) error {
	bucket := u.Host
	key := u.Path[1:len(u.Path)]
	region := u.Query().Get("region")
	sess := session.Must(session.NewSession(&aws.Config{Region: &region}))
	svc := s3.New(sess)
	s3response, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return errors.Wrap(err, "S3 GetObject")
	}
	io.Copy(file, s3response.Body)
	err = file.Close()
	if err != nil {
		return errors.Wrap(err, "close file")
	}
	err = s3response.Body.Close()
	if err != nil {
		return errors.Wrap(err, "close S3 response")
	}
	return nil
}
