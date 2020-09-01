package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3crypto"
	"github.com/matheushr97/aws-services-samples/settings"
)

func main() {
	err := upload()
	//err := download()
	if err != nil {
		fmt.Println(err)
	}
}

func upload() error {
	arn := settings.Arn
	sess := session.New(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	// This is our key wrap handler, used to generate cipher keys and IVs for
	// our cipher builder. Using an IV allows more “spontaneous” encryption.
	// The IV makes it more difficult for hackers to use dictionary attacks.
	// The key wrap handler behaves as the master key. Without it, you can’t
	// encrypt or decrypt the data.
	keywrap := s3crypto.NewKMSKeyGenerator(kms.New(sess), arn)
	// This is our content cipher builder, used to instantiate new ciphers
	// that enable us to encrypt or decrypt the payload.
	builder := s3crypto.AESGCMContentCipherBuilder(keywrap)
	// Let's create our crypto client!
	client := s3crypto.NewEncryptionClient(sess, builder)

	input := &s3.PutObjectInput{
		Bucket: &settings.BucketName,
		Key:    &settings.ObjectKey,
		Body:   bytes.NewReader([]byte("Hello encryption world")),
	}

	_, err := client.PutObject(input)
	// What to expect as errors? You can expect any sort of S3 errors, http://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html.
	// The s3crypto client can also return some errors:
	//  * MissingCMKIDError - when using AWS KMS, the user must specify their key's ARN
	if err == nil {
		fmt.Println("Object sucessfully uploaded")
	}

	return err
}

func download() error {
	sess := session.New(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	client := s3crypto.NewDecryptionClient(sess)

	input := &s3.GetObjectInput{
		Bucket: &settings.BucketName,
		Key:    &settings.ObjectKey,
	}

	result, err := client.GetObject(input)
	// Aside from the S3 errors, here is a list of decryption client errors:
	//   * InvalidWrapAlgorithmError - returned on an unsupported Wrap algorithm
	//   * InvalidCEKAlgorithmError - returned on an unsupported CEK algorithm
	//   * V1NotSupportedError - the SDK doesn’t support v1 because security is an issue for AES ECB
	// These errors don’t necessarily mean there’s something wrong. They just tell us we couldn't decrypt some data.
	// Users can choose to log this and then continue decrypting the data that they can, or simply return the error.
	if err != nil {
		return err
	}

	// Let's read the whole body from the response
	b, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}
